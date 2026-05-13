# Requirements Specification for Tissquest

## Overview
Tissquest is a personal project aimed at creating an educational platform for studying microscopical images of biological tissues. The initial focus is on plant anatomy, with plans to expand to other tissue types in the future. The system will allow users to browse, view, and learn from high-quality microscopy images organized into atlases and tissue records.

**Project Timeline**: Deliver functional system before July 2026.

**Inspiration and Learning Resources**:
- [Requirements Engineering University](https://requirements.university/) - For learning requirements engineering principles
- [High-Level Requirements by Leslie Lamport](https://lamport.azurewebsites.net/pubs/high-level.pdf) - Focus on high-level, abstract requirements rather than implementation details

**Terminology**: See [Requirements Glossary](REQUIREMENTS-GLOSSARY.md) for definitions of key terms used in this specification.

## High-Level Requirements (Following Lamport's Approach)

### Goal
Provide an accessible, web-based platform for students and enthusiasts to study plant tissue microscopy without requiring expensive equipment or physical specimens.

### Key Properties
1. **Educational Focus**: Content should be organized for learning, with clear taxonomic classifications and descriptive notes.
2. **Image Quality**: High-resolution microscopy images suitable for detailed study.
3. **Accessibility**: Web-based interface that works on standard devices.
4. **Scalability**: Architecture that can accommodate expansion to other tissue types.
5. **Maintainability**: Clean code structure for personal development and future contributions.

## Functional Requirements

### Core Features (Plant Anatomy Focus)
1. **Atlas Management**
   - Create and organize collections of plant tissue samples
   - Categorize atlases by plant type, tissue type, or educational level
   - View atlas details including:
     - Atlas metadata (name, description, category, creation date)
     - List of associated tissue records organized by taxonomic categories
     - Thumbnail images or preview of key slides
     - Navigation to individual tissue record pages

2. **Tissue Record Management**
   - Store individual tissue samples with:
     - Scientific name
     - Taxonomic classification
     - Descriptive notes
     - Associated microscopy slides
   - CRUD operations for tissue records

3. **Slide Management**
   - Store microscopy images with metadata:
     - Magnification level
     - Staining technique used
     - Image URL or file reference
   - Display slides within tissue record context

4. **User Interface**
   - Web-based browsing interface
   - Responsive design for different screen sizes
   - Image gallery views
   - Search and filtering capabilities
   - **Atlas Detail Page**:
     - Header with atlas title, description, and category
     - Organized display of tissue records by taxonomic hierarchy (Species → Organ → Tissue Type)
     - Thumbnail images for each tissue record
     - Clickable links to detailed tissue record views
     - Breadcrumb navigation (Home > Atlases > [Atlas Name])
     - Educational notes and context for the atlas

5. **Library Desk Home Page**
   - The entry point of the platform must function as a discovery interface, not a marketing page
   - A prominent search bar allows students to find tissue records by name, taxon, or staining technique
   - Two primary navigation paths are presented with equal weight: Collections and Tissue Browser
   - A curated selection of randomly chosen tiled slides invites students to open the Virtual Microscope directly from the home page
   - The home page must refresh its slide selection on each visit to surface different specimens over time

6. **Image Tiling Pipeline**
   - Content authors must be able to prepare a slide for interactive viewing without leaving the platform
   - A single-slide tiling action must be available from the slide management interface
   - A batch tiling operation must be available as a command-line tool for processing multiple slides at once
   - The batch operation must automatically identify all slides that have a source image but have not yet been prepared for interactive viewing
   - Both operations must update the slide record upon completion so the interactive viewer becomes available immediately

### API Requirements
- RESTful API for data operations
- JSON responses for programmatic access
- Pagination for large datasets

## Non-Functional Requirements

### Performance
- Page load times under 2 seconds for typical use cases
- Support for concurrent users (target: 10-50 simultaneous users)

### Usability
- Intuitive navigation for non-technical users
- Clear labeling of scientific terms
- Educational context provided with images

### Reliability
- Data persistence with backup capabilities
- Error handling with user-friendly messages
- Basic validation of input data

### Security
- Basic input sanitization
- No user authentication required initially (public educational content)

## Use Cases

### Primary User Scenarios
1. **Student Learning**
   - Arrive at the home page and search for a tissue by name or taxon
   - Discover specimens through the randomly featured slides on the home page
   - Browse plant tissue collections
   - Open the Virtual Microscope directly from the home page or from a tissue record
   - View detailed microscopy images at different magnification levels
   - Read educational notes and classifications
   - Compare different tissue types
   - **Navigate atlas structure**: Explore tissue records organized by taxonomic categories within an atlas

2. **Content Management**
   - Add new tissue samples and images
   - Prepare slides for interactive viewing using the tiling action on the slide card or the batch pipeline tool
   - Organize content into logical atlases
   - Update metadata as needed

## Technical Constraints
- Built with Go and web technologies
- SQLite database for simplicity
- Docker containerization for deployment
- Personal development project (single developer)

## Virtual Microscope Viewer

> For technology-specific decisions, tooling, and implementation guidance see [Virtual Microscope — Technical Design](VIRTUAL-MICROSCOPE-TECH.md).

### Goal
Replace the static image gallery with an interactive viewer that emulates the experience of operating a real light microscope — allowing a student to navigate a tissue specimen at different magnification levels, measure structures, and return to a curated starting position, all from a web browser without any installed software.

### Key Properties
1. **Progressive Detail**: The viewer must reveal increasing levels of detail as the user zooms in, without requiring the full high-resolution image to be downloaded upfront.
2. **Microscope Analogy**: Navigation gestures and controls must map intuitively to the physical actions a student already knows from a laboratory microscope (focus knob, objective switcher, stage movement).
3. **Calibrated Measurement**: The viewer must always display a scale indicator that reflects real physical dimensions of the specimen, so students can estimate the size of structures they observe.
4. **Contextual Entry Point**: Each slide may define a curated starting view that highlights the most educationally relevant region of the specimen.
5. **Graceful Degradation**: Slides that have not been processed for interactive viewing must still be accessible through the existing static image display.
6. **Integration**: The viewer is not a standalone feature — it is the primary way a [Slide](REQUIREMENTS-GLOSSARY.md#slide) is presented within a [Tissue Record](REQUIREMENTS-GLOSSARY.md#tissue-record).

### Functional Requirements

#### FR-VM-1: Image Preparation Pipeline
- High-resolution source images must be processable into a [tiled image format](REQUIREMENTS-GLOSSARY.md#tiled-image-format) through an automated, repeatable pipeline.
- The pipeline must be triggerable without manual image editing steps.
- Processed tile sets must be stored in a remote object store accessible by the web application.
- The pipeline must associate the resulting tile set with the corresponding Slide record in the database.

#### FR-VM-2: Slide Metadata Extension
- The [Slide](REQUIREMENTS-GLOSSARY.md#slide) entity must be extended to carry:
  - A reference to its [tiled image](REQUIREMENTS-GLOSSARY.md#tiled-image-format) resource.
  - The [base magnification](REQUIREMENTS-GLOSSARY.md#base-magnification) at which the specimen was captured.
  - The [spatial calibration](REQUIREMENTS-GLOSSARY.md#spatial-calibration) value (physical size per image pixel) required for accurate measurement display.
- Slides without a tiled image reference must remain fully functional.

#### FR-VM-3: Viewer Metadata API
- The system must expose an endpoint that returns all viewer-initialization data for a given Slide in a single request.
- The response must include the tiled image location, base magnification, and spatial calibration value.
- This endpoint feeds the viewer exclusively; it does not replace the existing slide detail API.

#### FR-VM-4: Interactive Viewer
- The Slide detail page must embed an interactive viewer capable of smooth, continuous zoom and pan over the tiled image.
- The viewer must feel responsive — tile loading must not freeze or stutter the interface.
- The viewer must occupy the primary content area of the page on all supported screen sizes.

#### FR-VM-5: Objective Lens Switcher
- The viewer must provide discrete zoom controls labeled with standard microscopy objective values (e.g., 4×, 10×, 40×).
- Activating a control must animate the viewer to the corresponding zoom level relative to the slide's base magnification.
- The currently active objective must be visually distinguishable from the others.
- Controls must be operable by keyboard as well as pointer.

#### FR-VM-6: Scale Indicator
- The viewer must display a dynamic scale bar that shows a real-world length (in micrometers, µm) at the current zoom level.
- The scale bar must update continuously as the user zooms.
- The displayed measurement must be derived from the slide's [spatial calibration](REQUIREMENTS-GLOSSARY.md#spatial-calibration) value.

#### FR-VM-7: Curated Home View
- A Slide may store a designated starting viewport (position and zoom level) chosen by the content author.
- When a student opens a slide with a home view defined, the viewer must open at that position.
- Content authors must be able to set the home view to the current viewport position through the management interface.

#### FR-VM-8: Touch and Mobile Support
- All viewer interactions (zoom, pan, objective switching) must be fully operable on touch-screen devices.
- The layout must not break or overflow on small screens.

#### FR-VM-9: Static Image Fallback
- When a Slide has no tiled image, the viewer area must display the slide's static image without errors or blank space.

### Non-Functional Requirements

#### NFR-VM-1: Initial Load Performance
- The viewer must display the first visible region of the specimen within 2 seconds on a standard broadband connection, regardless of the full image resolution.

#### NFR-VM-2: Stability Under Adverse Conditions
- Intermittent network failures during tile loading must not crash or freeze the viewer.
- Memory consumption must remain bounded during extended browsing sessions.

#### NFR-VM-3: Accessibility
- The viewer area must carry a descriptive label readable by assistive technologies.
- Keyboard-only users must be able to zoom and pan without a pointing device.

### Acceptance Criteria
- At least 20 plant tissue slides are accessible through the interactive viewer.
- The objective lens switcher navigates to the correct zoom level for each objective.
- The scale bar displays accurate µm values, verified against a reference slide with known dimensions.
- The viewer reaches first paint within 2 seconds on a standard broadband connection.
- Zoom and pan work correctly on a physical touch-screen device.
- Slides without a tiled image display the static fallback without JavaScript errors.

---

## TissExplorer — Tissue Search and Discovery

> The TissExplorer is the primary discovery interface of the platform. It allows any student to find tissue records through free-text search, hierarchical category filters, or a combination of both — without needing to know the exact name of a specimen in advance.
> The interaction model is inspired by faceted search in scientific databases (e.g., RCSB PDB) and product catalogues: the user narrows a large collection down to a relevant subset through progressive filtering, while the results update to reflect the current selection at all times.

### Goal
Give students and researchers a single, unified entry point to discover tissue records across the entire library — by name, taxon, staining technique, organ system, or any combination of categories — without requiring prior knowledge of the catalogue structure.

### Key Properties
1. **Unified Entry Point**: A single search bar handles all discovery. Students do not need to know which collection a specimen belongs to before finding it.
2. **Faceted Filtering**: Category filters narrow results progressively. Filters from different dimensions (e.g., taxon + staining + organ) can be combined freely.
3. **Hierarchical Organisation**: Categories are presented as a navigable tree, reflecting the natural hierarchy of biological classification and anatomical systems.
4. **Visual-First Results**: Results are presented as an image-led grid. A specimen's slide thumbnail is the primary identifier, not its database ID.
5. **Semantic Readiness**: The search architecture must not preclude future integration of vector-based semantic search as the content library grows.

### Functional Requirements

#### FR-TE-1: Unified Search Bar
- The system must perform text-based similarity matching against Tissue Record titles and scientific names.
- The search must return results as the user types or upon explicit submission.
- An empty search with no filters active must display the full catalogue, paginated.
- The architecture must allow future integration of vector-based semantic search without requiring a redesign of the results interface.

#### FR-TE-2: Hierarchical Faceted Filtering
- Filters must be presented as a nested category tree reflecting biological and anatomical hierarchies (e.g., Kingdom → Phylum → Class, or Organ System → Organ → Tissue Type).
- Each node in the tree must display the count of Tissue Records that belong to that branch.
- Counts must update dynamically as other filters are applied, so students always see how many results remain reachable.
- A student must be able to select nodes from different branches simultaneously (e.g., "Fungi" and "H&E staining" active at the same time).
- Selecting a parent node must include all descendant records in the result set.

#### FR-TE-3: Results Grid
- Results must be displayed as a grid of cards, each showing:
  - A representative slide thumbnail, or a placeholder if no image is available.
  - The Tissue Record name and scientific name.
  - The primary category tags (taxon, organ, staining).
  - The number of associated slides.
- Each card must link directly to the Tissue Record workspace.
- If a tiled slide is available, the card must offer a direct path to the Virtual Microscope viewer.

#### FR-TE-4: Result Count and Feedback
- The interface must always display the total number of records matching the current search and filter combination.
- When no results are found, the system must display a clear message and suggest broadening the search.

### Non-Functional Requirements

#### NFR-TE-1: Responsiveness
- Filter changes and search input must produce updated results within 1 second for a catalogue of up to 500 records.

#### NFR-TE-2: Usability
- A student with no prior knowledge of the catalogue must be able to find a specimen using only the category tree, without typing any text.
- The filter tree must remain usable on mobile screen sizes.

### Acceptance Criteria
- A student can find a tissue record by typing part of its name or scientific name.
- A student can filter results by selecting a taxon node and see only records belonging to that taxon.
- Selecting filters from two different category dimensions (e.g., a taxon and a staining type) returns only records that satisfy both.
- Each category node displays the correct record count, and the count updates when other filters are applied.
- A result card with a tiled slide shows a direct link to the Virtual Microscope.
- The interface displays a meaningful message when no records match the current query.

---
- Animal tissues
- Fungal tissues
- Protozoan samples
- User accounts and contributions
- Advanced search features
- Image annotation tools

## Acceptance Criteria
- Functional web interface for browsing plant tissue images
- At least 20 sample tissue records with associated slides
- Complete CRUD operations for content management
- Responsive design working on desktop and mobile
- Deployable via Docker

## Risk Assessment
- **Timeline Risk**: July deadline may be tight for complete system
- **Content Acquisition**: Need to source appropriate microscopy images
- **Scope Creep**: Focus on plant anatomy only to meet deadline
- **Technical Complexity**: Keep architecture simple but extensible

## References
- Current codebase analysis
- Educational microscopy resources
- Go web development best practices</content>
<parameter name="filePath">/workspaces/tissquest/REQUIREMENTS.md