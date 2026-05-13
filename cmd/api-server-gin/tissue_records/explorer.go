package tissue_records

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"mcba/tissquest/cmd/api-server-gin/shared"
	"mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

const explorerPageSize = 24

var explorerTemplateFiles = []string{
	"web/templates/layouts/base.html",
	"web/templates/pages/tissue_explorer.html",
	"web/templates/includes/main-menu.html",
	"web/templates/includes/breadcrumb.html",
	"web/templates/includes/explorer_results.html",
}

var explorerResultsTemplateFiles = []string{
	"web/templates/includes/explorer_results.html",
}

func categoryService() *services.CategoryService {
	return services.NewCategoryService(repositories.NewCategoryRepository())
}

// ExplorerPage renders the full TissExplorer page.
// GET /tissue_records
func ExplorerPage(c *gin.Context) {
	q := c.Query("q")
	page := 1
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = v
	}
	categoryIDs := parseCategoryIDs(c.Query("categories"))

	records, total, _ := trService().Search(q, categoryIDs, page, explorerPageSize)
	categoriesWithCounts, _ := categoryService().ListWithCounts()

	totalPages := int((total + int64(explorerPageSize) - 1) / int64(explorerPageSize))
	if totalPages < 1 {
		totalPages = 1
	}

	prevPage, nextPage := 0, 0
	if page > 1 {
		prevPage = page - 1
	}
	if page < totalPages {
		nextPage = page + 1
	}

	data := gin.H{
		"Title":         "TissExplorer — Explorador de Tejidos",
		"Query":         q,
		"CategoryIDs":   categoryIDs,
		"TissueRecords": records,
		"Total":         total,
		"Page":          page,
		"TotalPages":    totalPages,
		"PrevPage":      prevPage,
		"NextPage":      nextPage,
		"FilterTree":    buildFilterTree(categoriesWithCounts, categoryIDs),
		"Crumbs": []trBreadcrumbItem{
			{Label: "Inicio", URL: "/"},
			{Label: "Explorador de Tejidos"},
		},
	}

	if shared.IsHTMX(c) {
		shared.RenderFragment(c, explorerResultsTemplateFiles, "explorer-results", data)
		return
	}
	shared.RenderPage(c, explorerTemplateFiles, "content", data)
}

// parseCategoryIDs parses a comma-separated list of category IDs from a query param.
func parseCategoryIDs(raw string) []uint {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		if v, err := strconv.ParseUint(strings.TrimSpace(p), 10, 32); err == nil {
			ids = append(ids, uint(v))
		}
	}
	return ids
}

// FilterNode is a view model for one node in the category filter tree.
type FilterNode struct {
	category.CategoryWithCount
	Children []FilterNode
	Active   bool
}

// buildFilterTree assembles a two-level tree grouped by category type.
func buildFilterTree(cats []category.CategoryWithCount, activeIDs []uint) []FilterNode {
	activeSet := make(map[uint]bool, len(activeIDs))
	for _, id := range activeIDs {
		activeSet[id] = true
	}

	// Index by ID for parent lookup
	byID := make(map[uint]*category.CategoryWithCount, len(cats))
	for i := range cats {
		byID[cats[i].ID] = &cats[i]
	}

	// Build tree: roots first, then attach children
	var roots []FilterNode
	children := make(map[uint][]FilterNode)

	for _, c := range cats {
		node := FilterNode{CategoryWithCount: c, Active: activeSet[c.ID]}
		if c.ParentID == nil {
			roots = append(roots, node)
		} else {
			children[*c.ParentID] = append(children[*c.ParentID], node)
		}
	}

	// Attach children to roots
	for i := range roots {
		roots[i].Children = children[roots[i].ID]
	}
	return roots
}
