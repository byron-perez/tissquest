# Implementation Plan: CRUD UI

## Overview

Incrementally add full CRUD capabilities to the TissQuest web interface using Go + Gin, server-side Go templates, HTMX, and DaisyUI. Each task builds on the previous, ending with all routes wired into the router and the base layout updated.

## Tasks

- [x] 1. Extend domain models and add validation
  - [x] 1.1 Add `ID`, `TissueRecordID` fields and `Validate()` to `internal/core/slide/slide.go`
    - Add `ErrEmptyName` and `ErrInvalidMagnification` error vars
    - Implement `Validate()` checking non-empty name and positive magnification
    - _Requirements: 3.2, 3.4_

  - [ ]* 1.2 Write unit tests for `slide.Validate()`
    - Test empty name → error, non-positive magnification → error, valid slide → nil
    - _Requirements: 3.4_

  - [x] 1.3 Add `Validate()` to `internal/core/taxon/taxon.go`
    - Add `ErrEmptyName` and `ErrInvalidRank` error vars
    - Implement `Validate()` checking non-empty name and valid rank enum
    - _Requirements: 4.4_

  - [ ]* 1.4 Write unit tests for `taxon.Validate()`
    - Test empty name → error, invalid rank → error, valid taxon → nil
    - _Requirements: 4.4_

  - [ ]* 1.5 Write unit tests for existing `atlas.Validate()` and `category.Validate()`
    - Cover empty name, name > 100 chars, invalid type, circular parent
    - _Requirements: 1.4, 1.5, 5.4, 5.5_

- [x] 2. Add `SlideRepository` interface and GORM implementation
  - [x] 2.1 Create `internal/core/slide/repository_interface.go`
    - Define `RepositoryInterface` with `Save`, `GetByID`, `Update`, `Delete`, `ListByTissueRecord`
    - _Requirements: 3.1, 3.3, 3.6, 3.7_

  - [x] 2.2 Create `internal/persistence/repositories/gorm_slide_repository.go`
    - Implement all five methods against GORM
    - Add `NewSlideRepository()` factory function
    - _Requirements: 3.3, 3.6, 3.7_

  - [x] 2.3 Add `NewSlideRepository()` to `internal/persistence/repositories/factory.go`
    - _Requirements: 3.3_

- [x] 3. Add GORM implementation for `taxon.RepositoryInterface`
  - [x] 3.1 Create `internal/persistence/repositories/gorm_taxon_repository.go`
    - Implement `Save`, `GetByID`, `GetLineage`, `ListByRank`, and add `List()` and `Delete()` methods
    - _Requirements: 4.3, 4.6, 4.7_

  - [x] 3.2 Add `NewTaxonRepository()` to `internal/persistence/repositories/factory.go` (update existing stub)
    - _Requirements: 4.3_

- [x] 4. Add `TaxonService` and `CategoryService`; extend `SlideService`
  - [x] 4.1 Create `internal/services/taxon_service.go`
    - Implement `Create`, `GetByID`, `Update`, `Delete`, `List`, `ListByRank`
    - Call `Validate()` before persist; return domain errors
    - _Requirements: 4.3, 4.4, 4.6, 4.7_

  - [x] 4.2 Create `internal/services/category_service.go`
    - Implement `Create`, `GetByID`, `Update`, `Delete`, `List`
    - Call `Validate()` before persist; return domain errors including `ErrCircularParent`
    - _Requirements: 5.3, 5.4, 5.5, 5.7, 5.8_

  - [x] 4.3 Extend `internal/services/slide_service.go`
    - Add `Create`, `GetByID`, `Update`, `Delete`, `ListByTissueRecord` methods
    - Call `Validate()` before persist
    - _Requirements: 3.3, 3.4, 3.6, 3.7_

- [x] 5. Create shared render helpers
  - [x] 5.1 Create `cmd/api-server-gin/shared/render.go`
    - Implement `IsHTMX`, `RenderPage`, `RenderFragment`, `RenderError`
    - `RenderPage` wraps in `base.html` for non-HTMX; calls `RenderFragment` for HTMX
    - _Requirements: 9.2, 9.3_

  - [x] 5.2 Create `cmd/api-server-gin/shared/flash.go`
    - Implement `SetFlash(c *gin.Context, message string)` setting `HX-Trigger` header as JSON
    - _Requirements: 7.1, 7.5, 9.8_

  - [ ]* 5.3 Write property test for `IsHTMX` / `RenderPage` fragment vs full page (Property 3)
    - **Property 3: HTMX requests receive fragments, direct requests receive full pages**
    - **Validates: Requirements 9.2, 9.3**

- [x] 6. Update base layout and navigation templates
  - [x] 6.1 Update `web/templates/layouts/base.html`
    - Add HTMX CDN `<script>` tag
    - Add `<div id="flash-region">` with `hx-on:show-flash` handler
    - _Requirements: 9.1, 7.3, 7.4_

  - [x] 6.2 Update `web/templates/includes/main-menu.html`
    - Add nav links for Atlases, TissueRecords, Taxa, Categories using `<a href>`
    - _Requirements: 8.1_

  - [x] 6.3 Create `web/templates/includes/flash.html`
    - Flash message region fragment for OOB swap
    - _Requirements: 7.3, 7.5, 9.6_

  - [x] 6.4 Create `web/templates/includes/confirm-delete.html`
    - Confirmation snippet with `hx-delete` confirm button and `hx-get` cancel button
    - _Requirements: 6.1, 6.2_

  - [x] 6.5 Create `web/templates/includes/delete-trigger.html`
    - Original delete button fragment (used by cancel to restore pre-confirmation state)
    - _Requirements: 6.3_

  - [x] 6.6 Create `web/templates/includes/breadcrumb.html`
    - Breadcrumb component accepting a slice of `{Label, URL}` pairs
    - _Requirements: 8.2_

- [x] 7. Implement Atlas CRUD handler and templates
  - [x] 7.1 Create `web/templates/pages/atlas_list.html`
    - List all atlases; each row has inline edit `hx-get` and delete trigger
    - Include breadcrumb and "New Atlas" button
    - _Requirements: 1.1, 8.2, 9.4, 9.5_

  - [x] 7.2 Create `web/templates/pages/atlas_form.html`
    - Single template for create and edit (nil `.Atlas` check for create mode)
    - Fields: name, description, category; inline error display; cancel `hx-get`
    - _Requirements: 1.2, 1.4, 1.5, 1.6, 1.8, 8.3_

  - [x] 7.3 Create `cmd/api-server-gin/atlas/atlas_crud.go`
    - Implement `NewAtlasForm`, `CreateAtlas`, `EditAtlasForm`, `UpdateAtlas`, `DeleteAtlas`, `ConfirmDeleteAtlas`
    - Use `shared.RenderPage`/`RenderFragment`, `shared.SetFlash`, return 422 on validation error, 404 on not found
    - OOB swap flash region on success
    - _Requirements: 1.1–1.10, 6.1–6.4, 7.1, 7.5, 9.2–9.10_

  - [ ]* 7.4 Write property test for Atlas mutation handlers (Property 1)
    - **Property 1: Successful mutations always set HX-Trigger**
    - **Validates: Requirements 1.3, 1.7, 1.9, 7.1, 9.8**

  - [ ]* 7.5 Write property test for Atlas form validation (Property 2)
    - **Property 2: Invalid form submissions always return HTTP 422**
    - **Validates: Requirements 1.4, 1.5, 1.8**

  - [ ]* 7.6 Write property test for Atlas confirm-delete endpoint (Property 4)
    - **Property 4: Confirmation snippet contains both confirm and cancel actions**
    - **Validates: Requirements 6.1, 6.2**

  - [ ]* 7.7 Write property test for Atlas cancel restores delete trigger (Property 5)
    - **Property 5: Cancel restores the original delete trigger**
    - **Validates: Requirements 6.3**

  - [ ]* 7.8 Write property test for Atlas OOB flash swap (Property 6)
    - **Property 6: Successful mutation responses include an OOB flash swap**
    - **Validates: Requirements 7.5, 9.6**

  - [ ]* 7.9 Write property test for Atlas error responses are fragments (Property 10)
    - **Property 10: Error responses are fragments, not full pages**
    - **Validates: Requirements 1.10, 9.10**

  - [ ]* 7.10 Write property test for Atlas form cancel link (Property 11)
    - **Property 11: Form fragments always contain a cancel link with hx-get**
    - **Validates: Requirements 8.3**

- [x] 8. Checkpoint — Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 9. Implement TissueRecord CRUD handler and templates
  - [x] 9.1 Create `web/templates/pages/tissue_record_list.html`
    - Paginated list (20/page); each row has inline edit and delete trigger; pagination controls
    - _Requirements: 2.1, 9.4, 9.5_

  - [x] 9.2 Create `web/templates/pages/tissue_record_form.html`
    - Fields: name, notes, optional taxon selector; inline errors; cancel `hx-get`
    - _Requirements: 2.2, 2.4, 2.5, 8.3_

  - [x] 9.3 Create `web/templates/pages/tissue_record_detail.html`
    - Detail page with breadcrumb and embedded Slide_Gallery region
    - _Requirements: 3.1, 8.2_

  - [x] 9.4 Create `cmd/api-server-gin/tissue_records/tissue_record_crud.go`
    - Implement `ListTissueRecords` (pagination), `NewTissueRecordForm`, `CreateTissueRecord`, `EditTissueRecordForm`, `UpdateTissueRecord`, `DeleteTissueRecord`, `ConfirmDeleteTissueRecord`, `ViewTissueRecord`
    - _Requirements: 2.1–2.8, 6.1–6.4, 7.1, 7.5, 9.2–9.10_

  - [ ]* 9.5 Write property test for TissueRecord pagination (Property 7)
    - **Property 7: TissueRecord list pagination never exceeds page size**
    - **Validates: Requirements 2.1**

  - [ ]* 9.6 Write property test for TissueRecord mutation handlers (Property 1)
    - **Property 1: Successful mutations always set HX-Trigger**
    - **Validates: Requirements 2.3, 2.6, 2.7, 7.1, 9.8**

  - [ ]* 9.7 Write property test for TissueRecord form validation (Property 2)
    - **Property 2: Invalid form submissions always return HTTP 422**
    - **Validates: Requirements 2.4, 9.7**

- [x] 10. Implement Slide CRUD handler and templates
  - [x] 10.1 Create `web/templates/pages/slide_form.html`
    - Fields: name, image URL, magnification, staining, inclusion method, reagents, protocol, notes; inline errors; cancel `hx-get`
    - _Requirements: 3.2, 3.4, 3.5, 8.3_

  - [x] 10.2 Create `cmd/api-server-gin/slides/slides_crud.go`
    - Implement `CreateSlide`, `EditSlideForm`, `UpdateSlide`, `DeleteSlide`, `ConfirmDeleteSlide`
    - Slide is always scoped to a parent TissueRecord; return updated Slide_Gallery fragment on success
    - _Requirements: 3.1–3.8, 6.1–6.4, 7.1, 7.5, 9.2–9.10_

  - [ ]* 10.3 Write property test for Slide mutation handlers (Property 1)
    - **Property 1: Successful mutations always set HX-Trigger**
    - **Validates: Requirements 3.3, 3.6, 3.7, 7.1, 9.8**

  - [ ]* 10.4 Write property test for Slide form validation (Property 2)
    - **Property 2: Invalid form submissions always return HTTP 422**
    - **Validates: Requirements 3.4, 9.7**

- [x] 11. Implement Taxon CRUD handler and templates
  - [x] 11.1 Create `web/templates/pages/taxon_list.html`
    - List taxa grouped by rank; each row has inline edit and delete trigger
    - _Requirements: 4.1, 9.4, 9.5_

  - [x] 11.2 Create `web/templates/pages/taxon_form.html`
    - Fields: name, rank selector, optional parent taxon selector; inline errors; cancel `hx-get`
    - _Requirements: 4.2, 4.4, 4.5, 8.3_

  - [x] 11.3 Create `cmd/api-server-gin/taxa/taxa.go`
    - Implement `ListTaxa`, `NewTaxonForm`, `CreateTaxon`, `EditTaxonForm`, `UpdateTaxon`, `DeleteTaxon`, `ConfirmDeleteTaxon`
    - _Requirements: 4.1–4.8, 6.1–6.4, 7.1, 7.5, 9.2–9.10_

  - [ ]* 11.4 Write property test for Taxon list grouped by rank (Property 9)
    - **Property 9: Taxon list groups taxa by rank**
    - **Validates: Requirements 4.1**

  - [ ]* 11.5 Write property test for Taxon mutation handlers (Property 1)
    - **Property 1: Successful mutations always set HX-Trigger**
    - **Validates: Requirements 4.3, 4.6, 4.7, 7.1, 9.8**

  - [ ]* 11.6 Write property test for Taxon form validation (Property 2)
    - **Property 2: Invalid form submissions always return HTTP 422**
    - **Validates: Requirements 4.4, 9.7**

- [x] 12. Implement Category CRUD handler and templates
  - [x] 12.1 Create `web/templates/pages/category_list.html`
    - List categories with type and optional parent; each row has inline edit and delete trigger
    - _Requirements: 5.1, 9.4, 9.5_

  - [x] 12.2 Create `web/templates/pages/category_form.html`
    - Fields: name, type selector, description, optional parent category selector; inline errors; cancel `hx-get`
    - _Requirements: 5.2, 5.4, 5.5, 5.6, 8.3_

  - [x] 12.3 Create `cmd/api-server-gin/categories/categories.go`
    - Implement `ListCategories`, `NewCategoryForm`, `CreateCategory`, `EditCategoryForm`, `UpdateCategory`, `DeleteCategory`, `ConfirmDeleteCategory`
    - _Requirements: 5.1–5.9, 6.1–6.4, 7.1, 7.5, 9.2–9.10_

  - [ ]* 12.4 Write property test for circular parent rejection (Property 8)
    - **Property 8: Circular parent reference is always rejected**
    - **Validates: Requirements 5.5**

  - [ ]* 12.5 Write property test for Category mutation handlers (Property 1)
    - **Property 1: Successful mutations always set HX-Trigger**
    - **Validates: Requirements 5.3, 5.7, 5.8, 7.1, 9.8**

  - [ ]* 12.6 Write property test for Category form validation (Property 2)
    - **Property 2: Invalid form submissions always return HTTP 422**
    - **Validates: Requirements 5.4, 9.7**

- [x] 13. Checkpoint — Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 14. Wire all routes into the router
  - [x] 14.1 Update `cmd/api-server-gin/main.go` `setupRouter`
    - Register all new routes for atlases, tissue_records, slides, taxa, categories as specified in the design
    - Remove or convert existing JSON-returning routes to HTML
    - _Requirements: 9.1–9.10_

  - [x] 14.2 Update `logStartupInfo` in `cmd/api-server-gin/main.go`
    - Log all new routes
    - _Requirements: 9.1_

- [x] 15. Final checkpoint — Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for a faster MVP
- Property tests use the [rapid](https://github.com/flyingmutant/rapid) library (minimum 100 iterations per property)
- Each property test file should include a comment: `// Feature: crud-ui, Property N: <property text>`
- All form submissions use `application/x-www-form-urlencoded` except slide image upload (`multipart/form-data`)
- HTMX version 2 is loaded from `https://unpkg.com/htmx.org@2`
