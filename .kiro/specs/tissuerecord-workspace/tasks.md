# Implementation Plan: TissueRecord Workspace

## Overview

Implement the TissueRecord Workspace feature in Go/Gin. The work proceeds in eight steps: extend the repository interface and GORM implementation with association methods, add delegation methods to the service layer, add Alpine.js to the base layout, create the four HTML templates, implement the ten workspace handler functions, register the new routes in `main.go`, update the list row template, and verify everything compiles and tests pass.

## Tasks

- [x] 1. Extend repository interface and GORM implementation with association methods
  - [x] 1.1 Add six association method signatures to `internal/core/tissuerecord/repository_interface.go`
    - Add `AddAtlas(trID, atlasID uint) error`, `RemoveAtlas(trID, atlasID uint) error`, `ListAtlases(trID uint) ([]atlas.Atlas, error)`
    - Add `AddCategory(trID, catID uint) error`, `RemoveCategory(trID, catID uint) error`, `ListCategories(trID uint) ([]category.Category, error)`
    - _Requirements: 6.1, 6.2, 7.1, 7.2_

  - [x] 1.2 Implement the six methods on `GormTissueRecordRepository` in `internal/persistence/repositories/gorm_tissuerecord_repository.go`
    - Use GORM `Association` API on `TissueRecordModel` (join tables `atlas_tissue_records` and `tissue_record_categories` already exist)
    - `AddAtlas`/`AddCategory`: use `db.Model(&model).Association("Atlases").Append(...)` — idempotent via GORM's upsert behaviour
    - `RemoveAtlas`/`RemoveCategory`: use `db.Model(&model).Association("Atlases").Delete(...)`
    - `ListAtlases`/`ListCategories`: use `db.Model(&model).Association("Atlases").Find(&result)`
    - Map results to core domain types (`atlas.Atlas`, `category.Category`)
    - _Requirements: 6.1, 6.2, 6.3, 7.1, 7.2, 7.3_

  - [ ]* 1.3 Write property tests for repository association methods
    - **Property 8: Category add is idempotent** — call `AddCategory` 1–5 times, assert `ListCategories` count for that ID == 1
    - **Validates: Requirements 6.3**
    - **Property 9: Atlas add is idempotent** — same pattern for `AddAtlas` / `ListAtlases`
    - **Validates: Requirements 7.3**
    - **Property 10: Category add→list round-trip** — `AddCategory` then assert catID present in `ListCategories`
    - **Validates: Requirements 6.1, 6.4**
    - **Property 11: Category remove→list round-trip** — `RemoveCategory` then assert catID absent from `ListCategories`
    - **Validates: Requirements 6.2**
    - **Property 12: Atlas add→list round-trip** — `AddAtlas` then assert atlasID present in `ListAtlases`
    - **Validates: Requirements 7.1, 7.4**
    - **Property 13: Atlas remove→list round-trip** — `RemoveAtlas` then assert atlasID absent from `ListAtlases`
    - **Validates: Requirements 7.2**
    - Use [gopter](https://github.com/leanovate/gopter), minimum 100 iterations per property
    - Tag each test: `// Feature: tissuerecord-workspace, Property N: <text>`
    - Place in `internal/core/tissuerecord/tests/`

- [x] 2. Add association delegation methods to `TissueRecordService`
  - Add six thin delegation methods to `internal/services/tissuerecord_service.go`:
    - `AddAtlas`, `RemoveAtlas`, `ListAtlases`, `AddCategory`, `RemoveCategory`, `ListCategories`
  - Each method calls the corresponding method on `s.repo`
  - _Requirements: 6.1, 6.2, 7.1, 7.2_

- [x] 3. Add Alpine.js CDN to base layout
  - Add `<script src="https://cdn.jsdelivr.net/npm/alpinejs@3" defer></script>` to `web/templates/layouts/base.html` (after the existing HTMX script tag)
  - _Requirements: 2.2, 2.3, 4.2, 5.2_

- [x] 4. Create workspace HTML templates
  - [x] 4.1 Create `web/templates/pages/tissue_record_workspace.html`
    - Extends `base` layout; defines `content` block
    - Two-column grid (`lg:grid-cols-3`): left column (1/3) stacks basic-info, atlas, category sections; right column (2/3) holds slide gallery
    - Each section wrapped in a named `div` target: `#basic-info-section`, `#atlas-section`, `#category-section`, `#slide-gallery`
    - Includes `{{template "workspace-basic-info" .}}`, `{{template "workspace-atlas-section" .}}`, `{{template "workspace-category-section" .}}`, `{{template "slide-gallery" .}}`
    - _Requirements: 1.2, 1.3, 1.4_

  - [x] 4.2 Create `web/templates/includes/workspace_basic_info.html`
    - Defines template `workspace-basic-info`
    - Alpine `x-data="{ editing: false }"` on the card wrapper
    - Display view (`x-show="!editing"`): shows name, notes, taxon; "Edit" button sets `editing = true`
    - Edit form (`x-show="editing"`): `hx-put` to `/tissue_records/{{.TissueRecord.ID}}/workspace/basic-info`, `hx-target="#basic-info-section"`, `hx-swap="innerHTML"`; fields for name, notes, taxon select
    - Cancel button: `@click="editing = false"` plus `hx-get` to basic-info fragment URL to restore server state
    - Renders `Errors["name"]` when present
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6_

  - [x] 4.3 Create `web/templates/includes/workspace_atlas_section.html`
    - Defines template `workspace-atlas-section`
    - Alpine `x-data="{ showDropdown: false }"` on the card wrapper
    - Lists `.Atlases` with a "Remove" button per entry: `hx-delete` to `/tissue_records/{{$.TissueRecord.ID}}/atlases/{{.ID}}`, `hx-target="#atlas-section"`, `hx-swap="innerHTML"`
    - "Add Atlas" dropdown (`x-show="showDropdown"`): iterates `.AvailAtlases`; each item `hx-post` to add URL, `@click="showDropdown = false"`
    - "Add Atlas" toggle button hidden when `.AvailAtlases` is empty (`{{if .AvailAtlases}}`)
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_

  - [x] 4.4 Create `web/templates/includes/workspace_category_section.html`
    - Defines template `workspace-category-section`
    - Alpine `x-data="{ showDropdown: false }"` on the card wrapper
    - Displays `.Categories` as badge tags; each badge has a remove button: `hx-delete` to `/tissue_records/{{$.TissueRecord.ID}}/categories/{{.ID}}`, `hx-target="#category-section"`, `hx-swap="innerHTML"`
    - "Add Category" dropdown (`x-show="showDropdown"`): iterates `.AvailCats`; each item `hx-post` to add URL, `@click="showDropdown = false"`
    - "Add Category" toggle button hidden when `.AvailCats` is empty (`{{if .AvailCats}}`)
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6_

- [x] 5. Implement workspace handlers in `cmd/api-server-gin/tissue_records/workspace.go`
  - [x] 5.1 Implement `WorkspaceHandler` (full page)
    - Parse and validate `:id`; 404 if not found
    - Load TissueRecord, all taxa, associated atlases, all atlases, associated categories, all categories
    - Compute `AvailAtlases` (all minus associated) and `AvailCats` (all minus associated)
    - Build `WorkspaceViewModel` with `Crumbs`: `[{Home, /}, {Tissue Records, /tissue_records}, {record.Name}]`
    - Render `tissue_record_workspace.html` via `shared.RenderPage`
    - _Requirements: 1.2, 1.3, 1.4, 1.5_

  - [x] 5.2 Implement `RedirectToWorkspace`
    - Parse `:id`; return HTTP 301 with `Location: /tissue_records/:id/workspace`
    - _Requirements: 8.3_

  - [x] 5.3 Implement `BasicInfoFragment` (GET)
    - Parse `:id`; 404 if not found
    - Load TissueRecord and all taxa
    - Render `workspace-basic-info` fragment (display mode, `Errors` empty)
    - _Requirements: 2.1, 2.6_

  - [x] 5.4 Implement `SaveBasicInfo` (PUT)
    - Parse `:id`; validate name non-empty (HTTP 422 + fragment with `Errors["name"]` on failure)
    - Persist updated name, notes, taxon via `trService().Update`
    - Return refreshed `workspace-basic-info` fragment (display mode)
    - _Requirements: 2.4, 2.5, 2.7_

  - [x] 5.5 Implement `AddAtlasToTissueRecord` (POST)
    - Parse `:id` and `:atlasID`; 400 on invalid params
    - Call `trService().AddAtlas(trID, atlasID)`
    - Reload associated and available atlases; render `workspace-atlas-section` fragment
    - _Requirements: 4.3, 7.1, 7.3_

  - [x] 5.6 Implement `RemoveAtlasFromTissueRecord` (DELETE)
    - Parse `:id` and `:atlasID`; 400 on invalid params
    - Call `trService().RemoveAtlas(trID, atlasID)`
    - Reload associated and available atlases; render `workspace-atlas-section` fragment
    - _Requirements: 4.4, 7.2_

  - [x] 5.7 Implement `AtlasSectionFragment` (GET)
    - Parse `:id`; 404 if not found
    - Reload associated and available atlases; render `workspace-atlas-section` fragment
    - _Requirements: 4.1, 4.2_

  - [x] 5.8 Implement `AddCategoryToTissueRecord` (POST)
    - Parse `:id` and `:categoryID`; 400 on invalid params
    - Call `trService().AddCategory(trID, catID)`
    - Reload associated and available categories; render `workspace-category-section` fragment
    - _Requirements: 5.3, 6.1, 6.3_

  - [x] 5.9 Implement `RemoveCategoryFromTissueRecord` (DELETE)
    - Parse `:id` and `:categoryID`; 400 on invalid params
    - Call `trService().RemoveCategory(trID, catID)`
    - Reload associated and available categories; render `workspace-category-section` fragment
    - _Requirements: 5.4, 6.2_

  - [x] 5.10 Implement `CategorySectionFragment` (GET)
    - Parse `:id`; 404 if not found
    - Reload associated and available categories; render `workspace-category-section` fragment
    - _Requirements: 5.1, 5.2_

  - [ ]* 5.11 Write property tests for handler-level properties
    - **Property 1: Workspace page title equals record name** — generate random names, call `WorkspaceHandler`, assert template data `Title == name`
    - **Validates: Requirements 1.2**
    - **Property 2: Breadcrumb structure is always correct** — generate random names, assert breadcrumb slice has exactly 3 items with correct labels and URLs
    - **Validates: Requirements 1.3**
    - **Property 3: Basic info display fragment contains current values** — generate random (name, notes, taxon), call `BasicInfoFragment`, assert fragment contains those values
    - **Validates: Requirements 2.1, 2.6**
    - **Property 4: Valid basic info save round-trip** — generate valid (name, notes), PUT to `SaveBasicInfo`, assert returned fragment contains new values
    - **Validates: Requirements 2.4**
    - **Property 6: Available atlases list is complement of associated atlases** — generate random atlas sets and association subsets, assert `AvailAtlases == all - associated`
    - **Validates: Requirements 4.2, 4.5**
    - **Property 7: Available categories list is complement of associated categories** — same pattern for categories
    - **Validates: Requirements 5.2, 5.5**
    - **Property 14: Old detail URL redirects to workspace for any valid ID** — generate random valid IDs, assert GET `/tissue_records/:id` returns 301 to `/tissue_records/:id/workspace`
    - **Validates: Requirements 8.3**
    - Use `httptest` and mock services; use gopter, minimum 100 iterations per property
    - Tag each test: `// Feature: tissuerecord-workspace, Property N: <text>`

- [x] 6. Register workspace routes in `cmd/api-server-gin/main.go`
  - Replace `r.GET("/tissue_records/:id", tissue_records.ViewTissueRecordHTML)` with `r.GET("/tissue_records/:id", tissue_records.RedirectToWorkspace)`
  - Add the nine new workspace routes:
    - `GET  /tissue_records/:id/workspace`
    - `GET  /tissue_records/:id/workspace/basic-info`
    - `PUT  /tissue_records/:id/workspace/basic-info`
    - `POST /tissue_records/:id/atlases/:atlasID`
    - `DELETE /tissue_records/:id/atlases/:atlasID`
    - `GET  /tissue_records/:id/atlases-section`
    - `POST /tissue_records/:id/categories/:categoryID`
    - `DELETE /tissue_records/:id/categories/:categoryID`
    - `GET  /tissue_records/:id/categories-section`
  - Update `logStartupInfo` to reflect the new routes
  - _Requirements: 1.1, 8.3_

- [x] 7. Update list row template `web/templates/includes/tr_row.html`
  - Replace the `<a href="/tissue_records/{{.ID}}">View</a>` anchor href with `/tissue_records/{{.ID}}/workspace`
  - Replace the HTMX inline-edit `<button>Edit</button>` with `<a href="/tissue_records/{{.ID}}/workspace" class="btn btn-ghost btn-sm">Edit</a>`
  - Remove the `hx-get`, `hx-target`, `hx-swap` attributes from the old Edit button
  - _Requirements: 8.1, 8.2_

- [x] 8. Checkpoint — build and tests pass
  - Run `go build ./...` and confirm zero errors
  - Run `go test ./...` and confirm all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Property tests use [gopter](https://github.com/leanovate/gopter) with a minimum of 100 iterations
- The `WorkspaceViewModel` is a local struct in `workspace.go`; no new package-level type is needed
- Both join tables (`atlas_tissue_records`, `tissue_record_categories`) already exist in the GORM migration models — no schema migration is required
