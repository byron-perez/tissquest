package tissue_records

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mcba/tissquest/cmd/api-server-gin/shared"
	"mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/core/collection"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/taxon"
	coreTR "mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

// WorkspaceViewModel holds all data needed by the workspace templates.
type WorkspaceViewModel struct {
	TissueRecord    coreTR.TissueRecord
	Taxa            []taxon.Taxon
	Collections     []collection.Collection // currently associated (via section assignments)
	Categories      []category.Category     // currently associated
	AvailCats       []category.Category     // not yet associated
	Slides          []slide.DisplaySlide    // for slide gallery
	TissueRecordID  uint                    // for slide gallery
	Crumbs          []wsBreadcrumb
	Errors          map[string]string
	SelectedTaxonID uint
}

type wsBreadcrumb struct {
	Label string
	URL   string
}

var workspaceTemplateFiles = []string{
	"web/templates/layouts/base.html",
	"web/templates/pages/tissue_record_workspace.html",
	"web/templates/includes/main-menu.html",
	"web/templates/includes/breadcrumb.html",
	"web/templates/includes/workspace_summary.html",
	"web/templates/includes/workspace_basic_info.html",
	"web/templates/includes/workspace_collection_section.html",
	"web/templates/includes/workspace_category_section.html",
	"web/templates/includes/slide_gallery.html",
}

var basicInfoTemplateFiles = []string{
	"web/templates/includes/workspace_basic_info.html",
}

var collectionSectionTemplateFiles = []string{
	"web/templates/includes/workspace_collection_section.html",
}

var workspaceSummaryTemplateFiles = []string{
	"web/templates/includes/workspace_summary.html",
}

var categorySectionTemplateFiles = []string{
	"web/templates/includes/workspace_category_section.html",
}

func wsCollectionService() *services.CollectionService {
	return services.NewCollectionService(repositories.NewCollectionRepository(), repositories.NewTissueRecordRepository())
}

func wsCategoryService() *services.CategoryService {
	return services.NewCategoryService(repositories.NewCategoryRepository())
}

func wsSlideService() *services.SlideService {
	return services.NewSlideService(nil, repositories.NewSlideRepository())
}

// subtractCategories returns all categories not present in associated.
func subtractCategories(all, associated []category.Category) []category.Category {
	assocSet := make(map[uint]struct{}, len(associated))
	for _, c := range associated {
		assocSet[c.ID] = struct{}{}
	}
	var result []category.Category
	for _, c := range all {
		if _, found := assocSet[c.ID]; !found {
			result = append(result, c)
		}
	}
	return result
}

// parseID parses the :id path param and returns (id, ok).
func parseID(c *gin.Context) (uint, bool) {
	v, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid tissue record ID")
		return 0, false
	}
	return uint(v), true
}

// loadCollectionsForTR loads collections that contain the given tissue record.
func loadCollectionsForTR(trID uint) ([]collection.Collection, error) {
	allCollections, err := wsCollectionService().ListCollections()
	if err != nil {
		return nil, err
	}
	// Find collections that have this TR assigned in any section
	var result []collection.Collection
	for _, col := range allCollections {
		fullCol, err := wsCollectionService().GetCollection(col.ID)
		if err != nil {
			continue
		}
		if collectionContainsTR(fullCol, trID) {
			result = append(result, col)
		}
	}
	return result, nil
}

func collectionContainsTR(col *collection.Collection, trID uint) bool {
	for _, sec := range col.Sections {
		for _, a := range sec.Assignments {
			if a.TissueRecordID == trID {
				return true
			}
		}
		for _, sub := range sec.Subsections {
			for _, a := range sub.Assignments {
				if a.TissueRecordID == trID {
					return true
				}
			}
		}
	}
	return false
}

// loadCategorySection loads associated and available categories for a tissue record.
func loadCategorySection(trID uint) (assoc []category.Category, avail []category.Category, err error) {
	assoc, err = trService().ListCategories(trID)
	if err != nil {
		return
	}
	all, err := wsCategoryService().List()
	if err != nil {
		return
	}
	avail = subtractCategories(all, assoc)
	return
}

// renderCollectionSection renders the collection section fragment for a tissue record.
func renderCollectionSection(c *gin.Context, tr coreTR.TissueRecord) {
	cols, err := loadCollectionsForTR(tr.ID)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load collections")
		return
	}
	shared.RenderFragment(c, collectionSectionTemplateFiles, "workspace-collection-section", gin.H{
		"TissueRecord": tr,
		"Collections":  cols,
	})
}

// renderCategorySection renders the category section fragment for a tissue record.
func renderCategorySection(c *gin.Context, tr coreTR.TissueRecord) {
	assoc, avail, err := loadCategorySection(tr.ID)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load categories")
		return
	}
	shared.RenderFragment(c, categorySectionTemplateFiles, "workspace-category-section", gin.H{
		"TissueRecord": tr,
		"Categories":   assoc,
		"AvailCats":    avail,
	})
}

// WorkspaceHandler renders the full tissue record workspace page.
func WorkspaceHandler(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	record, status := trService().GetByID(id)
	if status == 0 {
		shared.RenderError(c, http.StatusNotFound, "Tissue record not found")
		return
	}

	taxa, _ := taxonService().List()

	cols, err := loadCollectionsForTR(id)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load collections")
		return
	}

	assocCats, err := trService().ListCategories(id)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load categories")
		return
	}
	allCats, err := wsCategoryService().List()
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load categories")
		return
	}

	slides, _ := wsSlideService().ListDisplayByTissueRecord(id, slide.ImageSizeThumb)

	var selectedTaxonID uint
	if record.TaxonID != nil {
		selectedTaxonID = *record.TaxonID
	}

	shared.RenderPage(c, workspaceTemplateFiles, "content", gin.H{
		"Title":           record.Name,
		"TissueRecord":    record,
		"Taxa":            taxa,
		"Collections":     cols,
		"Categories":      assocCats,
		"AvailCats":       subtractCategories(allCats, assocCats),
		"Slides":          slides,
		"TissueRecordID":  record.ID,
		"SlideCount":      len(slides),
		"CollectionCount": len(cols),
		"SelectedTaxonID": selectedTaxonID,
		"Errors":          map[string]string{},
		"Crumbs": []wsBreadcrumb{
			{Label: "Home", URL: "/"},
			{Label: "Tissue Records", URL: "/tissue_records"},
			{Label: record.Name},
		},
	})
}

// RedirectToWorkspace redirects the old detail URL to the workspace page.
func RedirectToWorkspace(c *gin.Context) {
	id := c.Param("id")
	c.Redirect(http.StatusMovedPermanently, fmt.Sprintf("/tissue_records/%s/workspace", id))
}

// BasicInfoFragment returns the basic-info section fragment (display mode).
func BasicInfoFragment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	record, status := trService().GetByID(id)
	if status == 0 {
		shared.RenderError(c, http.StatusNotFound, "Tissue record not found")
		return
	}

	taxa, _ := taxonService().List()

	var selectedTaxonID uint
	if record.TaxonID != nil {
		selectedTaxonID = *record.TaxonID
	}

	shared.RenderFragment(c, basicInfoTemplateFiles, "workspace-basic-info", gin.H{
		"TissueRecord":    record,
		"Taxa":            taxa,
		"SelectedTaxonID": selectedTaxonID,
		"Errors":          map[string]string{},
	})
}

// SaveBasicInfo validates and persists basic info, then returns a refreshed display fragment.
func SaveBasicInfo(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	existing, status := trService().GetByID(id)
	if status == 0 {
		shared.RenderError(c, http.StatusNotFound, "Tissue record not found")
		return
	}

	name := c.PostForm("name")
	notes := c.PostForm("notes")
	taxonIDStr := c.PostForm("taxon_id")

	if name == "" {
		taxa, _ := taxonService().List()
		var selectedTaxonID uint
		if existing.TaxonID != nil {
			selectedTaxonID = *existing.TaxonID
		}
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, basicInfoTemplateFiles, "workspace-basic-info", gin.H{
			"TissueRecord":    existing,
			"Taxa":            taxa,
			"SelectedTaxonID": selectedTaxonID,
			"Errors":          map[string]string{"name": "Name is required"},
		})
		return
	}

	existing.Name = name
	existing.Notes = notes
	existing.TaxonID = nil

	if taxonIDStr != "" {
		if tid, err := strconv.ParseUint(taxonIDStr, 10, 32); err == nil && tid > 0 {
			uid := uint(tid)
			existing.TaxonID = &uid
		}
	}

	trService().Update(id, &existing)

	// Reload to get updated Taxon association
	updated, _ := trService().GetByID(id)
	taxa, _ := taxonService().List()

	var selectedTaxonID uint
	if updated.TaxonID != nil {
		selectedTaxonID = *updated.TaxonID
	}

	shared.RenderFragment(c, basicInfoTemplateFiles, "workspace-basic-info", gin.H{
		"TissueRecord":    updated,
		"Taxa":            taxa,
		"SelectedTaxonID": selectedTaxonID,
		"Errors":          map[string]string{},
	})
}

// CollectionSectionFragment returns the collection section fragment for a tissue record.
func CollectionSectionFragment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	record, status := trService().GetByID(id)
	if status == 0 {
		shared.RenderError(c, http.StatusNotFound, "Tissue record not found")
		return
	}

	renderCollectionSection(c, record)
}

// AddCategoryToTissueRecord adds a category association and returns the refreshed category section.
func AddCategoryToTissueRecord(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	catIDVal, err := strconv.ParseUint(c.Param("categoryID"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	record, status := trService().GetByID(id)
	if status == 0 {
		shared.RenderError(c, http.StatusNotFound, "Tissue record not found")
		return
	}

	if err := trService().AddCategory(id, uint(catIDVal)); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to add category")
		return
	}

	renderCategorySection(c, record)
}

// RemoveCategoryFromTissueRecord removes a category association and returns the refreshed category section.
func RemoveCategoryFromTissueRecord(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	catIDVal, err := strconv.ParseUint(c.Param("categoryID"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	record, status := trService().GetByID(id)
	if status == 0 {
		shared.RenderError(c, http.StatusNotFound, "Tissue record not found")
		return
	}

	if err := trService().RemoveCategory(id, uint(catIDVal)); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to remove category")
		return
	}

	renderCategorySection(c, record)
}

// CategorySectionFragment returns the category section fragment for a tissue record.
func CategorySectionFragment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	record, status := trService().GetByID(id)
	if status == 0 {
		shared.RenderError(c, http.StatusNotFound, "Tissue record not found")
		return
	}

	renderCategorySection(c, record)
}
