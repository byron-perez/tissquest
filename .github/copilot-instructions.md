# Tissquest Workspace Instructions

## Overview
This workspace contains the tissquest Go application, an educational platform for studying microscopical images of biological tissues, initially focused on plant anatomy. The project follows hexagonal/clean architecture with Gin web framework, GORM ORM, and supports SQLite/PostgreSQL databases.

## Build and Run
- **Development**: Use `air` for hot-reload (configured in `.air.toml`)
- **Build**: `go build -o ./tmp/main ./cmd/api-server-gin`
- **Run**: `./tmp/main` (serves on :8080)
- **Docker**: `docker build -t tissquest . && docker run -p 8080:8080 -e DB_TYPE=sqlite tissquest`
- **Tests**: `go test ./...` (integration tests in `internal/core/tissuerecord/tests/`)

## Architecture
Follows hexagonal architecture:
- **Core Domain** (`internal/core/`): Business entities with validation
- **Services** (`internal/services/`): Use cases and orchestration
- **Persistence** (`internal/persistence/`): GORM repositories and migrations
- **Handlers** (`cmd/api-server-gin/`): Gin HTTP endpoints
- **Web** (`web/`): Static assets and HTML templates

Each domain package defines a `RepositoryInterface` for dependency inversion.

## Database Configuration
- Supports SQLite (default) and PostgreSQL via `DB_TYPE` env var
- Copy `.env.example` to `.env` and configure connection details
- Migrations run automatically on startup via `migration.RunMigration()`
- Models: `TissueRecordModel`, `SlideModel`, `AtlasModel`, `StainingModel`

## Key Conventions
- **Dependency Injection**: Services receive repositories as constructor parameters
- **Validation**: Domain entities implement `Validate() error`
- **Repository Pattern**: Separate implementations for different DB backends
- **Error Handling**: Mix of panic and returned errors (inconsistent)
- **Module Path**: `mcba/tissquest`

## Common Pitfalls
- Hard-coded PostgreSQL repo in `index.go` (fails with SQLite)
- `AtlasModel` missing from migrations (table not created)
- No connection pooling (performance issue)
- Tests require real DB connection
- Working directory sensitive for web assets

## Domain Concepts
- **Atlas**: Collection of tissue samples
- **TissueRecord**: Individual specimen with metadata
- **Slide**: Microscopy image with magnification/staining
- **Category**: Hierarchical taxonomic classifications

## Links
- [Requirements](REQUIREMENTS.md) - High-level requirements and use cases
- [README.md](README.md) - Project overview
- [CHANGELOG.md](CHANGELOG.md) - Version history</content>
<parameter name="filePath">/workspaces/tissquest/.github/copilot-instructions.md