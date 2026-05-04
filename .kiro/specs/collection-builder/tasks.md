# Implementation Plan: Collection Builder

## Overview

Replace the `atlas` domain concept with `collection` end-to-end: new DB models and migration, new domain package, repository, service, HTTP handlers, and templates. The `atlases` table is dropped and recreated as `collections`; seed data is recreated. All atlas-related packages, routes, and templates are removed and replaced.

## Tasks

- [x] 1. DB migration — replace atlases table and add collection structure tables
  - [x] 1.1 Replace `AtlasModel` with `CollectionModel` in `internal/persistence/migration/`
    - Delete `atlas_model.go`; create `collection_model.go` with `CollectionModel` mapping to table `collections` (fields: `Name`, `Description`, `Goals`, `Type`, `Authors`, plus `gorm.Model`)
    - Add `CollectionSectionModel` (fields: `CollectionID`, `ParentID *uint`, `Name`, `Position`) with `TableName() = "collection_sections"`
    - Add `CollectionSectionAssignmentModel` (fields: `SectionID`, `TissueRecordID`, `Position`) with `TableName() = "collection_section_assignments"`
    - _Requirements: 8.1, 8.2, 6.1_

  - [x] 1.2 Update `migration.RunMigration()` to migrate new models and drop atlas references
    - Replace `&AtlasModel{}` with `&CollectionModel{}` in `AutoMigrate` call
    - Add `&CollectionSectionModel{}` and `&CollectionSectionAssignmentModel{}` to `AutoMigrate`
    - Update `seedSampleTissueRecords` to create a `CollectionModel` seed record instead of `AtlasModel`
    - _Requirements: 8.2, 6.1_

- [x] 2. Domain layer — `internal/core/collection` package
  - [x] 2.1 Create `internal/core/collection/collection.go`
    - Define `CollectionType` string type and constants: `atlas`, `database`, `reference`, `other`
    - Define `Collection` struct (ID, Name, Description, Goals, Type, Authors, Sections, CreatedAt, UpdatedAt)
    - Define `Section` struct (ID, CollectionID, ParentID `*uint`, Name, Position, Assignments, Subsections)
    - Define `SectionAssignment` struct (ID, SectionID, TissueRecordID, Position)
    - Implement `Collection.Validate()`: reject empty/whitespace-only name, name > 200 chars, invalid type enum
    - Implement `Section.Validate()`: reject empty/whitespace-only name
    - Define domain errors: `ErrEmptyName`, `ErrNameTooLong`, `ErrInvalidType`, `ErrNotFound`, `ErrDuplicateAssignment`, `ErrMaxDepthExceeded`
    - _Requirements: 1.2, 1.3, 1.4, 2.2_

- [x] 3. Repository interface and GORM implementation
  - [x] 3.1 Create `internal/core/collection/repository_interface.go`
    - Define `RepositoryInterface` with CRUD methods for Collection, plus `CreateSection`, `UpdateSection`, `DeleteSection`, `ReorderSections`, `CreateAssignment`, `DeleteAssignment`, `ReorderAssignments`
    - _Requirements: 6.1, 6.2, 6.3_

  - [x] 3.2 Create `internal/persistence/repositories/gorm_collection_repository.go`
    - Implement `GormCollectionRepository` satisfying `collection.RepositoryInterface`
    - `Save`: insert `CollectionModel`, return ID
    - `Retrieve`: preload `Sections` → `Assignments` and `Sections` → `Subsections` → `Assignments`; map to domain types; default empty `Type` to `"atlas"`
    - `Update`: update `CollectionModel` fields
    - `Delete`: delete collection and cascade-delete sections and assignments via explicit deletes
    - `List`: return all collections ordered by `created_at DESC`
    - `CreateSection`: insert `CollectionSectionModel` with `Position = count_of_existing + 1`
    - `UpdateSection`: update name
    - `DeleteSection`: delete all `CollectionSectionAssignmentModel` for section, then delete section
    - `ReorderSections`: batch-update positions from the provided map
    - `CreateAssignment`: insert with `Position = count_of_existing + 1`; return `ErrDuplicateAssignment` if (SectionID, TissueRecordID) already exists
    - `DeleteAssignment`: delete assignment, then resequence remaining positions in that section (1..N-1)
    - `ReorderAssignments`: batch-update positions from the provided map
    - _Requirements: 2.1, 2.3, 2.4, 2.5, 2.7, 3.1, 3.2, 3.3, 3.4, 6.1, 6.2, 6.3_

  - [x] 3.3 Update `internal/persistence/repositories/factory.go`
    - Remove `NewAtlasRepository()` function
    - Add `NewCollectionRepository() collection.RepositoryInterface` returning `NewGormCollectionRepository()`
    - _Requirements: 8.1_

- [x] 4. Service layer — `internal/services/collection_service.go`
  - [x] 4.1 Create `internal/services/collection_service.go`
    - Define `CollectionService` struct with `repo collection.RepositoryInterface` and `trRepo tissuerecord.RepositoryInterface`
    - Implement: `CreateCollection`, `GetCollection`, `UpdateCollection`, `DeleteCollection`, `ListCollections`
    - Implement: `CreateSection`, `RenameSection`, `DeleteSection`, `ReorderSections`
    - Implement: `AssignTissueRecord` (calls `repo.CreateAssignment`; propagates `ErrDuplicateAssignment`)
    - Implement: `RemoveAssignment` (calls `repo.DeleteAssignment` which resequences)
    - Implement: `ReorderAssignments`
    - Implement: `SearchTissueRecords(query string) ([]tissuerecord.TissueRecord, error)` — case-insensitive substring match on name and taxon name
    - Implement: `CreateTissueRecordAndAssign(tr *tissuerecord.TissueRecord, sectionID uint) error` — persist TR then create assignment
    - _Requirements: 1.5, 1.6, 2.1, 2.3, 2.4, 2.5, 2.7, 3.1, 3.2, 3.3, 3.4, 4.2, 5.3, 6.1, 6.2_

- [x] 5. Checkpoint — domain, repository, and service compile cleanly
  - Run `go build ./internal/... ./cmd/api-server-gin/...` and confirm zero errors before proceeding

- [x] 6. HTTP handlers — `cmd/api-server-gin/collections/` package
  - [x] 6.1 Create `cmd/api-server-gin/collections/collections.go`
    - Define `collectionService()` helper returning `*services.CollectionService`
    - Implement `ListCollections`: render `pages/collection_list.html` with all collections
    - Implement `NewCollectionForm` / `NewCollectionFormCancel`: HTMX inline form fragment
    - Implement `CreateCollection`: parse form fields (name, description, goals, type, authors), validate, persist, flash + redirect to `/collections`
    - Implement `EditCollectionForm` / `EditCancelCollection`: HTMX inline edit fragment
    - Implement `UpdateCollection`: update metadata, flash, return updated row fragment
    - Implement `DeleteCollection`: delete with cascade, flash, return empty fragment
    - Implement `ConfirmDeleteCollection` / `ConfirmDeleteCollectionCancel`
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 6.2, 7.1, 7.3_

  - [x] 6.2 Create `cmd/api-server-gin/collections/builder.go`
    - Implement `BuilderPage`: load collection with sections/subsections/assignments; render `pages/collection_builder.html`
    - Implement `CreateSection`: parse name, call service, return sections list fragment
    - Implement `UpdateSection`: rename section, return updated section fragment
    - Implement `DeleteSection`: delete section + assignments, return empty fragment
    - Implement `ReorderSections`: parse positions map from form, call service, return updated sections list fragment
    - Implement `CreateAssignment`: parse `tissue_record_id`, call service; on duplicate return 409 info fragment; on success return updated assignment list fragment
    - Implement `DeleteAssignment`: remove assignment, return updated assignment list fragment
    - Implement `ReorderAssignments`: parse positions map, call service, return updated assignment list fragment
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 3.1, 3.2, 3.3, 3.4, 6.5, 7.2, 7.4_

  - [x] 6.3 Add `SearchTissueRecords` handler to `cmd/api-server-gin/tissue_records/tissue_records.go`
    - `GET /tissue_records/search?q=<term>&section_id=<id>` — call `CollectionService.SearchTissueRecords(q)`, render `includes/tr_search_results.html` fragment
    - _Requirements: 4.1, 4.2, 4.3, 4.4_

  - [x] 6.4 Add inline tissue record creation to `cmd/api-server-gin/tissue_records/tissue_record_crud.go`
    - Handle `section_id` hidden field on `POST /tissue_records`: after persisting the TR, if `section_id` is present call `CollectionService.AssignTissueRecord`; return updated assignment list fragment
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [x] 7. Templates — collections list and fragments
  - [x] 7.1 Create `web/templates/pages/collection_list.html`
    - Table with columns: Name, Type, Created At, Actions (Builder link, Delete)
    - "New Collection" button triggers HTMX inline form swap into `#collection-form-container`
    - Empty state message when no collections exist
    - _Requirements: 7.1, 7.3_

  - [x] 7.2 Create `web/templates/pages/collection_form.html` (fragment `{{define "collection-form"}}`)
    - Fields: name (text), description (textarea), goals (textarea), type (select: atlas/database/reference/other), authors (text)
    - Inline validation error display
    - Cancel button with HTMX swap
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [x] 7.3 Create `web/templates/includes/collection_row.html` (fragment `{{define "collection-row"}}`)
    - Table row with name, type badge, created_at, edit/delete actions using HTMX
    - _Requirements: 7.1_

- [x] 8. Templates — collection builder page
  - [x] 8.1 Create `web/templates/pages/collection_builder.html`
    - Two-panel layout: left metadata panel (name, description, goals, type, authors, Save Metadata button via HTMX PUT), right sections panel
    - Sections panel: "Add Section" button (HTMX POST), renders `#sections-list` container
    - Breadcrumb: Home → Collections → [Collection Name]
    - _Requirements: 1.1, 1.5, 1.6, 7.2, 7.4_

  - [x] 8.2 Create `web/templates/includes/collection_sections_list.html` (fragment `{{define "collection-sections-list"}}`)
    - Renders the full ordered list of top-level sections; each section includes its subsections and assignments
    - Each section row: name, up/down reorder buttons (HTMX POST to reorder endpoint), delete button
    - "Add Subsection" button per section
    - Includes `{{template "collection-assignments-list"}}` per section/subsection
    - _Requirements: 2.1, 2.3, 2.5, 2.6, 2.7_

  - [x] 8.3 Create `web/templates/includes/collection_assignments_list.html` (fragment `{{define "collection-assignments-list"}}`)
    - Ordered list of tissue record assignments within a section
    - Each assignment: TR name, up/down buttons (HTMX POST reorder), remove button (HTMX DELETE)
    - "Search / Add TR" trigger (HTMX GET to `/tissue_records/search?section_id=X`) swapping into `#tr-search-{sectionID}`
    - "Create new TR" button opening the inline creation modal
    - _Requirements: 3.1, 3.3, 3.4, 4.4, 5.1_

  - [x] 8.4 Create `web/templates/includes/tr_search_results.html` (fragment `{{define "tr-search-results"}}`)
    - Search input (HTMX GET on input with debounce) + results list
    - Each result: TR name, taxon, "Add" button (HTMX POST to assignment endpoint)
    - Empty state message when no results
    - _Requirements: 4.2, 4.3, 4.4_

  - [x] 8.5 Create `web/templates/includes/collection_tr_modal.html` (fragment `{{define "collection-tr-modal"}}`)
    - Alpine.js modal (`x-data="{ open: false }"`) with fields: name, notes, taxon (select)
    - Form uses `hx-post="/tissue_records"` with hidden `section_id` field
    - Cancel button closes modal without persisting (`@click="open = false"`)
    - Inline validation error display within modal
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ] 9. Templates — collection public view page
  - [x] 9.1 Create `web/templates/pages/collection_view.html`
    - Section-aware layout: renders each top-level section as a collapsible group header
    - Within each section: renders subsections (if any) as sub-groups, then tissue record cards
    - Tissue record card: name, taxon lineage badges, notes, slide count badge (same card design as existing `atlas_view.html`)
    - Empty state per section when no assignments exist
    - _Requirements: 7.1, 7.2_

- [x] 10. Update main-menu and remove all atlas artifacts
  - [x] 10.1 Update `web/templates/includes/main-menu.html`
    - Replace all `/atlases` hrefs with `/collections`
    - Replace "Atlases" label with "Collections"
    - _Requirements: 7.1_

  - [x] 10.2 Delete atlas templates
    - Delete `web/templates/pages/atlas_list.html`, `atlas_form.html`, `atlas_view.html`
    - Delete `web/templates/includes/atlas_row.html`, `atlas_tbody.html`
    - _Requirements: (cleanup)_

  - [x] 10.3 Delete atlas handler, domain, and repository files
    - Delete `cmd/api-server-gin/atlas/atlas.go`, `atlas_crud.go`, `atlas_view.go`
    - Delete `internal/core/atlas/atlas.go`, `internal/core/atlas/repository_interface.go`
    - Delete `internal/persistence/repositories/gorm_atlas_repository.go`, `postgres_atlas_repository.go`
    - Delete `internal/services/atlas_service.go`
    - _Requirements: (cleanup)_

  - [x] 10.4 Update tissue record workspace to use collection references
    - Rename `web/templates/includes/workspace_atlas_section.html` → `workspace_collection_section.html`
    - Update `cmd/api-server-gin/tissue_records/workspace.go`: replace atlas service calls with collection service equivalents
    - Update routes in `main.go` accordingly
    - _Requirements: (cleanup / consistency)_

- [x] 11. Route registration in `cmd/api-server-gin/main.go`
  - Remove import of `atlas` handler package; add import of `collections` handler package
  - Remove all `/atlases` and `/atlas/:id` route registrations
  - Add all `/collections` routes: list, new, create, builder, edit, update, delete, sections CRUD, assignments CRUD
  - Add `GET /tissue_records/search` route
  - Update `logStartupInfo()` to reflect new routes
  - _Requirements: 7.1, 7.2_

- [x] 12. Final checkpoint — application compiles and runs end-to-end
  - Run `go build ./...` and confirm zero errors
  - Start the app and verify: collections list loads, builder page loads, section creation works, tissue record assignment works, public view renders sections

- [ ] 13. Property-based tests — all 15 correctness properties
  - [x] 13.1 Write property test: Property 1 — whitespace and empty names rejected
    - Create `internal/core/collection/tests/collection_test.go`
    - Use `pgregory.net/rapid` to generate whitespace-only strings; assert `Collection.Validate()` and `Section.Validate()` return `ErrEmptyName`
    - `// Feature: collection-builder, Property 1: Whitespace and empty names are rejected`
    - _Requirements: 1.2, 2.2, 5.2_

  - [x] 13.2 Write property test: Property 2 — name length boundary enforcement
    - Generate strings of length 201–500; assert `Collection.Validate()` returns `ErrNameTooLong`
    - `// Feature: collection-builder, Property 2: Name length boundary enforcement`
    - _Requirements: 1.3_

  - [x] 13.3 Write property test: Property 3 — collection type enum enforcement
    - Generate arbitrary strings not in `{atlas, database, reference, other}`; assert `Collection.Validate()` returns `ErrInvalidType`
    - `// Feature: collection-builder, Property 3: Collection type enum enforcement`
    - _Requirements: 1.4_

  - [x] 13.4 Write property test: Property 4 — collection metadata round-trip
    - Use SQLite test DB; generate valid Collection structs; save and retrieve; assert field equality
    - `// Feature: collection-builder, Property 4: Collection metadata round-trip`
    - _Requirements: 1.5, 1.6, 6.1_

  - [x] 13.5 Write property test: Property 5 — section creation assigns next position
    - Generate a Collection with 0–10 existing sections; call `CreateSection`; assert `position = N+1` and correct `CollectionID`
    - `// Feature: collection-builder, Property 5: Section creation assigns next position`
    - _Requirements: 2.1, 2.5_

  - [x] 13.6 Write property test: Property 6 — reorder persists new positions
    - Generate N sections/assignments; apply a random permutation via `ReorderSections`/`ReorderAssignments`; retrieve; assert persisted order matches permutation
    - `// Feature: collection-builder, Property 6: Reorder persists new positions`
    - _Requirements: 2.3, 2.7, 3.3, 6.4_

  - [x] 13.7 Write property test: Property 7 — section deletion removes all assignments
    - Generate a Section with 0–20 assignments; call `DeleteSection`; query `collection_section_assignments` for that `section_id`; assert count = 0
    - `// Feature: collection-builder, Property 7: Section deletion removes all assignments`
    - _Requirements: 2.4_

  - [x] 13.8 Write property test: Property 8 — assignment creation appends at end
    - Generate a Section with 0–10 assignments; call `AssignTissueRecord`; assert new assignment `position = N+1`
    - `// Feature: collection-builder, Property 8: Assignment creation appends at end`
    - _Requirements: 3.1_

  - [x] 13.9 Write property test: Property 9 — duplicate assignment rejection
    - Generate a Section with at least one assignment; attempt to assign the same tissue record again; assert error and count unchanged
    - `// Feature: collection-builder, Property 9: Duplicate assignment rejection`
    - _Requirements: 3.2_

  - [x] 13.10 Write property test: Property 10 — assignment removal resequences positions
    - Generate a Section with 2–10 assignments; remove one at a random index; retrieve remaining; assert positions are 1..N-1 with no gaps
    - `// Feature: collection-builder, Property 10: Assignment removal resequences positions`
    - _Requirements: 3.4_

  - [x] 13.11 Write property test: Property 11 — search returns only matching records
    - Generate a pool of tissue records with random names/taxa; generate a random query string; call `SearchTissueRecords`; assert every result contains the query as a case-insensitive substring
    - `// Feature: collection-builder, Property 11: Search returns only matching records`
    - _Requirements: 4.2_

  - [x] 13.12 Write property test: Property 12 — inline creation persists record and creates assignment
    - Generate valid tissue record data and a target section; call `CreateTissueRecordAndAssign`; assert TR exists in pool and assignment exists in section
    - `// Feature: collection-builder, Property 12: Inline creation persists record and creates assignment`
    - _Requirements: 5.3_

  - [x] 13.13 Write property test: Property 13 — collection deletion cascades
    - Generate a Collection with a random tree of sections and assignments; call `DeleteCollection`; assert zero sections and assignments remain for that collection ID
    - `// Feature: collection-builder, Property 13: Collection deletion cascades`
    - _Requirements: 6.2_

  - [x] 13.14 Write property test: Property 14 — tissue record deletion removes all assignments
    - Generate multiple collections with assignments to a shared tissue record; delete the tissue record; assert zero assignments remain with that `tissue_record_id`
    - `// Feature: collection-builder, Property 14: Tissue record deletion removes all assignments`
    - _Requirements: 6.3_

  - [x] 13.15 Write property test: Property 15 — collection list rendering includes required fields
    - Generate a random set of Collections; render the list; assert each collection's name, type, and creation date appear in the output
    - `// Feature: collection-builder, Property 15: Collection list rendering includes required fields`
    - _Requirements: 7.1_

## Notes

- Property-based tests use `pgregory.net/rapid` and live in `internal/core/collection/tests/`
- Each property test is annotated with its property number and requirements clause
- The `atlases` table is dropped and recreated as `collections`; seed data is recreated from scratch
