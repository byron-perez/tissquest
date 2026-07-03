# Implementation Plan: Annotation Session Management

## Overview

Replace the viewer's per-event API calls with a session model: an in-memory
`AnnotationStore` accumulates all annotation operations, and a single Save_Button
transmits the accumulated diff to a new `BatchSaveAnnotations` endpoint that applies
every change in one database transaction. A custom delete button in the viewer
toolbar handles deletion of persisted annotations.

The implementation proceeds bottom-up: Go batch endpoint → JS store module →
Annotorious wiring + Save_Button → Delete button → property tests.

## Tasks

- [-] 1. Add the `BatchSaveAnnotations` handler and route (Go)
  - [x] 1.1 Define `BatchSaveRequest` struct and implement `BatchSaveAnnotations` in `cmd/api-server-gin/slides/annotations.go`
    - Add `BatchSaveRequest` struct with fields `Created []json.RawMessage`, `Updated []json.RawMessage`, `DeletedIDs []string` (json tag `deleted_ids`)
    - Implement `BatchSaveAnnotations(c *gin.Context)`:
      - Parse `:id` as `slideID`
      - Bind JSON body to `BatchSaveRequest`; return 400 on parse error with field-level message
      - Open `db.Transaction(func(tx *gorm.DB) error { … })` that:
        - Iterates `Created`: extracts `annotorious_id` from each `json.RawMessage`, inserts an `AnnotationModel`
        - Iterates `Updated`: updates `annotation_json` where `slide_id = ? AND annotorious_id = ?`
        - Iterates `DeletedIDs`: soft-deletes where `slide_id = ? AND annotorious_id = ?`
      - On transaction error return 500 `{"error":"…"}` (GORM rolls back automatically)
      - On success query the full current annotation list for the slide and return 200 with the JSON array
    - Return 400 with a clear error when a `Created` entry is missing its `id` field
    - Unknown IDs in `Updated`/`DeletedIDs` are treated as no-ops (log a warning, continue)
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6_

  - [ ] 1.2 Register the route in `cmd/api-server-gin/main.go`
    - Add `r.POST("/api/slides/:id/annotations/batch", slides.BatchSaveAnnotations)` immediately after the existing `r.DELETE("/api/slides/:id/annotations/:annotationID", …)` line
    - Add a corresponding log line in `logStartupInfo()`
    - _Requirements: 4.1_

- [x] 2. Checkpoint — Go backend compiles and route responds
  - Ensure `go build ./cmd/api-server-gin/...` succeeds with no errors.
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 3. Create the `AnnotationStore` JS module
  - [ ] 3.1 Create `web/static/js/annotation-store.js` with the `createAnnotationStore` factory
    - Implement internal state: `_persisted` (Map), `_created` (Map), `_updated` (Map), `_deletedIds` (Set)
    - Implement `add(annotation)`: add to `_created`; must not touch `_persisted`, `_updated`, or `_deletedIds`
    - Implement `update(annotation)`:
      - If `annotation.id` is in `_created` → update in `_created` (stays there, never moves to `_updated`)
      - If `annotation.id` is in `_persisted` → put updated version into `_updated`
    - Implement `deletePending(id)`: remove from `_created` only; do NOT add to `_deletedIds`
    - Implement `deletePersisted(id)`: remove from `_persisted` and `_updated`; add to `_deletedIds`
    - Implement `getDiff()`: returns `{ created: [..._created.values()], updated: [..._updated.values()], deletedIds: [..._deletedIds] }`
    - Implement `isEmpty()`: returns `true` when `_created`, `_updated`, and `_deletedIds` are all empty
    - Implement `getAll()`: returns merged array of `_persisted` items (excluding `_deletedIds`) plus `_created` items plus `_updated` items (overriding their `_persisted` version)
    - Implement `commitDiff(serverAnnotations)`: clears `_created`, `_updated`, `_deletedIds`; replaces `_persisted` with a new Map built from `serverAnnotations`
    - Export via `export { createAnnotationStore }` (ES module) AND attach to `window.createAnnotationStore` for non-module use in the viewer
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.6, 2.2, 2.4, 3.4_

  - [ ]* 3.2 Write property-based tests for `AnnotationStore` in `web/static/js/annotation-store.test.js`
    - Use fast-check loaded from CDN (no build tooling); run with Node's `--experimental-vm-modules` or a plain `node` runner
    - **Property 1: Store initialisation round-trip** — `getAll()` returns the same annotations used at init, `getDiff()` is empty
      - **Validates: Requirements 1.1, 1.6**
    - **Property 2: Add goes to `created`, not `updated`/`deletedIds`** — after `add(a)`, `getDiff().created` contains `a`, `getDiff().updated` is empty, `getDiff().deletedIds` is empty
      - **Validates: Requirements 1.2**
    - **Property 3: Edit routes by origin** — updating a persisted annotation puts it in `updated`; updating a pending annotation keeps it in `created`
      - **Validates: Requirements 1.3**
    - **Property 4: Deleting a pending annotation leaves `deletedIds` empty** — after `add(a)` then `deletePending(a.id)`, all sets are empty
      - **Validates: Requirements 1.4, 2.4**
    - **Property 5: Deleting a persisted annotation records a Pending_Deletion** — after `deletePersisted(id)`, `getDiff().deletedIds` contains `id` and `getAll()` does not contain the annotation
      - **Validates: Requirements 2.2**
    - **Property 6: Successful save clears the diff** — after `commitDiff(serverAnnotations)`, `getDiff()` is empty and `getAll()` reflects `serverAnnotations`
      - **Validates: Requirements 3.4**
    - **Property 7: Failed save preserves the diff** — if `batchSave()` mock returns 5xx, the diff before and after the call is identical
      - **Validates: Requirements 3.5**
    - Each property runs ≥ 100 fast-check iterations
    - Tag each test: `// Feature: annotation-session-management, Property N: <title>`

- [x] 4. Checkpoint — AnnotationStore module is self-consistent
  - Run `node web/static/js/annotation-store.test.js` (or equivalent); all property tests must pass.
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 5. Integrate `AnnotationStore` into the viewer and wire Annotorious events
  - [x] 5.1 Load `annotation-store.js` and initialise the store in `slide_viewer.html`
    - Add `<script src="/static/js/annotation-store.js"></script>` in the `{{define "head"}}` block (before Annotorious scripts)
    - After the `fetch("/api/slides/…/annotations")` success callback, replace the `anno.setAnnotations(annotations)` call with:
      1. `const store = createAnnotationStore(annotations);`
      2. `anno.setAnnotations(store.getAll());`
    - Handle load failure per Req 5.3: on `fetch` rejection, initialise `store = createAnnotationStore([])` and display `#anno-error-banner` with a non-blocking message
    - _Requirements: 1.1, 1.6, 5.1, 5.2, 5.3_

  - [x] 5.2 Replace the three `anno.on(…)` handlers with store method calls
    - Replace `anno.on("createAnnotation", …)` → `store.add(annotation)` (remove the `fetch POST`)
    - Replace `anno.on("updateAnnotation", …)` → `store.update(annotation)` (remove the `fetch PUT`)
    - Replace `anno.on("deleteAnnotation", …)` → `store.deletePending(annotation.id)` (remove the `fetch DELETE`)
    - _Requirements: 1.2, 1.3, 1.4_

- [ ] 6. Add the Save_Button and `batchSave()` function
  - [x] 6.1 Add Save_Button markup to the viewer top bar
    - Insert `<button id="btn-save-annotations" class="btn btn-sm btn-primary shrink-0" title="Guardar anotaciones">💾 Guardar anotaciones</button>` into the right-hand flex group of the top bar (alongside the existing metadata span)
    - Add `#anno-error-banner` div: `<div id="anno-error-banner" style="display:none" class="..."></div>` positioned above the toolbar; include a dismiss button
    - _Requirements: 3.1_

  - [x] 6.2 Implement `batchSave()` and wire it to the Save_Button
    - Implement `batchSave()` async function:
      1. If `store.isEmpty()`: flash "Sin cambios" on the button for 1.5 s; return (no request sent)
      2. Within 200 ms of click: disable button, change text to "Guardando…" (satisfies NFR-AS-5)
      3. Call `fetch("/api/slides/" + slideID + "/annotations/batch", { method: "POST", headers: {"Content-Type": "application/json"}, body: JSON.stringify(store.getDiff()) })`
      4. On success (2xx): call `store.commitDiff(responseAnnotations)`; flash "✓ Guardado" for 1.5 s; re-enable button
      5. On 4xx/5xx or network error: re-enable button; show `#anno-error-banner` with the error message; diff is preserved (Req 3.5)
    - Attach `batchSave` to `document.getElementById("btn-save-annotations").addEventListener("click", batchSave)`
    - _Requirements: 3.2, 3.3, 3.4, 3.5, NFR-AS-5_

- [ ] 7. Add the Delete button for persisted annotations
  - [x] 7.1 Add `#btn-delete-annotation` to `#anno-toolbar`
    - Insert `<button id="btn-delete-annotation" title="Eliminar anotación" style="display:none">🗑️ Eliminar</button>` as the last child inside `<div id="anno-toolbar">`
    - _Requirements: 2.1_

  - [x] 7.2 Wire selection events and implement delete logic
    - Declare `let selectedAnnotationId = null;` in the viewer IIFE scope
    - Listen to `anno.on("clickAnnotation", function(annotation) { … })` (or `selectAnnotation` if available): set `selectedAnnotationId = annotation.id`; show `#btn-delete-annotation`
    - Listen to deselection (canvas click outside annotation or a `clearSelection` event): set `selectedAnnotationId = null`; hide `#btn-delete-annotation`
    - Wire `#btn-delete-annotation` click:
      - If `selectedAnnotationId` is in the current session's `_created` set → call `store.deletePending(selectedAnnotationId)` then `anno.removeAnnotation(selectedAnnotationId)` (Req 1.4 / Req 2.4)
      - Otherwise → call `store.deletePersisted(selectedAnnotationId)` then `anno.removeAnnotation(selectedAnnotationId)` (Req 2.2, 2.3)
      - Hide `#btn-delete-annotation`; reset `selectedAnnotationId = null`
    - _Requirements: 2.1, 2.2, 2.3, 2.4_

- [ ] 8. Final checkpoint — end-to-end wiring complete
  - Ensure `go build ./cmd/api-server-gin/...` succeeds.
  - Verify the viewer HTML has no broken template expressions.
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 9. Write Go integration tests for the batch endpoint
  - [ ]* 9.1 Write a property-based test for the batch endpoint response reconciliation
    - File: `cmd/api-server-gin/slides/annotations_batch_test.go`
    - Use `pgregory.net/rapid` (add to `go.mod`/`go.sum`)
    - **Property 8: Batch endpoint returns the full reconciled annotation set** — for any valid batch request against an arbitrary pre-existing DB state, the response equals (pre-existing − deleted) ∪ updated ∪ created with no duplicates
    - **Validates: Requirements 4.5**
    - Run ≥ 100 rapid iterations
    - Tag: `// Feature: annotation-session-management, Property 8: Batch endpoint response contains the full reconciled annotation set`
  - [ ]* 9.2 Write example-based tests for edge cases
    - Empty batch → 200, no DB writes (Req 4.6)
    - Transaction rollback on injected failure → 500, no partial writes (Req 4.3, 4.4)
    - Route exists (not 404) smoke test (Req 4.1)
    - Invalid JSON in `created` → 400 before transaction (Req 4.2)

## Notes

- Tasks marked with `*` are optional and can be skipped for a faster MVP
- The three existing individual CRUD endpoints (`POST`, `PUT`, `DELETE /api/slides/:id/annotations/:annotationID`) are left unchanged; only the viewer is updated to stop calling them
- No schema migration is required
- Each task references the specific requirements or properties it satisfies for traceability
- Property tests validate universal correctness; unit/integration tests cover examples and edge cases
