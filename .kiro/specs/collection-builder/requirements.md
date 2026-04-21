# Requirements Document

## Introduction

The Collection Builder feature generalizes TissQuest's existing "Atlas" concept into a broader "Collection" abstraction. A Collection is a named, curated grouping of tissue records with optional internal structure (sections and subsections). This enables curators to build atlases, scientific databases, personal reference sets, and other organized groupings from the shared pool of tissue records.

The feature introduces two distinct user roles: the **Contributor**, who adds tissue records to the global pool, and the **Curator**, who builds structured collections from existing records. The Collection Builder is a single-screen interface where a curator can manage collection metadata, create and reorder sections, and assign tissue records to sections.

The existing `Atlas` domain concept is replaced by `Collection` at the domain level. The underlying database table (`atlases`) is preserved initially to avoid breaking migrations.

## Glossary

- **Collection**: A named, curated grouping of tissue records with optional internal structure. Generalizes the former "Atlas" concept. Can represent an atlas, a scientific database, a personal reference set, or any other organized grouping.
- **Collection_Builder**: The single-screen interface used by a Curator to create and manage a Collection, its Sections, and the assignment of Tissue Records to those Sections.
- **Section**: A named, ordered subdivision of a Collection. Sections are owned by exactly one Collection and are not reusable across Collections. Sections may contain Subsections, forming a two-level hierarchy.
- **Subsection**: A Section nested inside another Section. Subsections follow the same ordering rules as top-level Sections within their parent.
- **Tissue_Record**: An independently existing record representing a tissue sample, including its scientific name, taxonomic classification, notes, and associated slides. A Tissue Record may belong to zero or many Collections and may appear in multiple Sections across multiple Collections.
- **Section_Assignment**: The association between a Tissue Record and a Section, including an explicit display order within that Section.
- **Curator**: A user role responsible for creating and managing Collections, Sections, and Section Assignments.
- **Contributor**: A user role responsible for adding Tissue Records to the global pool.
- **Collection_Type**: A classification for a Collection indicating its intended purpose. Valid values are: `atlas`, `database`, `reference`, `other`.
- **Collection_Metadata**: The descriptive attributes of a Collection: name, description, goals, type, and authors (free-text).

---

## Requirements

### Requirement 1: Collection Metadata Management

**User Story:** As a Curator, I want to create and edit a Collection's metadata, so that I can describe the purpose, scope, and authorship of the Collection.

#### Acceptance Criteria

1. THE Collection_Builder SHALL provide input fields for the Collection's name, description, goals, type, and authors.
2. WHEN a Curator submits a new Collection with an empty name, THE Collection_Builder SHALL reject the submission and display a validation error message.
3. WHEN a Curator submits a Collection name longer than 200 characters, THE Collection_Builder SHALL reject the submission and display a validation error message.
4. WHEN a Curator selects a Collection type, THE Collection_Builder SHALL restrict valid values to `atlas`, `database`, `reference`, and `other`.
5. WHEN a Curator saves valid Collection metadata, THE Collection_Builder SHALL persist the Collection and display a confirmation to the Curator.
6. WHEN a Curator edits an existing Collection's metadata and saves, THE Collection_Builder SHALL update the persisted Collection and display a confirmation.

---

### Requirement 2: Section Management

**User Story:** As a Curator, I want to create, rename, reorder, and delete Sections within a Collection, so that I can give the Collection a meaningful internal structure.

#### Acceptance Criteria

1. WHEN a Curator creates a Section, THE Collection_Builder SHALL associate the Section with the current Collection and assign it the next available display order position.
2. WHEN a Curator submits a Section with an empty name, THE Collection_Builder SHALL reject the submission and display a validation error message.
3. WHEN a Curator reorders Sections, THE Collection_Builder SHALL persist the new display order for all affected Sections within the Collection.
4. WHEN a Curator deletes a Section that contains Section Assignments, THE Collection_Builder SHALL remove all Section Assignments belonging to that Section before deleting the Section.
5. WHEN a Curator creates a Subsection under an existing Section, THE Collection_Builder SHALL associate the Subsection with the parent Section and assign it the next available display order position within that parent.
6. THE Collection_Builder SHALL support a maximum nesting depth of two levels (Section → Subsection).
7. WHEN a Curator reorders Subsections within a parent Section, THE Collection_Builder SHALL persist the new display order for all affected Subsections within that parent Section.

---

### Requirement 3: Tissue Record Assignment to Sections

**User Story:** As a Curator, I want to assign existing Tissue Records to Sections within a Collection, so that I can populate the Collection with relevant content in a deliberate order.

#### Acceptance Criteria

1. WHEN a Curator assigns a Tissue Record to a Section, THE Collection_Builder SHALL create a Section Assignment with an explicit display order position at the end of the Section's current assignments.
2. WHEN a Curator assigns a Tissue Record that is already assigned to the same Section, THE Collection_Builder SHALL reject the duplicate assignment and display an informational message.
3. WHEN a Curator reorders Section Assignments within a Section, THE Collection_Builder SHALL persist the new display order for all affected Section Assignments.
4. WHEN a Curator removes a Tissue Record from a Section, THE Collection_Builder SHALL delete the Section Assignment and resequence the remaining display order positions within that Section.
5. THE Collection_Builder SHALL allow the same Tissue Record to be assigned to multiple Sections within the same Collection.
6. THE Collection_Builder SHALL allow the same Tissue Record to be assigned to Sections in different Collections without restriction.

---

### Requirement 4: Tissue Record Search and Selection

**User Story:** As a Curator, I want to search for existing Tissue Records from the global pool and add them to a Section, so that I can build a Collection without leaving the Collection Builder screen.

#### Acceptance Criteria

1. WHEN a Curator initiates a Tissue Record search, THE Collection_Builder SHALL query the global Tissue Record pool and respond within a reasonable time for typical dataset sizes.
2. WHEN a Curator enters a search term, THE Collection_Builder SHALL filter Tissue Records by name and taxonomic classification using a case-insensitive substring match.
3. WHEN a search returns no results, THE Collection_Builder SHALL display a message indicating no matching Tissue Records were found.
4. WHEN a Curator selects a Tissue Record from search results, THE Collection_Builder SHALL add it to the currently active Section as a new Section Assignment.

---

### Requirement 5: Inline Tissue Record Creation

**User Story:** As a Curator, I want to create a new Tissue Record directly from the Collection Builder, so that I can add content that does not yet exist in the global pool without navigating away.

#### Acceptance Criteria

1. WHEN a Curator initiates inline Tissue Record creation, THE Collection_Builder SHALL display a modal form with fields for name, notes, and taxonomic classification.
2. WHEN a Curator submits the inline form with an empty name, THE Collection_Builder SHALL reject the submission and display a validation error within the modal.
3. WHEN a Curator successfully submits the inline form, THE Collection_Builder SHALL persist the new Tissue Record to the global pool and automatically create a Section Assignment in the currently active Section.
4. WHEN the inline Tissue Record creation is cancelled, THE Collection_Builder SHALL close the modal without persisting any data.

---

### Requirement 6: Collection Persistence and Data Integrity

**User Story:** As a Curator, I want the Collection and all its structure to be reliably persisted, so that my work is not lost between sessions.

#### Acceptance Criteria

1. THE Collection_Builder SHALL persist Collections, Sections, Subsections, and Section Assignments in the TissQuest database.
2. WHEN a Collection is deleted, THE Collection_Builder SHALL cascade-delete all Sections, Subsections, and Section Assignments belonging to that Collection.
3. WHEN a Tissue Record is deleted from the global pool, THE Collection_Builder SHALL remove all Section Assignments referencing that Tissue Record across all Collections.
4. THE Collection_Builder SHALL preserve the display order of Sections and Section Assignments across page reloads and concurrent sessions.
5. WHEN a Curator saves any structural change (section creation, reorder, assignment), THE Collection_Builder SHALL confirm the save with a non-blocking flash notification.

---

### Requirement 7: Collection Listing and Navigation

**User Story:** As a Curator, I want to view a list of all Collections and navigate to any Collection's builder screen, so that I can manage my work across multiple Collections.

#### Acceptance Criteria

1. THE Collection_Builder SHALL provide a Collections list page that displays all Collections with their name, type, and creation date.
2. WHEN a Curator selects a Collection from the list, THE Collection_Builder SHALL navigate to the Collection Builder screen for that Collection.
3. WHEN no Collections exist, THE Collection_Builder SHALL display a message prompting the Curator to create the first Collection.
4. THE Collection_Builder SHALL provide a breadcrumb navigation trail on the Collection Builder screen showing: Home → Collections → [Collection Name].

---

### Requirement 8: Backward Compatibility with Existing Atlas Data

**User Story:** As a developer, I want the Collection domain model to reuse the existing `atlases` database table, so that existing data is preserved without a destructive migration.

#### Acceptance Criteria

1. THE Collection_Builder SHALL map the Collection domain entity to the existing `atlases` database table.
2. WHEN the application starts, THE Collection_Builder SHALL apply additive-only schema migrations (new columns, new tables) without dropping or renaming existing columns in the `atlases` table.
3. THE Collection_Builder SHALL treat existing Atlas records as Collections of type `atlas` by default.

---

### Requirement 9: Author Management (Deferred)

**User Story:** As a Curator, I want to record authorship information on Collections and Tissue Records, so that credit and provenance are captured.

#### Acceptance Criteria

1. FOR MVP, the `authors` field on both Collections and Tissue Records SHALL be stored as a free-text string field with no structured validation.
2. Full author management (structured author entities, linking, search by author) is deferred to a future spec.
