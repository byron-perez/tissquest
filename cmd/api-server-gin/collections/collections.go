package collections

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mcba/tissquest/cmd/api-server-gin/shared"
	coreCollection "mcba/tissquest/internal/core/collection"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

type breadcrumbItem struct {
	Label string
	URL   string
}

var (
	listTemplateFiles = []string{
		"web/templates/layouts/base.html",
		"web/templates/pages/collection_list.html",
		"web/templates/includes/main-menu.html",
		"web/templates/includes/breadcrumb.html",
	}
	formTemplateFiles = []string{
		"web/templates/pages/collection_form.html",
	}
	rowTemplateFiles = []string{
		"web/templates/includes/collection_row.html",
	}
	confirmDeleteTemplateFiles = []string{
		"web/templates/includes/confirm-delete.html",
	}
	deleteTriggerTemplateFiles = []string{
		"web/templates/includes/delete-trigger.html",
	}
)

func collectionService() *services.CollectionService {
	return services.NewCollectionService(
		repositories.NewCollectionRepository(),
		repositories.NewTissueRecordRepository(),
	)
}

// ListCollections renders the collections list page.
func ListCollections(c *gin.Context) {
	cols, err := collectionService().ListCollections()
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load collections")
		return
	}

	shared.RenderPage(c, listTemplateFiles, "content", gin.H{
		"Title":       "Collections",
		"Collections": cols,
		"Crumbs": []breadcrumbItem{
			{Label: "Home", URL: "/"},
			{Label: "Collections"},
		},
	})
}

// NewCollectionForm renders an empty collection form fragment.
func NewCollectionForm(c *gin.Context) {
	shared.RenderFragment(c, formTemplateFiles, "collection-form", gin.H{
		"Collection":   nil,
		"Errors":       map[string]string{},
		"CancelURL":    "/collections/new-form-cancel",
		"CancelTarget": "#collection-form-container",
		"CancelSwap":   "innerHTML",
	})
}

// NewCollectionFormCancel clears the collection form container.
func NewCollectionFormCancel(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// CreateCollection handles form submission to create a new collection.
func CreateCollection(c *gin.Context) {
	name := c.PostForm("name")
	description := c.PostForm("description")
	goals := c.PostForm("goals")
	collType := c.PostForm("type")
	authors := c.PostForm("authors")

	if collType == "" {
		collType = "atlas"
	}

	newCol := &coreCollection.Collection{
		Name:        name,
		Description: description,
		Goals:       goals,
		Type:        coreCollection.CollectionType(collType),
		Authors:     authors,
	}

	_, err := collectionService().CreateCollection(newCol)
	if err != nil {
		errors := map[string]string{}
		switch err {
		case coreCollection.ErrEmptyName:
			errors["name"] = "Name is required"
		case coreCollection.ErrNameTooLong:
			errors["name"] = "Name must be 200 characters or fewer"
		case coreCollection.ErrInvalidType:
			errors["type"] = "Invalid collection type"
		default:
			errors["name"] = err.Error()
		}
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, formTemplateFiles, "collection-form", gin.H{
			"Collection":   newCol,
			"Errors":       errors,
			"CancelURL":    "/collections/new-form-cancel",
			"CancelTarget": "#collection-form-container",
			"CancelSwap":   "innerHTML",
		})
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Collection \"%s\" created successfully", newCol.Name))
	c.Header("HX-Redirect", "/collections")
	c.Status(http.StatusOK)
}

// EditCollectionForm renders a pre-populated collection form fragment for editing.
func EditCollectionForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid collection ID")
		return
	}

	col, err := collectionService().GetCollection(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Collection not found")
		return
	}

	shared.RenderFragment(c, formTemplateFiles, "collection-form", gin.H{
		"Collection":   col,
		"Errors":       map[string]string{},
		"CancelURL":    fmt.Sprintf("/collections/%d/edit-cancel", col.ID),
		"CancelTarget": fmt.Sprintf("#collection-row-%d", col.ID),
		"CancelSwap":   "outerHTML",
	})
}

// EditCancelCollection restores the original collection row when edit is cancelled.
func EditCancelCollection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid collection ID")
		return
	}
	col, err := collectionService().GetCollection(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Collection not found")
		return
	}
	shared.RenderFragment(c, rowTemplateFiles, "collection-row", gin.H{
		"Collection": col,
	})
}

// UpdateCollection handles form submission to update an existing collection.
func UpdateCollection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid collection ID")
		return
	}

	existing, err := collectionService().GetCollection(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Collection not found")
		return
	}

	existing.Name = c.PostForm("name")
	existing.Description = c.PostForm("description")
	existing.Goals = c.PostForm("goals")
	collType := c.PostForm("type")
	if collType == "" {
		collType = "atlas"
	}
	existing.Type = coreCollection.CollectionType(collType)
	existing.Authors = c.PostForm("authors")

	if err := collectionService().UpdateCollection(uint(id), existing); err != nil {
		errors := map[string]string{}
		switch err {
		case coreCollection.ErrEmptyName:
			errors["name"] = "Name is required"
		case coreCollection.ErrNameTooLong:
			errors["name"] = "Name must be 200 characters or fewer"
		case coreCollection.ErrInvalidType:
			errors["type"] = "Invalid collection type"
		default:
			errors["name"] = err.Error()
		}
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, formTemplateFiles, "collection-form", gin.H{
			"Collection":   existing,
			"Errors":       errors,
			"CancelURL":    fmt.Sprintf("/collections/%d/edit-cancel", existing.ID),
			"CancelTarget": fmt.Sprintf("#collection-row-%d", existing.ID),
			"CancelSwap":   "outerHTML",
		})
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Collection \"%s\" updated successfully", existing.Name))

	// Reload to get updated data
	updated, _ := collectionService().GetCollection(uint(id))
	shared.RenderFragment(c, rowTemplateFiles, "collection-row", gin.H{
		"Collection": updated,
	})
}

// DeleteCollection deletes a collection and returns an empty fragment.
func DeleteCollection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid collection ID")
		return
	}

	if err := collectionService().DeleteCollection(uint(id)); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to delete collection")
		return
	}

	shared.SetFlash(c, "Collection deleted successfully")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// ConfirmDeleteCollection renders the confirm-delete fragment for a collection row.
func ConfirmDeleteCollection(c *gin.Context) {
	id := c.Param("id")
	shared.RenderFragment(c, confirmDeleteTemplateFiles, "confirm-delete", gin.H{
		"DeleteURL": fmt.Sprintf("/collections/%s", id),
		"CancelURL": fmt.Sprintf("/collections/%s/confirm-delete-cancel", id),
		"TargetID":  fmt.Sprintf("collection-row-%s-delete", id),
		"RowTarget": fmt.Sprintf("#collection-row-%s", id),
	})
}

// ConfirmDeleteCollectionCancel restores the original delete trigger for a collection row.
func ConfirmDeleteCollectionCancel(c *gin.Context) {
	id := c.Param("id")
	shared.RenderFragment(c, deleteTriggerTemplateFiles, "delete-trigger", gin.H{
		"ConfirmURL": fmt.Sprintf("/collections/%s/confirm-delete", id),
		"Target":     fmt.Sprintf("#collection-row-%s-delete", id),
	})
}
