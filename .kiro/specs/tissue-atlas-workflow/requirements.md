# Requirements Document

## Introduction

This feature covers the complete content-manager workflow for populating a TissQuest Atlas: creating a TissueRecord with taxon and notes, attaching slides to it, and then linking it to an Atlas. It also addresses the cramped inline-row editing UX by replacing it with a modal-based edit experience. The student-facing Atlas view must reflect the current set of linked TissueRecords accurately.

Two concrete problems are solved:

1. **Editing UX** — the inline edit form that opens inside a narrow table row is replaced with a spacious modal dialog, applied consistently to TissueRecords (and as a pattern available to other entities).
2. **Atlas ↔ TissueRecord linking** — a content manager can add and remove TissueRecords from an Atlas through the Atlas management UI; the student-facing Atlas view renders only the currently linked records.

## Glossary

- **Atlas**: A named, categorised collection of TissueRecords intended for student study.
- **TissueRecord**: A specimen entry containing a name, optional notes, an optional Taxon, and zero or more Slides.
- **Slide**: A microscopy image attached to a TissueRecord, carrying magnification and preparation metadata.
- **Taxon**: A node in the taxonomic hierarchy (e.g. species, genus) that classifies a TissueRecord.
- **Content_Manager**: An authenticated or trusted operator who creates and curates Atlas content.
- **Student**: A read-only user who browses published Atlas content.
- **Atlas_Detail_Page**: The student-facing page at `/atlases/:id` that displays an Atlas and its linked TissueRecords.
- **Atlas_Management_Page**: The content-manager page at `/atlases/:id/manage` (or equivalent) where TissueRecords are linked and unlinked.
- **Edit_Modal**: A DaisyUI modal dialog that hosts an edit form, replacing the inline table-row form.
- **Link_Panel**: The UI section on the Atlas_Management_Page that lists available TissueRecords and allows linking/unlinking.
- **Slide_Gallery**: The existing HTMX-driven slide management component on the TissueRecord detail page.
- **HTMX**: The hypermedia library used for all partial-page interactions (no full reloads).

---

## Requirements

### Requirement 1: Modal-Based TissueRecord Editing

**User Story:** As a Content_Manager, I want to edit a TissueRecord in a modal dialog, so that the form is not cramped inside a narrow table row.

#### Acceptance Criteria

1. WHEN a Content_Manager clicks the Edit button for a TissueRecord in the list, THE Edit_Modal SHALL open and display the pre-populated edit form for that TissueRecord.
2. THE Edit_Modal SHALL display the Name, Notes, and Taxon fields with sufficient width to be readable and usable.
3. WHEN a Content_Manager submits a valid edit form inside the Edit_Modal, THE System SHALL save the changes, close the Edit_Modal, and update the corresponding table row in place without a full page reload.
4. WHEN a Content_Manager submits an edit form with an empty Name field, THE Edit_Modal SHALL remain open and display a validation error message adjacent to the Name field.
5. WHEN a Content_Manager clicks Cancel or the modal backdrop, THE Edit_Modal SHALL close without saving any changes.
6. WHILE the Edit_Modal is open, THE System SHALL prevent interaction with the page content behind the modal.

---

### Requirement 2: TissueRecord Creation

**User Story:** As a Content_Manager, I want to create a new TissueRecord with a name, notes, and taxon, so that I have a specimen entry ready to receive slides.

#### Acceptance Criteria

1. THE System SHALL provide a form to create a TissueRecord with the following fields: Name (required), Notes (optional), Taxon (optional, selected from existing taxa).
2. WHEN a Content_Manager submits the creation form with a valid Name, THE System SHALL persist the new TissueRecord and redirect the Content_Manager to the TissueRecord list.
3. WHEN a Content_Manager submits the creation form with an empty Name, THE System SHALL return the form with a validation error and preserve the entered values for Notes and Taxon.
4. IF the Taxon list is empty, THEN THE System SHALL still allow TissueRecord creation with no Taxon selected.

---

### Requirement 3: Slide Attachment to a TissueRecord

**User Story:** As a Content_Manager, I want to add one or more slides to a TissueRecord, so that the record has microscopy images for students to study.

#### Acceptance Criteria

1. WHEN a Content_Manager views the TissueRecord detail page, THE Slide_Gallery SHALL display all slides currently attached to that TissueRecord.
2. THE Slide_Gallery SHALL provide an inline form to add a new Slide with the following fields: Name (required), URL (required), Magnification (required, positive integer), Staining (optional).
3. WHEN a Content_Manager submits a valid add-slide form, THE Slide_Gallery SHALL update in place via HTMX to show the newly added Slide without a full page reload.
4. WHEN a Content_Manager submits an add-slide form with a missing required field, THE Slide_Gallery SHALL display a validation error and preserve the entered values.
5. WHEN a Content_Manager clicks Remove on a Slide, THE System SHALL delete that Slide and THE Slide_Gallery SHALL update in place to remove it without a full page reload.
6. IF a TissueRecord has no slides, THEN THE Slide_Gallery SHALL display an empty-state message indicating no slides are attached.

---

### Requirement 4: Linking a TissueRecord to an Atlas

**User Story:** As a Content_Manager, I want to add a TissueRecord to an Atlas, so that students can find it when browsing that Atlas.

#### Acceptance Criteria

1. WHEN a Content_Manager navigates to the Atlas_Management_Page for an Atlas, THE Link_Panel SHALL display a list of all TissueRecords not yet linked to that Atlas.
2. WHEN a Content_Manager clicks "Add to Atlas" for a TissueRecord in the Link_Panel, THE System SHALL create the association in the `atlas_tissue_records` join table and THE Link_Panel SHALL update in place via HTMX to move that TissueRecord to the linked section.
3. THE Atlas_Management_Page SHALL display a separate section listing all TissueRecords currently linked to the Atlas.
4. IF a TissueRecord is already linked to the Atlas, THEN THE System SHALL not create a duplicate association.
5. WHEN a Content_Manager clicks "Remove from Atlas" for a linked TissueRecord, THE System SHALL delete the association from the `atlas_tissue_records` join table and THE Link_Panel SHALL update in place via HTMX to move that TissueRecord back to the available section.

---

### Requirement 5: Student-Facing Atlas View

**User Story:** As a Student, I want to see all tissue records linked to an Atlas on the Atlas detail page, so that I can study the specimens in that collection.

#### Acceptance Criteria

1. WHEN a Student navigates to the Atlas_Detail_Page, THE Atlas_Detail_Page SHALL display only the TissueRecords currently linked to that Atlas.
2. FOR EACH linked TissueRecord, THE Atlas_Detail_Page SHALL display: the record name, the taxon lineage badges (if a Taxon is assigned), the notes excerpt, the count of attached slides, and a thumbnail of the first Slide (if any slides exist).
3. WHEN a Student clicks on a TissueRecord card on the Atlas_Detail_Page, THE System SHALL navigate to the TissueRecord detail page for that record.
4. IF an Atlas has no linked TissueRecords, THEN THE Atlas_Detail_Page SHALL display an empty-state message.
5. WHILE the Atlas_Detail_Page is loading TissueRecord data, THE System SHALL not display stale or cached data from a previous Atlas.

---

### Requirement 6: Atlas Management Navigation

**User Story:** As a Content_Manager, I want to reach the Atlas_Management_Page from the Atlas list and Atlas detail page, so that I can manage the content of any Atlas without hunting for the right URL.

#### Acceptance Criteria

1. THE Atlas list page SHALL display a "Manage" link or button for each Atlas that navigates to the Atlas_Management_Page for that Atlas.
2. THE Atlas_Detail_Page SHALL display a "Manage" link or button that navigates to the Atlas_Management_Page for the current Atlas.
3. THE Atlas_Management_Page SHALL display breadcrumb navigation in the form: Home → Atlases → [Atlas Name] → Manage.
4. THE Atlas_Management_Page SHALL display a link back to the Atlas_Detail_Page.

---

### Requirement 7: End-to-End Content Workflow Integrity

**User Story:** As a Content_Manager, I want the full workflow of creating a TissueRecord, adding slides, and linking it to an Atlas to be consistent and error-free, so that I can populate an Atlas without data loss or broken states.

#### Acceptance Criteria

1. WHEN a Content_Manager completes the sequence — create TissueRecord → add at least one Slide → link to Atlas — THE Atlas_Detail_Page SHALL reflect the new TissueRecord with its slides immediately after the link is created.
2. IF a TissueRecord is deleted while it is linked to one or more Atlases, THEN THE System SHALL remove all corresponding rows from the `atlas_tissue_records` join table so that no Atlas references a non-existent TissueRecord.
3. IF a Slide is deleted from a TissueRecord that is linked to an Atlas, THEN THE Atlas_Detail_Page SHALL reflect the updated slide count and thumbnail for that TissueRecord on the next page load.
4. THE System SHALL preserve all existing TissueRecord data (name, notes, taxon, slides) when a TissueRecord is linked to or unlinked from an Atlas.
