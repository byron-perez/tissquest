# Requirements Document — Annotation Session Management

## Introduction

The Virtual Microscope Viewer currently persists every annotation operation (draw,
edit, delete) to the database as it happens. Because the database is a constrained
resource, this one-write-per-event strategy is too expensive for iterative annotation
workflows where a researcher may draw, adjust, and delete many shapes before settling
on a final set.

A **Researcher** is any user of the platform who has a working knowledge of histology
— professional or amateur — and who uploads and annotates microscopy images for study
or sharing. Researchers are the primary authors of annotation content.

This feature introduces an **annotation session**: a bounded period of work during
which all annotation operations are held in browser memory. The researcher commits the
entire accumulated set of changes to the server through a single explicit save action.
The server applies all changes atomically. Until the save is triggered, the database
is untouched.

The feature also provides a way to delete any annotation — whether it was saved in a
previous session or created in the current one — directly from the viewer, without
leaving the slide.

This feature is scoped to the Virtual Microscope Viewer page for tiled slides.

---

## Glossary

- **Researcher**: A user with working knowledge of histology who uses the platform to
  upload, view, and annotate microscopy images. May be a professional or an
  enthusiast. The primary stakeholder for annotation authoring.

- **Annotation_Session**: The period between the viewer loading a slide's annotations
  and the researcher explicitly saving. During this period all annotation operations
  affect only the in-memory state.

- **Pending_Annotation**: An annotation that has been created or modified in the
  current Annotation_Session but not yet persisted to the database.

- **Pending_Deletion**: A record of intent to remove an annotation from the database,
  held in memory for the duration of the Annotation_Session.

- **Session_Diff**: The complete set of changes accumulated during one
  Annotation_Session: new annotations, updated annotations, and the IDs of annotations
  to be deleted.

- **Batch_Save**: The single user-initiated action that transmits the Session_Diff to
  the server and causes the server to apply all changes in one database transaction.

- **Viewer**: The Virtual Microscope Viewer as defined in the main requirements
  specification. The component the researcher interacts with directly.

- **Annotation_Store**: The client-side in-memory structure that tracks the
  Session_Diff throughout an Annotation_Session.

- **Save_Button**: The UI control labeled "Guardar anotaciones" that triggers a
  Batch_Save.

---

## Requirements

### Requirement 1 — In-Memory Annotation Session

**User Story:** As a researcher, I want my annotation changes to stay in the browser
until I choose to save them, so that I can draw and adjust freely without triggering
database writes on every action.

#### Acceptance Criteria

1. WHEN the Viewer loads a slide, THE Annotation_Store SHALL be initialised with the
   annotations already persisted for that slide.

2. WHEN a researcher creates an annotation, THE Annotation_Store SHALL record it as a
   Pending_Annotation without sending any request to the server.

3. WHEN a researcher edits an existing annotation, THE Annotation_Store SHALL update
   the stored annotation data without sending any request to the server.

4. WHEN a researcher deletes an annotation that is a Pending_Annotation, THE
   Annotation_Store SHALL remove it from the Pending_Annotation set without sending
   any request to the server.

5. WHILE an Annotation_Session is active, THE Viewer SHALL reflect the current state
   of the Annotation_Store so that the researcher sees all their changes immediately.

6. WHEN the Viewer loads a slide, THE Annotation_Store SHALL contain no Session_Diff
   entries — the session starts clean.

---

### Requirement 2 — Delete Saved Annotations from the Viewer

**User Story:** As a researcher, I want to delete a previously saved annotation
directly from the viewer, so that I can correct mistakes without a separate management
interface.

#### Acceptance Criteria

1. WHEN a researcher selects an annotation that was persisted in a previous session,
   THE Viewer SHALL present a delete action for that annotation.

2. WHEN a researcher invokes the delete action on a persisted annotation, THE
   Annotation_Store SHALL record it as a Pending_Deletion without sending any request
   to the server.

3. WHEN an annotation is recorded as a Pending_Deletion, THE Viewer SHALL no longer
   display that annotation.

4. WHEN a researcher invokes the delete action on a Pending_Annotation, THE
   Annotation_Store SHALL remove it from the Pending_Annotation set (Requirement 1
   AC 4 applies; no Pending_Deletion entry is needed for an annotation that was never
   persisted).

---

### Requirement 3 — Batch Save

**User Story:** As a researcher, I want a single "Guardar anotaciones" button that
commits all my pending changes at once, so that I can save at natural checkpoints
without worrying about partial writes.

#### Acceptance Criteria

1. THE Viewer SHALL display the Save_Button at all times while a tiled slide is open.

2. WHEN the researcher activates the Save_Button, THE Viewer SHALL transmit the
   complete Session_Diff to the server in a single request.

3. WHEN the Save_Button is activated and the Session_Diff is empty, THE Viewer SHALL
   indicate to the researcher that there is nothing new to save.

4. IF the server returns a success response, THEN THE Annotation_Store SHALL clear the
   Session_Diff and mark all previously Pending_Annotations as persisted.

5. IF the server returns an error response, THEN THE Viewer SHALL display an error
   message and preserve the Session_Diff so the researcher does not lose unsaved work.

---

### Requirement 4 — Atomic Server-Side Batch Write

**User Story:** As a researcher, I want all my pending annotation changes to be
applied together or not at all, so that the database never holds a partial or
inconsistent set of annotations for a slide.

#### Acceptance Criteria

1. THE Annotation_API SHALL expose an endpoint that accepts the Session_Diff for a
   given slide in a single request.

2. THE Annotation_API endpoint SHALL accept three distinct components in the
   Session_Diff payload: new annotations to create, existing annotations to update,
   and annotation identifiers to delete.

3. WHEN the Annotation_API receives a Batch_Save request, THE Annotation_API SHALL
   apply all creates, updates, and deletes within a single database transaction.

4. IF any operation within the transaction fails, THEN THE Annotation_API SHALL roll
   back all changes and return an error response indicating that no changes were
   persisted.

5. WHEN the Annotation_API successfully completes a Batch_Save, THE Annotation_API
   SHALL return a response that includes all persisted annotations so the client can
   reconcile its local state.

6. WHEN the Session_Diff is empty, THE Annotation_API SHALL accept the request and
   return a success response without performing any database write.

---

### Requirement 5 — Annotation Load on Viewer Open

**User Story:** As a researcher, I want the viewer to show all annotations that were
previously saved for a slide when I open it, so that I can resume work from a prior
session.

#### Acceptance Criteria

1. WHEN the Viewer opens a tiled slide, THE Viewer SHALL retrieve all persisted
   annotations for that slide from the server before the researcher can interact
   with them.

2. WHEN the server returns an empty annotation list, THE Viewer SHALL initialise with
   no annotations and be ready for new ones.

3. IF the annotation retrieval request fails, THEN THE Viewer SHALL display an error
   indicator and allow the researcher to continue using the viewer without annotations
   rather than blocking the entire viewer.

---

## Non-Functional Requirements

### NFR-AS-1: Database Write Reduction

THE Annotation_Session model SHALL reduce the number of database write operations per
annotation workflow to at most one transaction per explicit save, regardless of how
many individual annotation operations the researcher performed during the session.

### NFR-AS-2: Session State Durability on Client

WHILE an Annotation_Session is active, THE Annotation_Store SHALL retain the complete
Session_Diff in memory for the full duration of the browser session so that no pending
work is silently lost due to internal viewer events.

### NFR-AS-3: Atomicity

THE Annotation_API SHALL guarantee that a Batch_Save request either applies all
changes or applies none. Partial writes to the database are not acceptable.

### NFR-AS-4: Responsiveness of Local Operations

WHEN the researcher performs any annotation operation (create, edit, delete) that
modifies only the Annotation_Store, THE Viewer SHALL reflect the change within the
current animation frame, with no perceptible delay introduced by the session
management layer.

### NFR-AS-5: Save Feedback Latency

WHEN the researcher activates the Save_Button, THE Viewer SHALL provide a visual
acknowledgement of the in-progress save within 200 ms of the activation event.

---

## Acceptance Criteria (Feature-Level)

- A researcher can draw multiple annotations, edit one, delete another, and activate
  the Save_Button; the server receives exactly one request containing the net diff.
- A researcher can select a previously saved annotation and delete it; after saving,
  the annotation is absent from the database and does not reappear on the next load.
- If the network is unavailable when the researcher activates the Save_Button, the
  Session_Diff is preserved and the researcher sees an error message.
- Activating the Save_Button with no pending changes produces a "nothing to save"
  indicator rather than an empty server request.
- Opening the viewer on a slide with existing annotations loads those annotations
  before the researcher can interact with them.
