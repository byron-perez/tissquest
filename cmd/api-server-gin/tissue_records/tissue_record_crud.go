package tissue_records

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"mcba/tissquest/cmd/api-server-gin/shared"
	coreTR "mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

const pageSize = 20

var (
	trListTemplateFiles = []string{
		"web/templates/layouts/base.html",
		"web/templates/pages/tissue_record_list.html",
		"web/templates/includes/main-menu.html",
		"web/templates/includes/breadcrumb.html",
	}
	trFormTemplateFiles = []string{
		"web/templates/pages/tissue_record_form.html",
	}
	trRowTemplateFiles = []string{
		"web/templates/includes/tr_row.html",
	}
	trDetailTemplateFiles = []string{
		"web/templates/layouts/base.html",
		"web/templates/pages/tissue_record_detail.html",
		"web/templates/includes/main-menu.html",
		"web/templates/includes/breadcrumb.html",
	}
	trConfirmDeleteTemplateFiles = []string{
		"web/templates/includes/confirm-delete.html",
	}
	trDeleteTriggerTemplateFiles = []string{
		"web/templates/includes/delete-trigger.html",
	}
)

type trBreadcrumbItem struct {
	Label string
	URL   string
}

func trService() *services.TissueRecordService {
	return services.NewTissueRecordService(repositories.NewTissueRecordRepository())
}

func taxonService() *services.TaxonService {
	return services.NewTaxonService(repositories.NewTaxonRepository())
}

// ListTissueRecordsHTML renders the paginated tissue record list page.
func ListTissueRecordsHTML(c *gin.Context) {
	page := 1
	if v, err := strconv.Atoi(c.Query("page")); err == nil && v > 0 {
		page = v
	}

	records, total, err := trService().List(page, pageSize)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load tissue records")
		return
	}

	totalPages := (int(total) + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	var prevPage, nextPage int
	if page > 1 {
		prevPage = page - 1
	}
	if page < totalPages {
		nextPage = page + 1
	}

	shared.RenderPage(c, trListTemplateFiles, "content", gin.H{
		"Title":         "Tissue Records",
		"TissueRecords": records,
		"Page":          page,
		"TotalPages":    totalPages,
		"PrevPage":      prevPage,
		"NextPage":      nextPage,
		"Crumbs": []trBreadcrumbItem{
			{Label: "Home", URL: "/"},
			{Label: "Tissue Records"},
		},
	})
}

// NewTissueRecordForm renders an empty tissue record form fragment.
func NewTissueRecordForm(c *gin.Context) {
	taxa, _ := taxonService().List()
	shared.RenderFragment(c, trFormTemplateFiles, "tr-form", gin.H{
		"TissueRecord":    nil,
		"Taxa":            taxa,
		"SelectedTaxonID": uint(0),
		"Errors":          map[string]string{},
		"CancelURL":       "/tissue_records/new-form-cancel",
		"CancelTarget":    "#tr-form-container",
		"CancelSwap":      "innerHTML",
	})
}

// NewTissueRecordFormCancel clears the tissue record form container.
func NewTissueRecordFormCancel(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// CreateTissueRecordHTML handles form submission to create a new tissue record.
func CreateTissueRecordHTML(c *gin.Context) {
	name := c.PostForm("name")
	notes := c.PostForm("notes")
	taxonIDStr := c.PostForm("taxon_id")
	sectionIDStr := c.PostForm("section_id")

	if name == "" {
		taxa, _ := taxonService().List()
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, trFormTemplateFiles, "tr-form", gin.H{
			"TissueRecord":    nil,
			"Taxa":            taxa,
			"SelectedTaxonID": uint(0),
			"Errors":          map[string]string{"name": "Name is required"},
		})
		return
	}

	tr := &coreTR.TissueRecord{
		Name:  name,
		Notes: notes,
	}

	if taxonIDStr != "" {
		if tid, err := strconv.ParseUint(taxonIDStr, 10, 32); err == nil && tid > 0 {
			uid := uint(tid)
			tr.TaxonID = &uid
		}
	}

	newID := trService().Create(tr)
	tr.ID = newID

	// If section_id is present, assign the TR to that section
	if sectionIDStr != "" {
		if sid, err := strconv.ParseUint(sectionIDStr, 10, 32); err == nil && sid > 0 {
			collSvc := services.NewCollectionService(repositories.NewCollectionRepository(), repositories.NewTissueRecordRepository())
			if _, err := collSvc.AssignTissueRecord(uint(sid), newID); err != nil {
				// Non-fatal: TR was created, assignment failed
				shared.SetFlash(c, fmt.Sprintf("Tissue record \"%s\" created but assignment failed", tr.Name))
			} else {
				shared.SetFlash(c, fmt.Sprintf("Tissue record \"%s\" created and assigned", tr.Name))
			}
			c.Header("HX-Redirect", c.GetHeader("HX-Current-URL"))
			c.Status(http.StatusOK)
			return
		}
	}

	shared.SetFlash(c, fmt.Sprintf("Tissue record \"%s\" created successfully", tr.Name))
	c.Header("HX-Redirect", "/tissue_records")
	c.Status(http.StatusOK)
}

// EditTissueRecordForm renders a pre-populated tissue record form fragment.
func EditTissueRecordForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid tissue record ID")
		return
	}

	record, status := trService().GetByID(uint(id))
	if status == 0 {
		shared.RenderError(c, http.StatusNotFound, "Tissue record not found")
		return
	}

	taxa, _ := taxonService().List()

	var selectedTaxonID uint
	if record.TaxonID != nil {
		selectedTaxonID = *record.TaxonID
	}

	shared.RenderFragment(c, trFormTemplateFiles, "tr-form", gin.H{
		"TissueRecord":    record,
		"Taxa":            taxa,
		"SelectedTaxonID": selectedTaxonID,
		"Errors":          map[string]string{},
		"CancelURL":       fmt.Sprintf("/tissue_records/%d/edit-cancel", record.ID),
		"CancelTarget":    fmt.Sprintf("#tr-row-%d", record.ID),
		"CancelSwap":      "outerHTML",
	})
}

// UpdateTissueRecordHTML handles form submission to update an existing tissue record.
func UpdateTissueRecordHTML(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid tissue record ID")
		return
	}

	existing, status := trService().GetByID(uint(id))
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
		shared.RenderFragment(c, trFormTemplateFiles, "tr-form", gin.H{
			"TissueRecord":    existing,
			"Taxa":            taxa,
			"SelectedTaxonID": selectedTaxonID,
			"Errors":          map[string]string{"name": "Name is required"},
			"CancelURL":       fmt.Sprintf("/tissue_records/%d/edit-cancel", existing.ID),
			"CancelTarget":    fmt.Sprintf("#tr-row-%d", existing.ID),
			"CancelSwap":      "outerHTML",
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

	trService().Update(uint(id), &existing)
	shared.SetFlash(c, fmt.Sprintf("Tissue record \"%s\" updated successfully", existing.Name))

	shared.RenderFragment(c, trRowTemplateFiles, "tr-row", gin.H{
		"ID":    existing.ID,
		"Name":  existing.Name,
		"Notes": existing.Notes,
		"Taxon": existing.Taxon,
	})
}

// DeleteTissueRecordHTML deletes a tissue record and returns an empty fragment.
func DeleteTissueRecordHTML(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid tissue record ID")
		return
	}

	trService().Delete(uint(id))
	shared.SetFlash(c, "Tissue record deleted successfully")

	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// EditCancelTissueRecord restores the original tissue record row when edit is cancelled.
func EditCancelTissueRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid tissue record ID")
		return
	}
	record, status := trService().GetByID(uint(id))
	if status == 0 {
		shared.RenderError(c, http.StatusNotFound, "Tissue record not found")
		return
	}
	shared.RenderFragment(c, trRowTemplateFiles, "tr-row", gin.H{
		"ID":    record.ID,
		"Name":  record.Name,
		"Notes": record.Notes,
		"Taxon": record.Taxon,
	})
}

// ConfirmDeleteTissueRecord renders the confirm-delete fragment for a tissue record row.
func ConfirmDeleteTissueRecord(c *gin.Context) {
	id := c.Param("id")

	shared.RenderFragment(c, trConfirmDeleteTemplateFiles, "confirm-delete", gin.H{
		"DeleteURL": fmt.Sprintf("/tissue_records/%s", id),
		"CancelURL": fmt.Sprintf("/tissue_records/%s/confirm-delete-cancel", id),
		"TargetID":  fmt.Sprintf("tr-row-%s-delete", id),
		"RowTarget": fmt.Sprintf("#tr-row-%s", id),
	})
}

// ConfirmDeleteTissueRecordCancel restores the original delete trigger for a tissue record row.
func ConfirmDeleteTissueRecordCancel(c *gin.Context) {
	id := c.Param("id")

	shared.RenderFragment(c, trDeleteTriggerTemplateFiles, "delete-trigger", gin.H{
		"ConfirmURL": fmt.Sprintf("/tissue_records/%s/confirm-delete", id),
		"Target":     fmt.Sprintf("#tr-row-%s-delete", id),
	})
}

// ViewTissueRecordHTML renders the tissue record detail page.
func ViewTissueRecordHTML(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid tissue record ID")
		return
	}

	record, status := trService().GetByID(uint(id))
	if status == 0 {
		shared.RenderError(c, http.StatusNotFound, "Tissue record not found")
		return
	}

	shared.RenderPage(c, trDetailTemplateFiles, "content", gin.H{
		"Title":        record.Name,
		"TissueRecord": record,
		"Crumbs": []trBreadcrumbItem{
			{Label: "Home", URL: "/"},
			{Label: "Tissue Records", URL: "/tissue_records"},
			{Label: record.Name},
		},
	})
}

// SearchTissueRecords handles GET /tissue_records/search?q=<term>&section_id=<id>&collection_id=<id>
// Returns an HTML fragment listing matching tissue records with "Add" buttons.
func SearchTissueRecords(c *gin.Context) {
	q := c.Query("q")
	sectionIDStr := c.Query("section_id")
	collectionIDStr := c.Query("collection_id")

	var sectionID uint
	if v, err := strconv.ParseUint(sectionIDStr, 10, 32); err == nil {
		sectionID = uint(v)
	}
	var collectionID uint
	if v, err := strconv.ParseUint(collectionIDStr, 10, 32); err == nil {
		collectionID = uint(v)
	}

	// Use a simple list search
	records, _, err := trService().List(1, 1000)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Search failed")
		return
	}

	var results []coreTR.TissueRecord
	ql := strings.ToLower(q)
	for _, r := range records {
		if strings.Contains(strings.ToLower(r.Name), ql) {
			results = append(results, r)
			continue
		}
		if r.Taxon != nil && strings.Contains(strings.ToLower(r.Taxon.Name), ql) {
			results = append(results, r)
		}
	}

	searchTemplateFiles := []string{
		"web/templates/includes/tr_search_results.html",
	}
	shared.RenderFragment(c, searchTemplateFiles, "tr-search-results", gin.H{
		"Results":      results,
		"Query":        q,
		"SectionID":    sectionID,
		"CollectionID": collectionID,
	})
}
