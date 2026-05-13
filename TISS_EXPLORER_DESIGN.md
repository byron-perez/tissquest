# Tiss Explorer Design Document

## Purpose

Tiss Explorer is a standalone, highly interactive exploration interface for the Tissquest application. It is designed to help users discover tissue records with semantic search, hierarchical category filters, and a rich visual browsing experience without full page reloads.

## Scope

This document defines the intended design and implementation plan for Tiss Explorer.
It covers:
- feature goals
- architecture decisions
- backend modules
- frontend components
- API design
- implementation phases
- open questions

## Goals

- Provide a dedicated page for discovery and exploration: `/explorer`
- Support search by tissue name and metadata
- Support hierarchical category filtering using metacategories
- Enable dynamic, no-full-page-reload interactions
- Keep the solution modular and reusable at the software component level
- Preserve progressive enhancement and compatibility with existing server-side rendering

## User Experience

Tiss Explorer should behave like an exploration tool rather than a traditional CRUD page.
It should include:
- a prominent search input
- a filter panel with category trees and counts
- a results grid with tissue preview cards
- clear filter state and pagination
- dynamic updates as users search and refine filters

## Architecture

### Standalone Page

Tiss Explorer will be implemented as a standalone page:
- route: `/explorer`
- dedicated template: `web/templates/pages/explorer.html`
- dedicated handler package: `cmd/api-server-gin/explorer/`

### Modular Components

The design separates concerns into modular components:
- `ExplorerService` for search and filtering business logic
- `CategoryService` extensions for filter hierarchy and counts
- `ExplorerHandler` for page rendering and API endpoints
- `search.js` or `explorer.js` for client-side state and interactions
- page template for layout and component wiring

### Presentation Layer Strategy

The proposed frontend approach is:
- HTMX as the server communication mechanism for partial updates
- Alpine.js as the client-side reactive state manager for interactivity

This hybrid approach supports high interactivity while keeping the page grounded in server-rendered templates.

## Proposed Backend Modules

### `internal/services/explorer_service.go`

Responsibilities:
- execute search queries across tissue records
- combine text search and category filters
- return paginated results and exploration context
- support exploration-specific metadata

API surface:
```go
func NewExplorerService(trRepo tissuerecord.RepositoryInterface, catRepo category.RepositoryInterface, metaRepo metacategory.RepositoryInterface) *ExplorerService
func (s *ExplorerService) Search(req ExplorationRequest) (*ExplorationResult, error)
func (s *ExplorerService) GetFilterHierarchy(query string) ([]MetacategoryWithCounts, error)
```

### `internal/services/category_service.go` extensions

Add or extend APIs for filter support:
```go
type CategoryWithCount struct {
    category.Category
    TissueRecordCount int
}

type MetacategoryWithCounts struct {
    metacategory.Metacategory
    Categories []CategoryWithCount
    TotalCount int
}

func (s *CategoryService) GetHierarchyWithCounts(query string) ([]MetacategoryWithCounts, error)
```

### `cmd/api-server-gin/explorer/`

Handler package responsibilities:
- render the `/explorer` page
- provide JSON APIs for search and filter data
- register routes

Planned files:
- `routes.go`
- `explorer.go`
- `api.go`
- `templates.go`

## API Design

### Routes

- `GET /explorer` - render the standalone explorer page
- `GET /api/explorer/search` - search results as JSON or partial HTML
- `GET /api/explorer/filters` - filter hierarchy with counts
- `GET /api/explorer/suggestions` - optional autocomplete suggestions

### Search Request

```json
{
  "query": "leaf",
  "category_ids": [12, 34],
  "page": 1,
  "page_size": 20,
  "sort_by": "relevance"
}
```

### Exploration Result

```json
{
  "records": [ ... ],
  "total_count": 123,
  "page": 1,
  "total_pages": 7,
  "context": {
    "query": "leaf",
    "active_filters": [ ... ],
    "available_filters": [ ... ]
  }
}
```

## Frontend Components

### Template: `web/templates/pages/explorer.html`

The page should include:
- search bar
- filter sidebar
- results panel
- summary and pagination controls
- HTMX and Alpine.js wiring

### Client Script: `web/static/js/explorer.js`

The reactive component should manage:
- query state
- selected filters
- loading state
- result count
- URL synchronization

Example methods:
- `init()`
- `explore()`
- `toggleFilter(categoryId)`
- `loadFromURL()`
- `updateURL()`

### Interaction Model

- HTMX performs requests for search and filter updates
- Alpine.js keeps UI state synchronized with the URL
- Debounced search triggers avoid excessive backend calls
- Filter selection updates results without full reload

## Data Model and Reuse

Reuse existing domain concepts:
- `tissuerecord.TissueRecord`
- `category.Category`
- `metacategory.Metacategory`

New data shapes are introduced for Explorer:
- `ExplorationRequest`
- `ExplorationResult`
- `CategoryWithCount`
- `MetacategoryWithCounts`

## Implementation Phases

### Phase 1: Core Explorer Page
- create `/explorer` route
- render basic page with search input
- implement simple text search backend
- show paginated tissue record cards

### Phase 2: Category Filtering
- build filter hierarchy endpoint
- display metacategory groups and category checkboxes
- support multi-category selection
- combine search + filters

### Phase 3: Enhanced Interactivity
- add Alpine.js for state management
- implement debounced search and loading states
- sync state with URL
- support partial updates via HTMX

### Phase 4: Polish and Optimization
- add category counts
- optionally add suggestions/autocomplete
- improve mobile UX
- performance tune backend queries

## Risks and Mitigations

- **Complex client state**: mitigate with Alpine.js and URL syncing
- **HTMX limitations**: use HTMX for server requests and Alpine.js for state
- **Performance on large datasets**: push filtering into backend and add indexes
- **Monolithic page**: keep architecture modular with separate packages and APIs

## Open Questions

1. Should the explorer include slides directly, or only tissue records?
2. Should filter counts be computed live or cached?
3. Do we want infinite scroll or traditional pagination?
4. Should exploration state be saved/bookmarked?
5. What level of educational context should be shown on explorer cards?
6. How should we handle mobile-first UX for the filter panel?

## Summary

Tiss Explorer is a standalone discovery interface that will be built as a modular, interactive page.
The recommended technical approach is a hybrid of HTMX for server-driven updates and Alpine.js for client-side state.
This design keeps the existing server-side rendering model while supporting a modern, no-full-reload experience.
