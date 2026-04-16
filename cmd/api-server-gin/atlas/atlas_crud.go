package atlas

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	coreAtlas "mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
	"mcba/tissquest/cmd/api-server-gin/shared"
)

var (
	listTemplateFiles = []string{
		"web/templates/layouts/base.html",
		"web/templates/pages/atlas_list.html",
		"web/templates/includes/main-menu.html",
		"web/templates/includes/breadcrumb.html",
	}
	formTemplateFiles = []string{
		"web/templates/pages/atlas_form.html",
	}
	confirmDeleteTemplateFiles = []string{
		"web/templates/includes/confirm-delete.html",
	}
)

type breadcrumbItem struct {
	Label string
	URL   string
}

func atlasService() *services.AtlasService {
	return services.NewAtlasService(repositories.NewAtlasRepository())
}

// NewAtlasForm renders an empty atlas form fragment for creating a new atlas.
func NewAtlasForm(c *gin.Context) {
	shared.RenderFragment(c, formTemplateFiles, "atlas-form", gin.H{
		"Atlas":        nil,
		"Errors":       map[string]string{},
		"CancelURL":    "/atlases/new-form-cancel",
		"CancelTarget": "#atlas-form-container",
		"CancelSwap":   "innerHTML",
	})
}

// NewAtlasFormCancel clears the atlas form container.
func NewAtlasFormCancel(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// ListAtlasesHTML renders the atlas list page.
func ListAtlasesHTML(c *gin.Context) {
	svc := atlasService()
	atlases, err := svc.ListAtlases()
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load atlases")
		return
	}

	shared.RenderPage(c, listTemplateFiles, "content", gin.H{
		"Title":   "Atlases",
		"Atlases": atlases,
		"Crumbs": []breadcrumbItem{
			{Label: "Home", URL: "/"},
			{Label: "Atlases"},
		},
	})
}

// CreateAtlasHTML handles form submission to create a new atlas.
func CreateAtlasHTML(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	category := c.PostForm("category")

	newAtlas := &coreAtlas.Atlas{
		Name:        name,
		Description: description,
		Category:    category,
	}

	svc := atlasService()
	_, err := svc.CreateAtlas(newAtlas)
	if err != nil {
		errors := map[string]string{}
		switch err {
		case coreAtlas.ErrEmptyName:
			errors["name"] = "Name is required"
		case coreAtlas.ErrNameTooLong:
			errors["name"] = "Name must be 100 characters or fewer"
		default:
			errors["name"] = err.Error()
		}
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, formTemplateFiles, "atlas-form", gin.H{
			"Atlas":  newAtlas,
			"Errors": errors,
		})
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Atlas \"%s\" created successfully", newAtlas.Name))
	c.Header("HX-Redirect", "/atlases")
	c.Status(http.StatusOK)
}

// EditAtlasForm renders a pre-populated atlas form fragment for editing.
func EditAtlasForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid atlas ID")
		return
	}

	svc := atlasService()
	a, err := svc.GetAtlas(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Atlas not found")
		return
	}

	shared.RenderFragment(c, formTemplateFiles, "atlas-form", gin.H{
		"Atlas":        a,
		"Errors":       map[string]string{},
		"CancelURL":    fmt.Sprintf("/atlases/%d/edit-cancel", a.ID),
		"CancelTarget": fmt.Sprintf("#atlas-row-%d", a.ID),
		"CancelSwap":   "outerHTML",
	})
}

// UpdateAtlasHTML handles form submission to update an existing atlas.
func UpdateAtlasHTML(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid atlas ID")
		return
	}

	svc := atlasService()
	existing, err := svc.GetAtlas(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Atlas not found")
		return
	}

	existing.Name = c.PostForm("name")
	existing.Description = c.PostForm("description")
	existing.Category = c.PostForm("category")

	if err := svc.UpdateAtlas(uint(id), existing); err != nil {
		errors := map[string]string{}
		switch err {
		case coreAtlas.ErrEmptyName:
			errors["name"] = "Name is required"
		case coreAtlas.ErrNameTooLong:
			errors["name"] = "Name must be 100 characters or fewer"
		default:
			errors["name"] = err.Error()
		}
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, formTemplateFiles, "atlas-form", gin.H{
			"Atlas":        existing,
			"Errors":       errors,
			"CancelURL":    fmt.Sprintf("/atlases/%d/edit-cancel", existing.ID),
			"CancelTarget": fmt.Sprintf("#atlas-row-%d", existing.ID),
			"CancelSwap":   "outerHTML",
		})
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Atlas \"%s\" updated successfully", existing.Name))

	// Return the updated row fragment
	rowTemplateFiles := []string{"web/templates/includes/atlas_row.html"}
	shared.RenderFragment(c, rowTemplateFiles, "atlas-row", gin.H{
		"ID":          existing.ID,
		"Name":        existing.Name,
		"Description": existing.Description,
		"Category":    existing.Category,
	})
}

// DeleteAtlasHTML handles deletion of an atlas and returns an empty fragment to remove the row.
func DeleteAtlasHTML(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid atlas ID")
		return
	}

	svc := atlasService()
	if err := svc.DeleteAtlas(uint(id)); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to delete atlas")
		return
	}

	shared.SetFlash(c, "Atlas deleted successfully")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// EditCancelAtlas restores the original atlas row when edit is cancelled.
func EditCancelAtlas(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid atlas ID")
		return
	}
	svc := atlasService()
	a, err := svc.GetAtlas(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Atlas not found")
		return
	}
	rowTemplateFiles := []string{"web/templates/includes/atlas_row.html"}
	shared.RenderFragment(c, rowTemplateFiles, "atlas-row", gin.H{
		"ID":          a.ID,
		"Name":        a.Name,
		"Description": a.Description,
		"Category":    a.Category,
	})
}

// ConfirmDeleteAtlas renders the confirm-delete fragment for an atlas row.
func ConfirmDeleteAtlas(c *gin.Context) {
	id := c.Param("id")

	shared.RenderFragment(c, confirmDeleteTemplateFiles, "confirm-delete", gin.H{
		"DeleteURL": fmt.Sprintf("/atlases/%s", id),
		"CancelURL": fmt.Sprintf("/atlases/%s/confirm-delete-cancel", id),
		"TargetID":  fmt.Sprintf("atlas-row-%s-delete", id),
		"RowTarget": fmt.Sprintf("#atlas-row-%s", id),
	})
}

// ConfirmDeleteAtlasCancel restores the original delete trigger for an atlas row.
func ConfirmDeleteAtlasCancel(c *gin.Context) {
	id := c.Param("id")

	deleteTriggerFiles := []string{"web/templates/includes/delete-trigger.html"}
	shared.RenderFragment(c, deleteTriggerFiles, "delete-trigger", gin.H{
		"ConfirmURL": fmt.Sprintf("/atlases/%s/confirm-delete", id),
		"Target":     fmt.Sprintf("#atlas-row-%s-delete", id),
	})
}
