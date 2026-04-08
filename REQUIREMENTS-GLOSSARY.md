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

## Related Documentation
- [Requirements Specification](REQUIREMENTS.md) - Main requirements document
- [Workspace Instructions](.github/copilot-instructions.md) - Development guidelines and architecture overview</content>
<parameter name="filePath">/workspaces/tissquest/REQUIREMENTS-GLOSSARY.md