# Requirements Document

## Introduction

This feature adds full Create, Read, Update, and Delete (CRUD) capabilities to the TissQuest web interface for all primary domain objects: Atlas, TissueRecord, Slide, Taxon, and Category. The current UI is read-only. This feature extends it with server-side HTML forms rendered via Gin and Go templates, using HTMX for all dynamic interactions. The interaction model follows HATEOAS principles: the server returns HTML fragments that embed the next available actions as `hx-*` attributes, driving all state transitions without a JavaScript framework. No authentication is required (public educational content).

## Glossary

- **CRUD_UI**: The web-based interface providing create, read, update, and delete operations for TissQuest domain objects via HTML forms, server-side rendering, and HTMX-driven partial updates.
- **Atlas**: A curated collection of TissueRecords grouped for educational purposes.
- **TissueRecord**: A digital record of an individual biological tissue specimen, including scientific name, taxonomic classification, notes, and associated Slides.
- **Slide**: A digital representation of a prepared microscopy slide, including image URL, magnification, and preparation metadata.
- **Taxon**: A node in the taxonomic classification hierarchy (kingdom → species), with an optional parent Taxon.
- **Category**: A hierarchical grouping entity used to organize TissueRecords by organ, species, tissue type, staining method, or custom label.
- **Handler**: A Gin HTTP handler function that processes a request and renders an HTML template response — either a full page layout or a partial HTML fragment, depending on the presence of the `HX-Request` header.
- **Form**: An HTML form submitted via HTMX (`hx-post` or `hx-put`) that carries user-supplied field values for creating or updating a domain object.
- **Fragment**: A partial HTML snippet returned by the Handler when the `HX-Request` header is present, intended to be swapped into the DOM by HTMX without a full page reload.
- **Flash_Message**: A one-time status message (success or error) delivered to the client via the `HX-Trigger` response header or an out-of-band swap (`hx-swap-oob`), displayed without a full page reload.
- **Confirmation_Snippet**: An inline HTML fragment swapped in place of a delete button that asks the user to confirm or cancel a destructive action, rendered without a page navigation.
- **OOB_Swap**: An out-of-band HTMX swap (`hx-swap-oob="true"`) that updates a secondary DOM target (such as a flash message region or list) alongside the primary swap target in a single server response.
- **HX-Request**: An HTTP request header set by HTMX on every request it initiates, used by the Handler to distinguish HTMX partial requests from full browser navigations.
- **HX-Trigger**: An HTTP response header used by the Handler to emit named browser events that HTMX listeners can react to, used here to deliver Flash_Messages.
- **Slide_Gallery**: The section of the TissueRecord detail page that displays all associated Slides and allows adding or removing Slides without leaving the page.

---

## Requirements

### Requirement 1: Atlas CRUD

**User Story:** As a content manager, I want to create, edit, and delete atlases via the web interface, so that I can maintain the collection of tissue atlases without direct database access.

#### Acceptance Criteria

1. THE CRUD_UI SHALL provide a page listing all atlases, where each row includes `hx-get` links to load the edit form inline and a delete trigger that loads the Confirmation_Snippet in place.
2. THE CRUD_UI SHALL provide a form for creating a new Atlas with fields: name (required, max 100 characters), description, and category; the form SHALL use `hx-post` to submit without a full page reload.
3. WHEN a valid atlas creation Form is submitted via HTMX, THE Handler SHALL persist the new Atlas, return an updated atlas list Fragment via OOB_Swap, and emit a success Flash_Message via the `HX-Trigger` response header.
4. IF the atlas creation Form contains an empty name field, THEN THE Handler SHALL return the form Fragment with an inline validation error message identifying the empty name field, targeting the form's container via `hx-target`.
5. IF the atlas creation Form contains a name exceeding 100 characters, THEN THE Handler SHALL return the form Fragment with an inline validation error message identifying the name length violation.
6. THE CRUD_UI SHALL load the pre-populated atlas edit form as a Fragment swapped inline via `hx-get` and `hx-swap`, without navigating away from the list page.
7. WHEN a valid atlas edit Form is submitted via HTMX, THE Handler SHALL update the Atlas record, return the updated atlas row Fragment, and emit a success Flash_Message via the `HX-Trigger` response header.
8. IF an atlas edit Form is submitted with invalid data, THEN THE Handler SHALL return the edit form Fragment with inline validation error messages, leaving the rest of the page unchanged.
9. WHEN a delete action is confirmed via the Confirmation_Snippet, THE Handler SHALL delete the Atlas record, return an empty Fragment removing the deleted row, and emit a success Flash_Message via the `HX-Trigger` response header.
10. IF a delete action targets a non-existent Atlas ID, THEN THE Handler SHALL return an HTTP 404 Fragment with a user-readable error message suitable for inline display.

---

### Requirement 2: TissueRecord CRUD

**User Story:** As a content manager, I want to create, edit, and delete tissue records via the web interface, so that I can manage individual specimen data including taxonomic classification and notes.

#### Acceptance Criteria

1. THE CRUD_UI SHALL provide a page listing all TissueRecords with pagination of 20 records per page, where each row includes `hx-get` links to load the edit form inline and a delete trigger that loads the Confirmation_Snippet in place.
2. THE CRUD_UI SHALL provide a form for creating a new TissueRecord with fields: name (required), notes, and an optional Taxon selection from existing taxa; the form SHALL use `hx-post` to submit without a full page reload.
3. WHEN a valid TissueRecord creation Form is submitted via HTMX, THE Handler SHALL persist the new TissueRecord, return an updated list Fragment via OOB_Swap, and emit a success Flash_Message via the `HX-Trigger` response header.
4. IF the TissueRecord creation Form contains an empty name field, THEN THE Handler SHALL return the form Fragment with an inline validation error message identifying the empty name field.
5. THE CRUD_UI SHALL load the pre-populated TissueRecord edit form as a Fragment swapped inline via `hx-get` and `hx-swap`, without navigating away from the list page.
6. WHEN a valid TissueRecord edit Form is submitted via HTMX, THE Handler SHALL update the TissueRecord, return the updated record row Fragment, and emit a success Flash_Message via the `HX-Trigger` response header.
7. WHEN a delete action is confirmed via the Confirmation_Snippet, THE Handler SHALL delete the TissueRecord, return an empty Fragment removing the deleted row, and emit a success Flash_Message via the `HX-Trigger` response header.
8. IF a delete action targets a non-existent TissueRecord ID, THEN THE Handler SHALL return an HTTP 404 Fragment with a user-readable error message suitable for inline display.

---

### Requirement 3: Slide CRUD (Slide Gallery)

**User Story:** As a content manager, I want to add and remove slides within a tissue record without leaving the page, so that I can associate microscopy images and their preparation metadata with specimens efficiently.

#### Acceptance Criteria

1. THE CRUD_UI SHALL display the Slide_Gallery on the TissueRecord detail page as a Fragment region identified by a stable DOM ID, showing all associated Slides with inline delete triggers.
2. THE CRUD_UI SHALL provide an "Add Slide" form within the Slide_Gallery with fields: name (required), image URL, magnification (integer, required), staining, inclusion method, reagents, protocol, and notes; the form SHALL use `hx-post` to submit without leaving the TissueRecord detail page.
3. WHEN a valid Slide creation Form is submitted via HTMX, THE Handler SHALL persist the new Slide linked to the parent TissueRecord, return the updated Slide_Gallery Fragment via `hx-target` and `hx-swap`, and emit a success Flash_Message via the `HX-Trigger` response header.
4. IF the Slide creation Form contains an empty name field or a non-integer magnification value, THEN THE Handler SHALL return the form Fragment with inline validation error messages identifying each invalid field, leaving the Slide_Gallery list unchanged.
5. THE CRUD_UI SHALL load the pre-populated Slide edit form as a Fragment swapped inline within the Slide_Gallery via `hx-get` and `hx-swap`, without navigating away from the TissueRecord detail page.
6. WHEN a valid Slide edit Form is submitted via HTMX, THE Handler SHALL update the Slide record, return the updated Slide_Gallery Fragment, and emit a success Flash_Message via the `HX-Trigger` response header.
7. WHEN a delete action is confirmed via the Confirmation_Snippet within the Slide_Gallery, THE Handler SHALL delete the Slide, return the updated Slide_Gallery Fragment with the deleted Slide removed, and emit a success Flash_Message via the `HX-Trigger` response header.
8. IF a delete action targets a non-existent Slide ID, THEN THE Handler SHALL return an HTTP 404 Fragment with a user-readable error message suitable for inline display within the Slide_Gallery.

---

### Requirement 4: Taxon CRUD

**User Story:** As a content manager, I want to create, edit, and delete taxa via the web interface, so that I can maintain the taxonomic classification hierarchy used to classify tissue records.

#### Acceptance Criteria

1. THE CRUD_UI SHALL provide a page listing all Taxa grouped by rank, where each row includes `hx-get` links to load the edit form inline and a delete trigger that loads the Confirmation_Snippet in place.
2. THE CRUD_UI SHALL provide a form for creating a new Taxon with fields: name (required), rank (required, one of: kingdom, phylum, class, order, family, genus, species), and an optional parent Taxon selection from existing taxa; the form SHALL use `hx-post` to submit without a full page reload.
3. WHEN a valid Taxon creation Form is submitted via HTMX, THE Handler SHALL persist the new Taxon, return an updated taxon list Fragment via OOB_Swap, and emit a success Flash_Message via the `HX-Trigger` response header.
4. IF the Taxon creation Form contains an empty name field or an invalid rank value, THEN THE Handler SHALL return the form Fragment with inline validation error messages identifying each invalid field.
5. THE CRUD_UI SHALL load the pre-populated Taxon edit form as a Fragment swapped inline via `hx-get` and `hx-swap`, without navigating away from the list page.
6. WHEN a valid Taxon edit Form is submitted via HTMX, THE Handler SHALL update the Taxon record, return the updated taxon row Fragment, and emit a success Flash_Message via the `HX-Trigger` response header.
7. WHEN a delete action is confirmed via the Confirmation_Snippet, THE Handler SHALL delete the Taxon, return an empty Fragment removing the deleted row, and emit a success Flash_Message via the `HX-Trigger` response header.
8. IF a delete action targets a non-existent Taxon ID, THEN THE Handler SHALL return an HTTP 404 Fragment with a user-readable error message suitable for inline display.

---

### Requirement 5: Category CRUD

**User Story:** As a content manager, I want to create, edit, and delete categories via the web interface, so that I can maintain the hierarchical groupings used to organize tissue records.

#### Acceptance Criteria

1. THE CRUD_UI SHALL provide a page listing all Categories with their type and optional parent, where each row includes `hx-get` links to load the edit form inline and a delete trigger that loads the Confirmation_Snippet in place.
2. THE CRUD_UI SHALL provide a form for creating a new Category with fields: name (required), type (required, one of: organ, species, tissue, stain, custom), description, and an optional parent Category selection from existing categories; the form SHALL use `hx-post` to submit without a full page reload.
3. WHEN a valid Category creation Form is submitted via HTMX, THE Handler SHALL persist the new Category, return an updated category list Fragment via OOB_Swap, and emit a success Flash_Message via the `HX-Trigger` response header.
4. IF the Category creation Form contains an empty name field or an invalid type value, THEN THE Handler SHALL return the form Fragment with inline validation error messages identifying each invalid field.
5. IF the Category creation Form specifies a ParentID equal to the Category's own ID, THEN THE Handler SHALL return the form Fragment with a validation error message indicating a circular parent reference is not allowed.
6. THE CRUD_UI SHALL load the pre-populated Category edit form as a Fragment swapped inline via `hx-get` and `hx-swap`, without navigating away from the list page.
7. WHEN a valid Category edit Form is submitted via HTMX, THE Handler SHALL update the Category record, return the updated category row Fragment, and emit a success Flash_Message via the `HX-Trigger` response header.
8. WHEN a delete action is confirmed via the Confirmation_Snippet, THE Handler SHALL delete the Category, return an empty Fragment removing the deleted row, and emit a success Flash_Message via the `HX-Trigger` response header.
9. IF a delete action targets a non-existent Category ID, THEN THE Handler SHALL return an HTTP 404 Fragment with a user-readable error message suitable for inline display.

---

### Requirement 6: Delete Confirmation

**User Story:** As a content manager, I want to be prompted for confirmation before any delete action is executed, so that I do not accidentally destroy data.

#### Acceptance Criteria

1. WHEN the user activates a delete trigger on any Atlas, TissueRecord, Slide, Taxon, or Category row, THE CRUD_UI SHALL use `hx-get` to load a Confirmation_Snippet that swaps in place of the delete trigger via `hx-swap="outerHTML"`.
2. THE Confirmation_Snippet SHALL embed a confirm button with `hx-delete` pointing to the resource URL and a cancel button with `hx-get` that restores the original delete trigger Fragment.
3. WHEN the user activates the cancel button in the Confirmation_Snippet, THE Handler SHALL return the original delete trigger Fragment, restoring the row to its pre-confirmation state without a page reload.
4. WHEN the user activates the confirm button in the Confirmation_Snippet, THE Handler SHALL execute the delete and return the appropriate Fragment as specified in the relevant entity CRUD requirement.

---

### Requirement 7: Flash Messages and User Feedback

**User Story:** As a content manager, I want to see clear success and error messages after form submissions, so that I know whether my actions succeeded or failed.

#### Acceptance Criteria

1. WHEN a create, update, or delete operation succeeds, THE Handler SHALL emit a success Flash_Message by setting the `HX-Trigger` response header to a named event (e.g., `showFlash`) carrying the message text.
2. WHEN a create or update operation fails due to a server error, THE Handler SHALL return the form Fragment with an error message embedded directly in the Fragment, visible without a page reload.
3. THE CRUD_UI SHALL include a dedicated flash message region identified by a stable DOM ID that listens for the `showFlash` event via `hx-on` or an equivalent HTMX event binding and renders the message text.
4. THE CRUD_UI SHALL display each Flash_Message only once; the flash message region SHALL clear itself after the message is displayed so that a subsequent HTMX request that does not set `HX-Trigger` does not re-display the previous message.
5. WHERE the Handler returns multiple Fragments in a single response (e.g., updated list and flash message), THE Handler SHALL use OOB_Swap to update the flash message region alongside the primary swap target.

---

### Requirement 8: Navigation

**User Story:** As a content manager, I want consistent navigation links throughout the CRUD UI, so that I can move between entity lists, detail pages, and forms without using the browser back button.

#### Acceptance Criteria

1. THE CRUD_UI SHALL include a navigation link to each entity list page (Atlases, TissueRecords, Taxa, Categories) in the main menu; these links SHALL use standard `<a href>` elements to perform full page navigations.
2. THE CRUD_UI SHALL include a breadcrumb trail on every full page and detail page showing the path from the home page to the current page.
3. THE CRUD_UI SHALL include a "Cancel" link on every inline create and edit form Fragment that uses `hx-get` to restore the previous DOM state (e.g., the empty form placeholder or the original row) without submitting the form and without a full page reload.

---

### Requirement 9: HTMX/HATEOAS Interaction Model

**User Story:** As a developer, I want the CRUD UI to follow HTMX and HATEOAS principles, so that all state transitions are driven by server-returned HTML fragments and I can learn HTMX patterns through this project.

#### Acceptance Criteria

1. THE CRUD_UI SHALL load HTMX from the official CDN (`https://unpkg.com/htmx.org`) via a `<script>` tag in the base layout, with no other JavaScript framework required.
2. WHEN the Handler receives a request with the `HX-Request: true` header, THE Handler SHALL return a partial HTML Fragment without the full page layout (no `<html>`, `<head>`, or `<body>` wrapper).
3. WHEN the Handler receives a request without the `HX-Request` header, THE Handler SHALL return a complete HTML page including the full layout, so that direct browser navigation and bookmarking work correctly.
4. THE CRUD_UI SHALL embed all next available actions as `hx-*` attributes directly within each returned Fragment, so that the client never needs to construct URLs or determine available actions independently (HATEOAS constraint).
5. THE CRUD_UI SHALL use `hx-target` and `hx-swap` attributes on all interactive elements to specify exactly which DOM region is replaced and by which swap strategy (`innerHTML`, `outerHTML`, `afterbegin`, etc.), with no implicit full-page replacements.
6. THE CRUD_UI SHALL use `hx-swap-oob="true"` on secondary Fragment regions within a single response when multiple DOM regions must be updated simultaneously (e.g., refreshing a list while also updating the flash message region).
7. WHEN a Form submission results in a validation error, THE Handler SHALL return HTTP 422 (Unprocessable Entity) with the form Fragment so that HTMX re-renders only the form region and does not update the list or other page regions.
8. WHEN a Form submission or delete operation succeeds, THE Handler SHALL return HTTP 200 with the replacement Fragment and set the `HX-Trigger` response header to deliver the Flash_Message as a browser event rather than via a cookie or session variable.
9. THE CRUD_UI SHALL set `hx-push-url` on top-level page navigations (list pages, detail pages) so that the browser URL bar and history remain consistent with the displayed content.
10. IF a Handler returns an error Fragment (HTTP 4xx or 5xx), THEN THE CRUD_UI SHALL display the error message within the relevant page region without replacing the entire page, preserving the user's current navigation context.
