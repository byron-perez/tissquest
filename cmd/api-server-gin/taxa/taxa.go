package taxa

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	coreTaxon "mcba/tissquest/internal/core/taxon"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
	"mcba/tissquest/cmd/api-server-gin/shared"
)

// RankGroup holds taxa for a single rank, used for ordered iteration in templates.
type RankGroup struct {
	Rank string
	Taxa []coreTaxon.Taxon
}

var rankOrder = []coreTaxon.Rank{
	coreTaxon.RankKingdom,
	coreTaxon.RankPhylum,
	coreTaxon.RankClass,
	coreTaxon.RankOrder,
	coreTaxon.RankFamily,
	coreTaxon.RankGenus,
	coreTaxon.RankSpecies,
}

var (
	listTemplateFiles = []string{
		"web/templates/layouts/base.html",
		"web/templates/pages/taxon_list.html",
		"web/templates/includes/main-menu.html",
		"web/templates/includes/breadcrumb.html",
	}
	formTemplateFiles = []string{
		"web/templates/pages/taxon_form.html",
	}
	rowTemplateFiles = []string{
		"web/templates/includes/taxon_row.html",
	}
	confirmDeleteTemplateFiles = []string{
		"web/templates/includes/confirm-delete.html",
	}
	deleteTriggerTemplateFiles = []string{
		"web/templates/includes/delete-trigger.html",
	}
)

type breadcrumbItem struct {
	Label string
	URL   string
}

func taxonService() *services.TaxonService {
	return services.NewTaxonService(repositories.NewTaxonRepository())
}

// groupByRank builds an ordered slice of RankGroups from a flat taxon list.
func groupByRank(all []coreTaxon.Taxon) []RankGroup {
	byRank := make(map[coreTaxon.Rank][]coreTaxon.Taxon)
	for _, t := range all {
		byRank[t.Rank] = append(byRank[t.Rank], t)
	}
	var groups []RankGroup
	for _, rank := range rankOrder {
		if taxa, ok := byRank[rank]; ok {
			groups = append(groups, RankGroup{Rank: string(rank), Taxa: taxa})
		}
	}
	return groups
}

// selectedParentID safely dereferences a *uint, returning 0 if nil.
func selectedParentID(t *coreTaxon.Taxon) uint {
	if t != nil && t.ParentID != nil {
		return *t.ParentID
	}
	return 0
}

// ListTaxa renders the taxa list page grouped by rank in biological order.
func ListTaxa(c *gin.Context) {
	svc := taxonService()
	all, err := svc.List()
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load taxa")
		return
	}

	shared.RenderPage(c, listTemplateFiles, "content", gin.H{
		"Title":        "Taxa",
		"RankedGroups": groupByRank(all),
		"AllTaxa":      all,
		"Crumbs": []breadcrumbItem{
			{Label: "Home", URL: "/"},
			{Label: "Taxa"},
		},
	})
}

// NewTaxonForm renders an empty taxon form fragment.
func NewTaxonForm(c *gin.Context) {
	svc := taxonService()
	all, _ := svc.List()

	shared.RenderFragment(c, formTemplateFiles, "taxon-form", gin.H{
		"Taxon":            nil,
		"Errors":           map[string]string{},
		"AllTaxa":          all,
		"SelectedParentID": uint(0),
		"CancelURL":        "/taxa/new-form-cancel",
		"CancelTarget":     "#taxon-form-container",
		"CancelSwap":       "innerHTML",
	})
}

// EditCancelTaxon restores the original taxon row when edit is cancelled.
func EditCancelTaxon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid taxon ID")
		return
	}

	svc := taxonService()
	t, err := svc.GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Taxon not found")
		return
	}

	shared.RenderFragment(c, rowTemplateFiles, "taxon-row", gin.H{
		"ID":     t.ID,
		"Name":   t.Name,
		"Rank":   t.Rank,
		"Parent": t.Parent,
	})
}

// NewTaxonFormCancel clears the taxon form container.
func NewTaxonFormCancel(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// CreateTaxon handles form submission to create a new taxon.
func CreateTaxon(c *gin.Context) {
	name := c.PostForm("name")
	rank := coreTaxon.Rank(c.PostForm("rank"))
	parentIDStr := c.PostForm("parent_id")

	newTaxon := &coreTaxon.Taxon{
		Name: name,
		Rank: rank,
	}
	if parentIDStr != "" {
		pid, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err == nil && pid > 0 {
			v := uint(pid)
			newTaxon.ParentID = &v
		}
	}

	svc := taxonService()
	_, err := svc.Create(newTaxon)
	if err != nil {
		errors := map[string]string{}
		switch err {
		case coreTaxon.ErrEmptyName:
			errors["name"] = "Name is required"
		case coreTaxon.ErrInvalidRank:
			errors["rank"] = "Please select a valid rank"
		default:
			errors["name"] = err.Error()
		}
		all, _ := svc.List()
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, formTemplateFiles, "taxon-form", gin.H{
			"Taxon":            newTaxon,
			"Errors":           errors,
			"AllTaxa":          all,
			"SelectedParentID": selectedParentID(newTaxon),
		})
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Taxon \"%s\" created successfully", newTaxon.Name))
	// Re-fetch and return OOB list refresh via hx-push-url redirect
	c.Header("HX-Redirect", "/taxa")
	c.Status(http.StatusOK)
}

// EditTaxonForm renders a pre-populated taxon form fragment for editing.
func EditTaxonForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid taxon ID")
		return
	}

	svc := taxonService()
	t, err := svc.GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Taxon not found")
		return
	}

	all, _ := svc.List()
	shared.RenderFragment(c, formTemplateFiles, "taxon-form", gin.H{
		"Taxon":            t,
		"Errors":           map[string]string{},
		"AllTaxa":          all,
		"SelectedParentID": selectedParentID(t),
		"CancelURL":        fmt.Sprintf("/taxa/%d/edit-cancel", t.ID),
		"CancelTarget":     fmt.Sprintf("#taxon-row-%d", t.ID),
		"CancelSwap":       "outerHTML",
	})
}

// UpdateTaxon handles form submission to update an existing taxon.
func UpdateTaxon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid taxon ID")
		return
	}

	svc := taxonService()
	existing, err := svc.GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Taxon not found")
		return
	}

	existing.Name = c.PostForm("name")
	existing.Rank = coreTaxon.Rank(c.PostForm("rank"))

	parentIDStr := c.PostForm("parent_id")
	if parentIDStr != "" {
		pid, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err == nil && pid > 0 {
			v := uint(pid)
			existing.ParentID = &v
		} else {
			existing.ParentID = nil
		}
	} else {
		existing.ParentID = nil
	}

	if err := svc.Update(uint(id), existing); err != nil {
		errors := map[string]string{}
		switch err {
		case coreTaxon.ErrEmptyName:
			errors["name"] = "Name is required"
		case coreTaxon.ErrInvalidRank:
			errors["rank"] = "Please select a valid rank"
		default:
			errors["name"] = err.Error()
		}
		all, _ := svc.List()
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, formTemplateFiles, "taxon-form", gin.H{
			"Taxon":            existing,
			"Errors":           errors,
			"AllTaxa":          all,
			"SelectedParentID": selectedParentID(existing),
			"CancelURL":        fmt.Sprintf("/taxa/%d/edit-cancel", existing.ID),
			"CancelTarget":     fmt.Sprintf("#taxon-row-%d", existing.ID),
			"CancelSwap":       "outerHTML",
		})
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Taxon \"%s\" updated successfully", existing.Name))
	shared.RenderFragment(c, rowTemplateFiles, "taxon-row", gin.H{
		"ID":     existing.ID,
		"Name":   existing.Name,
		"Rank":   existing.Rank,
		"Parent": existing.Parent,
	})
}

// DeleteTaxon deletes a taxon and returns an empty fragment to remove the row.
func DeleteTaxon(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid taxon ID")
		return
	}

	svc := taxonService()
	if err := svc.Delete(uint(id)); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to delete taxon")
		return
	}

	shared.SetFlash(c, "Taxon deleted successfully")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// ConfirmDeleteTaxon renders the confirm-delete fragment for a taxon row.
func ConfirmDeleteTaxon(c *gin.Context) {
	id := c.Param("id")

	shared.RenderFragment(c, confirmDeleteTemplateFiles, "confirm-delete", gin.H{
		"DeleteURL": fmt.Sprintf("/taxa/%s", id),
		"CancelURL": fmt.Sprintf("/taxa/%s/confirm-delete-cancel", id),
		"TargetID":  fmt.Sprintf("taxon-row-%s-delete", id),
		"RowTarget": fmt.Sprintf("#taxon-row-%s", id),
	})
}

// ConfirmDeleteTaxonCancel restores the original delete trigger for a taxon row.
func ConfirmDeleteTaxonCancel(c *gin.Context) {
	id := c.Param("id")

	shared.RenderFragment(c, deleteTriggerTemplateFiles, "delete-trigger", gin.H{
		"ConfirmURL": fmt.Sprintf("/taxa/%s/confirm-delete", id),
		"Target":     fmt.Sprintf("#taxon-row-%s-delete", id),
	})
}
