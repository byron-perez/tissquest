# Design Document — Annotation Session Management

## Overview

The current viewer fires an individual API call on every Annotorious event
(`createAnnotation`, `updateAnnotation`, `deleteAnnotation`). This design replaces
that behaviour with a **session model**: all annotation operations are intercepted and
written to an in-memory `Annotation_Store`; the database is only touched when the
researcher explicitly clicks the **Save_Button** ("Guardar anotaciones"), which sends
a single batch request to a new `POST /api/slides/:id/annotations/batch` endpoint that
applies all changes in one database transaction.

The scope is limited to the Virtual Microscope Viewer for tiled slides
(`slide_viewer.html`). No other pages are affected.

### Key goals

- **Decouple UI interactions from DB writes** — any number of draw/edit/delete
  operations during a session produce at most one DB transaction per save.
- **Atomic server-side apply** — the batch endpoint wraps everything in a single
  GORM/PostgreSQL transaction; partial writes are impossible.
- **Delete persisted annotations from the viewer** — a custom delete button wired to
  Annotorious selection events handles annotations that were saved in prior sessions,
  without needing a separate management UI.
- **Keep the implementation simple** — pure-JavaScript store with no framework, one
  new Go handler, one new route.

---

## Architecture

```
┌─────────────────────────────────────────────┐
│             Browser (slide_viewer.html)      │
│                                              │
│  ┌──────────────┐    events    ┌──────────┐ │
│  │  Annotorious │─────────────▶│ Annotation│ │
│  │  (OSD plugin)│◀─────────────│  Store   │ │
│  └──────────────┘  setAnnotations / remove  │ │
│                                ┌──────────┘ │
│  ┌──────────────┐    read diff │            │
│  │  Save_Button │─────────────▶│ batchSave()│ │
│  └──────────────┘              └────┬───────┘ │
└───────────────────────────────────┼───────────┘
                                    │ POST /api/slides/:id/annotations/batch
                                    ▼
┌───────────────────────────────────────────────┐
│              Go / Gin backend                  │
│                                               │
│  BatchSaveAnnotations handler                 │
│  ┌─────────────────────────────────────────┐  │
│  │  BEGIN TRANSACTION                      │  │
│  │    INSERT new annotations               │  │
│  │    UPDATE existing annotations          │  │
│  │    DELETE by annotorious_id             │  │
│  │  COMMIT  (or ROLLBACK on any error)     │  │
│  └─────────────────────────────────────────┘  │
│  → returns full persisted annotation list     │
└───────────────────────────────────────────────┘
```

The existing individual CRUD endpoints (`POST`, `PUT`, `DELETE
/api/slides/:id/annotations/:annotationID`) are **retained unchanged** for
potential future use (e.g., programmatic access), but the viewer no longer calls
them during a session.

---

## Components and Interfaces

### 2.1 Client — `AnnotationStore` (JavaScript module, inline in `slide_viewer.html`)

A single plain-JavaScript object created once per viewer load. It owns all
session state and exposes a small API consumed by the Annotorious event handlers
and the Save_Button.

```js
// Conceptual API (not a class; implemented as a factory function)
const store = createAnnotationStore(persistedAnnotations);

// Mutators — called from Annotorious event handlers
store.add(annotation);              // records a Pending_Annotation
store.update(annotation);           // updates Pending or existing
store.deletePending(annotoriousId); // removes a never-persisted annotation
store.deletePersisted(annotoriousId); // schedules removal of a saved annotation

// Queries
store.getDiff();   // → { created: [...], updated: [...], deletedIds: [...] }
store.isEmpty();   // → true if no diff entries exist
store.getAll();    // → all currently visible annotations (for Annotorious reload)

// Post-save
store.commitDiff(serverAnnotations); // clears diff; replaces persisted set
```

Internal state:

```js
{
  _persisted: Map<annotoriousId, annotationJSON>,  // loaded from server at init
  _created:   Map<annotoriousId, annotationJSON>,  // new in this session
  _updated:   Map<annotoriousId, annotationJSON>,  // edits to persisted items
  _deletedIds: Set<annotoriousId>,                 // persisted items to delete
}
```

Invariants maintained by the store:
- An `annotoriousId` appears in **at most one** of `_created`, `_updated`,
  `_deletedIds` at any time.
- Deleting a `_created` item removes it from `_created` and does **not** add it to
  `_deletedIds`.
- Editing a `_created` item updates it inside `_created` (it never moves to
  `_updated`).
- Editing a `_persisted` item moves its new state into `_updated`.

### 2.2 Annotorious Event Wiring

The three existing event handlers are **replaced**:

| Before | After |
|---|---|
| `anno.on("createAnnotation", …)` → `POST /api/…` | → `store.add(annotation)` |
| `anno.on("updateAnnotation", …)` → `PUT /api/…` | → `store.update(annotation)` |
| `anno.on("deleteAnnotation", …)` → `DELETE /api/…` | → `store.deletePending(annotation.id)` |

The `deleteAnnotation` Annotorious event fires when Annotorious's own internal
delete is used (e.g., pressing the Backspace key while an annotation is selected).
The custom Delete button (§2.4) calls `store.deletePersisted(id)` and then
`anno.removeAnnotation(id)` directly.

### 2.3 Save_Button

Added to the existing top bar in `slide_viewer.html`, positioned in the right-hand
flex group alongside the metadata span.

```html
<button id="btn-save-annotations"
        class="btn btn-sm btn-primary shrink-0"
        title="Guardar anotaciones">
  💾 Guardar anotaciones
</button>
```

States:
- **Idle** — default primary styling.
- **Saving** — button disabled + text changes to "Guardando…" within 200 ms of click
  (satisfies NFR-AS-5).
- **Success** — brief "✓ Guardado" flash (1.5 s), then returns to Idle.
- **Nothing to save** — brief "Sin cambios" flash (1.5 s), no request sent.
- **Error** — button re-enabled, "Error al guardar" shown in a dismissable alert
  overlay; diff is preserved.

### 2.4 Delete Button for Persisted Annotations

Annotorious v4's built-in popup is not reliably shown in this integration, so a
custom floating "Delete" button is used.

**Mechanism:**

1. Listen to Annotorious `clickAnnotation` (or `selectAnnotation`) event.
2. On selection, show a small floating button `#btn-delete-annotation` positioned
   near the annotation (or fixed in the toolbar overlay).
3. On deselection / canvas click, hide the button.
4. When clicked, the button calls `store.deletePersisted(selectedId)` (or
   `store.deletePending(selectedId)` for a newly created annotation) and then
   `anno.removeAnnotation(selectedId)`.

The button uses the same `#anno-toolbar` style block so it is visually consistent.

```html
<!-- added inside #anno-toolbar -->
<button id="btn-delete-annotation"
        title="Eliminar anotación"
        style="display:none">
  🗑️ Eliminar
</button>
```

### 2.5 Server — Batch Endpoint

**Route:** `POST /api/slides/:id/annotations/batch`

Registered in `cmd/api-server-gin/main.go` alongside the existing annotation
routes:

```go
r.POST("/api/slides/:id/annotations/batch", slides.BatchSaveAnnotations)
```

**Handler:** `cmd/api-server-gin/slides/annotations.go` — new function
`BatchSaveAnnotations`.

---

## Data Models

### 3.1 Batch Request Body (JSON)

```json
{
  "created": [
    { "id": "...", ... }   // full W3C annotation objects
  ],
  "updated": [
    { "id": "...", ... }   // full W3C annotation objects (by annotorious id)
  ],
  "deleted_ids": [
    "annotorious-uuid-1",
    "annotorious-uuid-2"
  ]
}
```

All three fields are required; each may be an empty array.

Go struct:

```go
type BatchSaveRequest struct {
    Created    []json.RawMessage `json:"created"`
    Updated    []json.RawMessage `json:"updated"`
    DeletedIDs []string          `json:"deleted_ids"`
}
```

### 3.2 Batch Response Body (JSON)

On success the server returns the **full current annotation list** for the slide
after applying the batch. This allows the client to call `store.commitDiff()` with
authoritative data and eliminates any client/server divergence.

```json
[
  { "id": "...", ... },
  ...
]
```

HTTP status: `200 OK`.

On error: `500 Internal Server Error` with body `{ "error": "…" }`.

### 3.3 Database Model (unchanged)

`AnnotationModel` in `internal/persistence/migration/annotation_model.go` is
unchanged. The batch handler reuses it directly:

```go
type AnnotationModel struct {
    gorm.Model
    SlideID        uint   `gorm:"not null;index"`
    AnnotoriousID  string `gorm:"column:annotorious_id;not null"`
    AnnotationJSON string `gorm:"column:annotation_json;type:text;not null"`
    DeletedAt      gorm.DeletedAt `gorm:"index"`
}
```

No schema migration is required for this feature.

---

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid
executions of a system — essentially, a formal statement about what the system should
do. Properties serve as the bridge between human-readable specifications and
machine-verifiable correctness guarantees.*

### Property 1: Store initialisation round-trip

*For any* set of persisted annotation objects, initialising the `Annotation_Store`
with those annotations and then calling `store.getAll()` SHALL return exactly the
same set of annotations (same IDs, same JSON content), and `store.getDiff()` SHALL
return an empty diff with no created, updated, or deleted entries.

**Validates: Requirements 1.1, 1.6**

---

### Property 2: Add goes to Pending_Annotation, not the diff's delete/update sets

*For any* annotation object added via `store.add(annotation)`, the annotation SHALL
appear in `getDiff().created`, SHALL NOT appear in `getDiff().updated`, and its ID
SHALL NOT appear in `getDiff().deletedIds`. No side-effects on `_persisted` or
`_updated` are produced.

**Validates: Requirements 1.2**

---

### Property 3: Edit routes correctly based on annotation origin

*For any* annotation already in the store (whether it arrived as persisted or was
created in the current session), calling `store.update(updated)` with a modified
version SHALL:
- Place the updated version in `getDiff().updated` if the annotation was originally
  persisted, OR keep it in `getDiff().created` (never in `updated`) if it was
  created in the current session.
- Produce no network calls.

**Validates: Requirements 1.3**

---

### Property 4: Deleting a Pending_Annotation leaves Pending_Deletions empty

*For any* annotation object that was added via `store.add()` (and never persisted),
calling `store.deletePending(id)` SHALL result in the ID being absent from all store
sets (`created`, `updated`, `deletedIds`), and `getDiff().deletedIds` SHALL remain
empty (no spurious Pending_Deletion entry is created).

**Validates: Requirements 1.4, 2.4**

---

### Property 5: Deleting a persisted annotation records a Pending_Deletion

*For any* annotorious ID that is present in the store's persisted set, calling
`store.deletePersisted(id)` SHALL add that ID to `getDiff().deletedIds`, remove the
annotation from `_persisted`, and produce no network calls.

**Validates: Requirements 2.2**

---

### Property 6: Successful save clears the diff

*For any* non-empty `Session_Diff`, after a successful call to `batchSave()` (mocked
server returning 200 with a full annotation list), `store.getDiff()` SHALL return an
empty diff and all previously pending annotations SHALL be present in the persisted
set.

**Validates: Requirements 3.4**

---

### Property 7: Failed save preserves the diff

*For any* non-empty `Session_Diff`, after a failed call to `batchSave()` (mocked
server returning a 5xx error), `store.getDiff()` SHALL be identical to the diff
before the call — no changes are lost.

**Validates: Requirements 3.5**

---

### Property 8: Batch endpoint response contains the full reconciled annotation set

*For any* valid batch request (arbitrary creates, updates, and deleted IDs against an
arbitrary pre-existing DB state for a slide), the batch endpoint SHALL return a JSON
array whose contents equal: (pre-existing annotations − deleted) ∪ (updated
versions) ∪ (created), with no duplicates.

**Validates: Requirements 4.5**

---

## Error Handling

| Scenario | Client behaviour | Server behaviour |
|---|---|---|
| Save_Button clicked, diff is empty | Show "Sin cambios" toast; no request sent | — |
| Network error / timeout during batch save | Show error alert; diff preserved | — |
| Server returns 4xx/5xx | Show error alert with message; diff preserved; button re-enabled | Return `{"error":"..."}` |
| `fetch` for initial annotation load fails | Show non-blocking error badge; viewer still usable, store starts empty | — |
| Batch transaction fails mid-way | — | GORM `tx.Rollback()`; return 500; no partial writes |
| Invalid JSON in created/updated array | — | Return 400 with field-level error before entering transaction |
| Unknown `annotoriousId` in updates/deletes | Treated as no-op (GORM `RowsAffected == 0` is acceptable) | Soft-skip; log warning; still commit if rest succeeds |

The error alert overlay is a small `<div id="anno-error-banner">` appended to the
viewer area, hidden by default, shown on error with a dismiss button. It is
positioned above the toolbar so it does not obscure the canvas.

---

## Testing Strategy

### Unit / property tests (JavaScript)

The `AnnotationStore` factory function is pure logic with no DOM or network
dependencies. It is the ideal target for property-based testing.

**PBT library:** [fast-check](https://fast-check.io) (well-maintained, no build
tooling required for inline tests; can be loaded from CDN for a test harness or
installed via npm in a `web/static/js/` test setup).

Each property test runs **≥ 100 iterations** with randomly generated annotation
objects (random IDs, random JSON bodies).

Test file: `web/static/js/annotation-store.test.js`

| Test | Property | Type |
|---|---|---|
| Init round-trip | Property 1 | PBT |
| Add to pending | Property 2 | PBT |
| Edit routes correctly | Property 3 | PBT |
| Delete pending leaves deletedIds empty | Property 4 | PBT |
| Delete persisted records deletion | Property 5 | PBT |
| Successful save clears diff | Property 6 | PBT |
| Failed save preserves diff | Property 7 | PBT |
| Save_Button with empty diff → no request | Req 3.3 | Example |
| Load failure → viewer still usable | Req 5.3 | Edge case |

### Integration tests (Go)

File: `cmd/api-server-gin/slides/annotations_batch_test.go`

| Test | Property | Type |
|---|---|---|
| Batch endpoint returns 200 + full list | Property 8 | PBT (using Go's `testing/quick` or `pgregory.net/rapid`) |
| Empty batch → 200, no DB writes | Req 4.6 | Example |
| Transaction rollback on injected failure | Req 4.3, 4.4 | Integration |
| Route exists (not 404) | Req 4.1 | Smoke |

**PBT library for Go:** [`pgregory.net/rapid`](https://pkg.go.dev/pgregory.net/rapid)
(idiomatic Go, no reflection magic, actively maintained).

Each property test runs ≥ 100 iterations.

Tag format for each test:
`// Feature: annotation-session-management, Property N: <property text>`
