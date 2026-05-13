package metacategories

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mcba/tissquest/cmd/api-server-gin/shared"
	coreMetacategory "mcba/tissquest/internal/core/metacategory"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

var (
	listTemplateFiles = []string{
		"web/templates/layouts/base.html",
		"web/templates/pages/metacategory_list.html",
		"web/templates/includes/main-menu.html",
		"web/templates/includes/breadcrumb.html",
	}
	formTemplateFiles = []string{
		"web/templates/pages/metacategory_form.html",
	}
	rowTemplateFiles = []string{
		"web/templates/includes/metacategory_row.html",
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

// MetacategoryRow is a view model for a single metacategory table row.
type MetacategoryRow struct {
	ID               uint
	Name             string
	ShortDescription string
	ParentName       string
	ParentID         *uint
	ChildCount       int
}

func metacategoryService() *services.MetacategoryService {
	return services.NewMetacategoryService(repositories.NewMetacategoryRepository())
}

// buildRows converts a slice of metacategories into MetacategoryRow view models.
func buildRows(mcs []coreMetacategory.Metacategory, byID map[uint]coreMetacategory.Metacategory) []MetacategoryRow {
	rows := make([]MetacategoryRow, 0, len(mcs))
	for _, m := range mcs {
		desc := m.Description
		if len(desc) > 60 {
			desc = desc[:60] + "…"
		}
		parentName := ""
		if m.ParentID != nil {
			if p, ok := byID[*m.ParentID]; ok {
				parentName = p.Name
			}
		}
		rows = append(rows, MetacategoryRow{
			ID:               m.ID,
			Name:             m.Name,
			ShortDescription: desc,
			ParentName:       parentName,
			ParentID:         m.ParentID,
			ChildCount:       len(m.Children),
		})
	}
	return rows
}

// indexByID builds a map from metacategory ID to metacategory for fast lookup.
func indexByID(mcs []coreMetacategory.Metacategory) map[uint]coreMetacategory.Metacategory {
	m := make(map[uint]coreMetacategory.Metacategory, len(mcs))
	for _, mc := range mcs {
		m[mc.ID] = mc
	}
	return m
}

// selectedParentID safely dereferences a *uint, returning 0 if nil.
func selectedParentID(m *coreMetacategory.Metacategory) uint {
	if m != nil && m.ParentID != nil {
		return *m.ParentID
	}
	return 0
}

// ListMetacategories renders the metacategories list page.
func ListMetacategories(c *gin.Context) {
	allMcs, err := metacategoryService().List()
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load metacategories")
		return
	}

	byID := indexByID(allMcs)
	rows := buildRows(allMcs, byID)

	shared.RenderPage(c, listTemplateFiles, "content", gin.H{
		"Title":            "Metacategories",
		"MetacategoryRows": rows,
		"Crumbs": []breadcrumbItem{
			{Label: "Home", URL: "/"},
			{Label: "Metacategories"},
		},
	})
}

// NewMetacategoryForm renders an empty metacategory form fragment.
func NewMetacategoryForm(c *gin.Context) {
	allMcs, err := metacategoryService().List()
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load metacategories")
		return
	}

	shared.RenderFragment(c, formTemplateFiles, "metacategory-form", gin.H{
		"Metacategory":  nil,
		"Errors":        map[string]string{},
		"ParentOptions": allMcs,
		"CancelURL":     "/metacategories/new-form-cancel",
		"CancelTarget":  "#metacategory-form-container",
		"CancelSwap":    "innerHTML",
	})
}

// NewMetacategoryFormCancel clears the metacategory form container.
func NewMetacategoryFormCancel(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// CreateMetacategory handles form submission to create a new metacategory.
func CreateMetacategory(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	parentIDStr := c.PostForm("parent_id")

	errors := make(map[string]string)

	if name == "" {
		errors["name"] = "Name is required"
	}

	if len(errors) > 0 {
		c.Status(http.StatusUnprocessableEntity)
		allMcs, _ := metacategoryService().List()
		shared.RenderFragment(c, formTemplateFiles, "metacategory-form", gin.H{
			"Metacategory":  nil,
			"Errors":        errors,
			"ParentOptions": allMcs,
			"CancelURL":     "/metacategories/new-form-cancel",
			"CancelTarget":  "#metacategory-form-container",
			"CancelSwap":    "innerHTML",
		})
		return
	}

	newMc := &coreMetacategory.Metacategory{
		Name:        name,
		Description: description,
	}

	if parentIDStr != "" && parentIDStr != "0" {
		parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err == nil {
			id := uint(parentID)
			newMc.ParentID = &id
		}
	}

	_, err := metacategoryService().Create(newMc)
	if err != nil {
		errors["name"] = err.Error()
		c.Status(http.StatusUnprocessableEntity)
		allMcs, _ := metacategoryService().List()
		shared.RenderFragment(c, formTemplateFiles, "metacategory-form", gin.H{
			"Metacategory":  newMc,
			"Errors":        errors,
			"ParentOptions": allMcs,
			"CancelURL":     "/metacategories/new-form-cancel",
			"CancelTarget":  "#metacategory-form-container",
			"CancelSwap":    "innerHTML",
		})
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Metacategory \"%s\" created successfully", newMc.Name))
	c.Header("HX-Redirect", "/metacategories")
	c.Status(http.StatusOK)
}

// EditMetacategoryForm renders a pre-populated metacategory form fragment for editing.
func EditMetacategoryForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid metacategory ID")
		return
	}

	mc, err := metacategoryService().GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Metacategory not found")
		return
	}

	allMcs, err := metacategoryService().List()
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load metacategories")
		return
	}

	shared.RenderFragment(c, formTemplateFiles, "metacategory-form", gin.H{
		"Metacategory":  mc,
		"Errors":        map[string]string{},
		"ParentOptions": allMcs,
		"CancelURL":     fmt.Sprintf("/metacategories/%d/edit-cancel", mc.ID),
		"CancelTarget":  "#metacategory-form-container",
		"CancelSwap":    "innerHTML",
	})
}

// EditMetacategoryFormCancel clears the metacategory form container.
func EditMetacategoryFormCancel(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// UpdateMetacategory handles form submission to update an existing metacategory.
func UpdateMetacategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid metacategory ID")
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")
	parentIDStr := c.PostForm("parent_id")

	errors := make(map[string]string)

	if name == "" {
		errors["name"] = "Name is required"
	}

	if len(errors) > 0 {
		c.Status(http.StatusUnprocessableEntity)
		mc, _ := metacategoryService().GetByID(uint(id))
		allMcs, _ := metacategoryService().List()
		shared.RenderFragment(c, formTemplateFiles, "metacategory-form", gin.H{
			"Metacategory":  mc,
			"Errors":        errors,
			"ParentOptions": allMcs,
			"CancelURL":     fmt.Sprintf("/metacategories/%d/edit-cancel", uint(id)),
			"CancelTarget":  "#metacategory-form-container",
			"CancelSwap":    "innerHTML",
		})
		return
	}

	mc := &coreMetacategory.Metacategory{
		Name:        name,
		Description: description,
	}

	if parentIDStr != "" && parentIDStr != "0" {
		parentID, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err == nil {
			id := uint(parentID)
			mc.ParentID = &id
		}
	}

	if err := metacategoryService().Update(uint(id), mc); err != nil {
		errors["name"] = err.Error()
		c.Status(http.StatusUnprocessableEntity)
		mc, _ := metacategoryService().GetByID(uint(id))
		allMcs, _ := metacategoryService().List()
		shared.RenderFragment(c, formTemplateFiles, "metacategory-form", gin.H{
			"Metacategory":  mc,
			"Errors":        errors,
			"ParentOptions": allMcs,
			"CancelURL":     fmt.Sprintf("/metacategories/%d/edit-cancel", uint(id)),
			"CancelTarget":  "#metacategory-form-container",
			"CancelSwap":    "innerHTML",
		})
		return
	}

	shared.SetFlash(c, "Metacategory updated successfully")
	c.Header("HX-Redirect", "/metacategories")
	c.Status(http.StatusOK)
}

// DeleteMetacategoryForm renders delete confirmation fragment.
func DeleteMetacategoryForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid metacategory ID")
		return
	}

	mc, err := metacategoryService().GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Metacategory not found")
		return
	}

	shared.RenderFragment(c, confirmDeleteTemplateFiles, "confirm-delete", gin.H{
		"ID":           mc.ID,
		"Label":        "Metacategory",
		"Name":         mc.Name,
		"DeleteURL":    fmt.Sprintf("/metacategories/%d", mc.ID),
		"CancelURL":    fmt.Sprintf("/metacategories/%d/delete-cancel", mc.ID),
		"CancelTarget": "#delete-trigger-container",
		"CancelSwap":   "innerHTML",
	})
}

// DeleteMetacategoryFormCancel clears the delete trigger container.
func DeleteMetacategoryFormCancel(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// DeleteMetacategory handles the actual deletion.
func DeleteMetacategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid metacategory ID")
		return
	}

	mc, err := metacategoryService().GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Metacategory not found")
		return
	}

	if err := metacategoryService().Delete(uint(id)); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to delete metacategory")
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Metacategory \"%s\" deleted successfully", mc.Name))
	c.Header("HX-Redirect", "/metacategories")
	c.Status(http.StatusOK)
}

// GetMetacategoryRow renders a single row for the table.
func GetMetacategoryRow(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid metacategory ID")
		return
	}

	mc, err := metacategoryService().GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Metacategory not found")
		return
	}

	allMcs, err := metacategoryService().List()
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load metacategories")
		return
	}

	byID := indexByID(allMcs)
	row := buildRows([]coreMetacategory.Metacategory{*mc}, byID)[0]

	shared.RenderFragment(c, rowTemplateFiles, "metacategory-row", gin.H{
		"Row": row,
	})
}
