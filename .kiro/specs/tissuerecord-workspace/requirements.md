# Requirements Document

## Introduction

The TissueRecord Workspace is a dedicated, single-page hub for managing every aspect of a TissueRecord in TissQuest. It replaces the cramped inline edit form in the record list and the limited detail page with a full workspace that lets a Content Manager configure basic info, slides, atlas associations, and category tags from one place — all without disrupting the rest of the page.

The workspace is reached by clicking "View" or "Edit" on any TissueRecord in the list. All four sections of the workspace — Basic Info, Slides, Atlas Associations, and Category Tags — can be updated independently without affecting one another.

## Glossary

- **Workspace**: The dedicated TissueRecord page that hosts all four management sections.
- **TissueRecord**: The central specimen entry in TissQuest, identified by a unique ID, with a name, optional notes, an optional taxon, slides, atlas associations, and category tags.
- **Basic_Info_Section**: The section of the Workspace that displays and allows inline editing of the TissueRecord's name, notes, and taxon.
- **Slide_Section**: The section of the Workspace that displays the slide gallery and allows adding and removing slides.
- **Atlas_Section**: The section of the Workspace that displays and manages the association between a TissueRecord and Atlases.
- **Category_Section**: The section of the Workspace that displays and manages the association between a TissueRecord and Categories.
- **Atlas**: A named collection entity that groups TissueRecords; already exists in the system.
- **Category**: A pre-existing classification entity with a name and type (e.g., organ, tissue, stain, custom) used to tag TissueRecords for search and filtering.
- **Content_Manager**: The human user who manages TissueRecord data in TissQuest.
- **System**: The TissQuest application responsible for reading and writing all data.

---

## Requirements

### Requirement 1: Workspace Page Entry Point

**User Story:** As a Content_Manager, I want to open a dedicated workspace page for a TissueRecord from the list, so that I can manage all aspects of the record from one place.

#### Acceptance Criteria

1. WHEN a Content_Manager clicks "View" or "Edit" for a TissueRecord in the tissue record list, THE System SHALL navigate to the Workspace for that TissueRecord.
2. THE Workspace SHALL display the TissueRecord's name as the page title.
3. THE Workspace SHALL include a breadcrumb trail: Home → Tissue Records → {record name}.
4. THE Workspace SHALL render all four sections — Basic_Info_Section, Slide_Section, Atlas_Section, and Category_Section — on initial load.
5. IF a TissueRecord with the requested ID does not exist, THEN THE System SHALL display a standard error page indicating the record was not found.

---

### Requirement 2: Basic Info Section — Display and Inline Edit

**User Story:** As a Content_Manager, I want to view and edit the name, notes, and taxon of a TissueRecord directly on the workspace page, so that I do not need to navigate to a separate form.

#### Acceptance Criteria

1. THE Basic_Info_Section SHALL display the TissueRecord's current name, notes, and taxon on initial load.
2. WHEN a Content_Manager activates edit mode for the Basic_Info_Section, THE Basic_Info_Section SHALL replace the display view with an inline form containing fields for name, notes, and taxon (selectable from existing taxa).
3. THE Basic_Info_Section SHALL activate edit mode without a full page reload, leaving the Slide_Section, Atlas_Section, and Category_Section undisturbed.
4. WHEN a Content_Manager submits the inline edit form with a non-empty name, THE System SHALL persist the updated name, notes, and taxon, and THE Basic_Info_Section SHALL return to display view showing the updated values.
5. IF a Content_Manager submits the inline edit form with an empty name, THEN THE Basic_Info_Section SHALL display a validation error message and retain the form with the entered values.
6. WHEN a Content_Manager cancels the inline edit, THE Basic_Info_Section SHALL return to display view showing the original values without persisting any changes.
7. THE Basic_Info_Section SHALL update independently so that the Slide_Section, Atlas_Section, and Category_Section are not affected.

---

### Requirement 3: Slide Section — Gallery with Add and Remove

**User Story:** As a Content_Manager, I want to add and remove slides for a TissueRecord from the workspace, so that I can manage the slide gallery without leaving the record hub.

#### Acceptance Criteria

1. THE Slide_Section SHALL display all slides associated with the TissueRecord, each showing the slide image (or a placeholder), name, magnification, and staining method.
2. WHEN a Content_Manager clicks "Add Slide", THE Slide_Section SHALL display an inline slide creation form without a full page reload.
3. WHEN a Content_Manager submits a valid slide creation form (non-empty name, positive magnification), THE System SHALL persist the new slide and THE Slide_Section SHALL refresh to include the new slide card.
4. IF a Content_Manager submits a slide creation form with an empty name or a non-positive magnification, THEN THE Slide_Section SHALL display field-level validation errors and retain the form with the entered values.
5. WHEN a Content_Manager confirms deletion of a slide, THE System SHALL remove the slide and THE Slide_Section SHALL refresh to reflect the removal.
6. THE Slide_Section SHALL update independently so that the Basic_Info_Section, Atlas_Section, and Category_Section are not affected.

---

### Requirement 4: Atlas Associations Section — Add and Remove

**User Story:** As a Content_Manager, I want to associate and disassociate Atlases with a TissueRecord from the workspace, so that I can control which atlases include this specimen.

#### Acceptance Criteria

1. THE Atlas_Section SHALL display all Atlases currently associated with the TissueRecord, each showing the atlas name.
2. THE Atlas_Section SHALL display an "Add Atlas" control that presents the list of Atlases not yet associated with the TissueRecord.
3. WHEN a Content_Manager selects an Atlas from the "Add Atlas" control, THE System SHALL create the association and THE Atlas_Section SHALL refresh to show the newly added atlas.
4. WHEN a Content_Manager clicks "Remove" on an associated Atlas, THE System SHALL delete the association and THE Atlas_Section SHALL refresh to reflect the removal.
5. IF all available Atlases are already associated with the TissueRecord, THEN THE Atlas_Section SHALL hide or disable the "Add Atlas" control.
6. THE Atlas_Section SHALL update independently so that the Basic_Info_Section, Slide_Section, and Category_Section are not affected.

---

### Requirement 5: Category Tags Section — Add and Remove

**User Story:** As a Content_Manager, I want to tag a TissueRecord with pre-existing Categories from the workspace, so that the record can be found and filtered by category in the future.

#### Acceptance Criteria

1. THE Category_Section SHALL display all Categories currently associated with the TissueRecord as visual tags, each showing the category name.
2. THE Category_Section SHALL display an "Add Category" control that presents the list of Categories not yet associated with the TissueRecord.
3. WHEN a Content_Manager selects a Category from the "Add Category" control, THE System SHALL create the association and THE Category_Section SHALL refresh to show the newly added category tag.
4. WHEN a Content_Manager clicks the remove control on a category tag, THE System SHALL delete the association and THE Category_Section SHALL refresh to reflect the removal.
5. IF all available Categories are already associated with the TissueRecord, THEN THE Category_Section SHALL hide or disable the "Add Category" control.
6. THE Category_Section SHALL update independently so that the Basic_Info_Section, Slide_Section, and Atlas_Section are not affected.

---

### Requirement 6: Persistence — Category Association

**User Story:** As a Content_Manager, I want category tag changes to be persisted reliably, so that tags survive page reloads and are available for future search and filtering.

#### Acceptance Criteria

1. WHEN a Category association is added to a TissueRecord, THE System SHALL record the relationship between the TissueRecord and the Category.
2. WHEN a Category association is removed from a TissueRecord, THE System SHALL remove the relationship between the TissueRecord and the Category.
3. IF a duplicate Category association is requested (the same Category is already associated with the TissueRecord), THEN THE System SHALL complete the operation without error and without creating a duplicate association.
4. WHEN a TissueRecord is loaded for the Workspace, THE System SHALL include all associated Categories.

---

### Requirement 7: Persistence — Atlas Association

**User Story:** As a Content_Manager, I want atlas association changes to be persisted reliably, so that atlas membership is consistent across the application.

#### Acceptance Criteria

1. WHEN an Atlas association is added to a TissueRecord, THE System SHALL record the relationship between the TissueRecord and the Atlas.
2. WHEN an Atlas association is removed from a TissueRecord, THE System SHALL remove the relationship between the TissueRecord and the Atlas.
3. IF a duplicate Atlas association is requested (the same Atlas is already associated with the TissueRecord), THEN THE System SHALL complete the operation without error and without creating a duplicate association.
4. WHEN a TissueRecord is loaded for the Workspace, THE System SHALL include all associated Atlases.

---

### Requirement 8: Navigation — List Page Update

**User Story:** As a Content_Manager, I want the "View" and "Edit" actions in the tissue record list to open the workspace, so that the workspace is the single entry point for record management.

#### Acceptance Criteria

1. THE tissue record list SHALL render a "View" link for each TissueRecord that navigates to the Workspace for that record.
2. THE tissue record list SHALL render an "Edit" link for each TissueRecord that navigates to the Workspace for that record, rather than triggering an inline row edit.
3. WHEN a Content_Manager navigates to the old TissueRecord detail page, THE System SHALL redirect to the Workspace for that record.
