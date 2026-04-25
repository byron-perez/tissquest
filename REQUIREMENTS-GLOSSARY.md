# Requirements Glossary

This glossary defines key terms used in the Tissquest requirements specification to ensure clarity and avoid ambiguity. Terms are organized by domain (biology, software architecture, microscopy).

## Biological/Microscopy Terms

### Atlas
**Definition**: A curated collection of biological tissue samples organized for educational purposes. In the context of Tissquest, an atlas represents a thematic grouping of tissue records (e.g., "Plant Anatomy 101" or "Leaf Tissue Types").

**Context**: Differs from traditional anatomical atlases by being digital and focused on microscopy education rather than gross anatomy.

**Usage in Code**: `Atlas` struct in `internal/core/atlas/atlas.go`, represents domain entity with metadata and associated tissue records.

### Tissue
**Definition**: Biological tissue - a group of cells that perform similar functions. In plant anatomy context, includes tissues like parenchyma, xylem, phloem, epidermis, etc.

**Context**: Tissquest focuses on plant tissues initially, with plans to expand to animal and fungal tissues.

**Usage in Code**: Referenced in `TissueRecord` entities and category classifications.

### Tissue Record
**Definition**: A digital record representing an individual biological tissue specimen, including metadata (scientific name, taxonomic classification, notes) and associated microscopy slides.

**Context**: Core data entity that connects biological specimens with their digital representations.

**Usage in Code**: `TissueRecord` struct in `internal/core/tissuerecord/tissuerecord.go`.

### Slide
**Definition**: A prepared microscopy slide containing a thin section of biological tissue, stained and mounted for microscopic examination.

**Context**: Digital representation includes image URL/file reference, magnification level, and staining technique used.

**Usage in Code**: `Slide` struct in `internal/core/slide/slide.go`, associated with tissue records.

### Staining/Stain
**Definition**: Histological staining - the process of applying dyes to tissue samples to enhance contrast and reveal cellular structures under microscopy.

**Context**: Common techniques include H&E (Hematoxylin and Eosin), Masson's Trichrome, PAS (Periodic Acid-Schiff), etc.

**Usage in Code**: `Staining` entity referenced in slide metadata and category classifications.

### Magnification
**Definition**: The degree to which an image is enlarged under a microscope, typically expressed as a multiplier (e.g., 10x, 40x, 100x).

**Context**: Indicates the level of detail visible in microscopy images.

**Usage in Code**: Stored as metadata in `Slide` entities.

### Taxonomic Classification
**Definition**: The scientific categorization of organisms according to a hierarchical system (Domain → Kingdom → Phylum → Class → Order → Family → Genus → Species).

**Context**: In Tissquest, used to organize tissue records by species, organ, tissue type, and staining method.

**Usage in Code**: `Category` entities in `internal/core/category/category.go` with hierarchical relationships.

## Software Architecture Terms

### Repository (Data Access Pattern)
**Definition**: A design pattern that mediates between the domain and data mapping layers, providing a collection-like interface for accessing domain objects.

**Context**: Follows the Repository pattern to abstract data access logic and enable dependency inversion.

**Usage in Code**: `RepositoryInterface` in each domain package (e.g., `atlas.RepositoryInterface`), implemented by GORM and PostgreSQL repositories.

### Service (Business Logic Layer)
**Definition**: A layer that contains business logic and orchestrates operations across multiple repositories.

**Context**: Implements use cases and coordinates domain objects, following clean architecture principles.

**Usage in Code**: Service structs like `AtlasService` in `internal/services/`, receive repository interfaces via dependency injection.

### Handler (HTTP Handler)
**Definition**: A function that processes HTTP requests and returns responses in a web framework.

**Context**: Gin framework handlers that map routes to business logic execution.

**Usage in Code**: Functions in `cmd/api-server-gin/` packages (e.g., `atlas.ViewAtlas`).

### Migration (Database Migration)
**Definition**: The process of updating database schema to match application data models.

**Context**: Automatic schema creation/updates using GORM's AutoMigrate feature.

**Usage in Code**: `migration.RunMigration()` in `internal/persistence/migration/migration.go`.

## Technology Terms

### GORM
**Definition**: A popular ORM (Object-Relational Mapping) library for Go that provides database operations through struct mappings.

**Context**: Used for database interactions with SQLite and PostgreSQL backends.

**Usage in Code**: Repository implementations in `internal/persistence/repositories/gorm_*.go`.

### Gin
**Definition**: A high-performance HTTP web framework for Go, known for its speed and middleware support.

**Context**: Provides routing, middleware, and template rendering for the web interface.

**Usage in Code**: Router setup in `cmd/api-server-gin/main.go`, handlers throughout the `cmd/api-server-gin/` directory.

## Virtual Microscope Terms

### Tiled Image Format
**Definition**: A representation of a high-resolution image split into a pyramid of small, fixed-size tiles at multiple zoom levels. Only the tiles covering the currently visible region need to be downloaded, making very large images practical to view in a browser.

**Context**: The standard format used by digital pathology and web-based microscopy viewers. Enables the performance requirement of sub-2-second initial load regardless of source image size.

**Usage in Requirements**: Referenced in [FR-VM-1](REQUIREMENTS.md#fr-vm-1-image-preparation-pipeline) and [FR-VM-2](REQUIREMENTS.md#fr-vm-2-slide-metadata-extension). Technology-specific format details (DZI) are in [Virtual Microscope — Technical Design](VIRTUAL-MICROSCOPE-TECH.md#image-tiling-fr-vm-1).

### Base Magnification
**Definition**: The objective lens power used on the physical microscope when the source image was captured (e.g., 4×, 10×, 40×). It is a property of the slide, not of the viewer.

**Context**: Required to correctly map viewer zoom levels to real objective equivalents in the Objective Lens Switcher.

**Usage in Requirements**: Referenced in [FR-VM-2](REQUIREMENTS.md#fr-vm-2-slide-metadata-extension) and [FR-VM-5](REQUIREMENTS.md#fr-vm-5-objective-lens-switcher).

### Spatial Calibration
**Definition**: The physical size of one image pixel expressed in real-world units (micrometers per pixel, µm/px). Derived from the microscope objective and camera sensor specifications at the time of capture.

**Context**: The value that allows the scale bar to display accurate physical measurements at any zoom level.

**Usage in Requirements**: Referenced in [FR-VM-2](REQUIREMENTS.md#fr-vm-2-slide-metadata-extension) and [FR-VM-6](REQUIREMENTS.md#fr-vm-6-scale-indicator).

### Home View
**Definition**: A curated starting position (viewport center and zoom level) within a slide, chosen by the content author to highlight the most educationally relevant region of the specimen.

**Context**: Ensures students open a slide at a meaningful anatomical feature rather than a default fit-to-screen view.

**Usage in Requirements**: Referenced in [FR-VM-7](REQUIREMENTS.md#fr-vm-7-curated-home-view).

### Objective Lens Switcher
**Definition**: A UI control that allows the user to jump directly to a zoom level equivalent to a standard microscope objective (4×, 10×, 40×), analogous to rotating the nosepiece on a physical microscope.

**Context**: Core part of the microscope analogy that makes the viewer feel familiar to students with lab experience.

**Usage in Requirements**: Defined in [FR-VM-5](REQUIREMENTS.md#fr-vm-5-objective-lens-switcher).

### Image Pyramid
**Definition**: The multi-resolution representation produced by the tiling pipeline. The same specimen is stored at many zoom levels simultaneously — from a single thumbnail tile at the top to hundreds of full-resolution tiles at the bottom. The viewer always fetches tiles from the level that matches the current zoom, so the number of tiles downloaded stays constant regardless of the total image size.

**Context**: The pyramid is what makes the "progressive detail" property in the requirements achievable. It is generated once by `vips dzsave` and never changes.

**Usage in Requirements**: Underpins [FR-VM-1](REQUIREMENTS.md#fr-vm-1-image-preparation-pipeline) and [NFR-VM-1](REQUIREMENTS.md#nfr-vm-1-initial-load-performance). Technical detail in [VIRTUAL-MICROSCOPE-TECH.md — The Image Pyramid](VIRTUAL-MICROSCOPE-TECH.md#the-image-pyramid).

### Viewport Coordinates
**Definition**: The normalized coordinate system used by the interactive viewer, where the full image width equals 1.0. A position is expressed as `{x, y}` values between 0.0 and 1.0, and zoom as a multiplier relative to the fit-to-screen state. Resolution-independent — the same coordinates are valid regardless of source image dimensions.

**Context**: Used to store and restore the [Home View](REQUIREMENTS-GLOSSARY.md#home-view) for each slide. Also the format used if a student shares a link to a specific region of a specimen.

**Usage in Requirements**: Referenced in [FR-VM-7](REQUIREMENTS.md#fr-vm-7-curated-home-view). Technical detail in [VIRTUAL-MICROSCOPE-TECH.md — Viewport Coordinates](VIRTUAL-MICROSCOPE-TECH.md#viewport-coordinates).

### S3 Direct Fetch
**Definition**: An architectural pattern where the browser retrieves image tiles directly from the object store (AWS S3), bypassing the application backend entirely. The backend's only role is to provide the URL of the tile set; all subsequent image data flows from S3 to the browser.

**Context**: The pattern that makes it possible to serve 50 concurrent users zooming on high-resolution slides without increasing load on the Go application or the database.

**Usage in Requirements**: Supports [NFR-VM-1](REQUIREMENTS.md#nfr-vm-1-initial-load-performance) and [NFR-VM-2](REQUIREMENTS.md#nfr-vm-2-stability-under-adverse-conditions). Technical detail in [VIRTUAL-MICROSCOPE-TECH.md — Architecture: S3 Direct Fetch Model](VIRTUAL-MICROSCOPE-TECH.md#architecture-s3-direct-fetch-model).

### CORS (Cross-Origin Resource Sharing)
**Definition**: A browser security mechanism that blocks a web page from requesting resources from a different domain unless that domain explicitly permits it. In this system, the browser (served from `tissquest.com`) requests tiles from `s3.amazonaws.com`, which requires an explicit CORS policy on the S3 bucket.

**Context**: A required infrastructure configuration step — without it the viewer will appear blank even though all code is correct.

**Usage in Requirements**: Infrastructure prerequisite for [FR-VM-1](REQUIREMENTS.md#fr-vm-1-image-preparation-pipeline). Configuration detail in [VIRTUAL-MICROSCOPE-TECH.md — CORS Configuration](VIRTUAL-MICROSCOPE-TECH.md#cors-configuration-aws-s3).


- [Requirements Specification](REQUIREMENTS.md) - Main requirements document
- [Virtual Microscope — Technical Design](VIRTUAL-MICROSCOPE-TECH.md) - Technology-specific implementation details for the viewer feature
- [Workspace Instructions](.github/copilot-instructions.md) - Development guidelines and architecture overview</content>
<parameter name="filePath">/workspaces/tissquest/REQUIREMENTS-GLOSSARY.md