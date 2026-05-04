package categories

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mcba/tissquest/cmd/api-server-gin/shared"
	coreCategory "mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

var (
	listTemplateFiles = []string{
		"web/templates/layouts/base.html",
		"web/templates/pages/category_list.html",
		"web/templates/includes/main-menu.html",
		"web/templates/includes/breadcrumb.html",
	}
	formTemplateFiles = []string{
		"web/templates/pages/category_form.html",
	}
	rowTemplateFiles = []string{
		"web/templates/includes/category_row.html",
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

// CategoryRow is a view model for a single category table row.
type CategoryRow struct {
	ID               uint
	Name             string
	Type             coreCategory.CategoryType
	ShortDescription string
	ParentName       string
	ParentID         *uint
}

func categoryService() *services.CategoryService {
	return services.NewCategoryService(repositories.NewCategoryRepository())
}

// buildRows converts a slice of categories into CategoryRow view models,
// resolving parent names from the provided lookup map.
func buildRows(cats []coreCategory.Category, byID map[uint]coreCategory.Category) []CategoryRow {
	rows := make([]CategoryRow, 0, len(cats))
	for _, c := range cats {
		desc := c.Description
		if len(desc) > 60 {
			desc = desc[:60] + "…"
		}
		parentName := ""
		if c.ParentID != nil {
			if p, ok := byID[*c.ParentID]; ok {
				parentName = p.Name
			}
		}
		rows = append(rows, CategoryRow{
			ID:               c.ID,
			Name:             c.Name,
			Type:             c.Type,
			ShortDescription: desc,
			ParentName:       parentName,
			ParentID:         c.ParentID,
		})
	}
	return rows
}

// indexByID builds a map from category ID to category for fast lookup.
func indexByID(cats []coreCategory.Category) map[uint]coreCategory.Category {
	m := make(map[uint]coreCategory.Category, len(cats))
	for _, c := range cats {
		m[c.ID] = c
	}
	return m
}

// selectedParentID safely dereferences a *uint, returning 0 if nil.
func selectedParentID(c *coreCategory.Category) uint {
	if c != nil && c.ParentID != nil {
		return *c.ParentID
	}
	return 0
}

// ListCategories renders the categories list page.
func ListCategories(c *gin.Context) {
	svc := categoryService()
	all, err := svc.List()
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load categories")
		return
	}

	byID := indexByID(all)
	rows := buildRows(all, byID)

	shared.RenderPage(c, listTemplateFiles, "content", gin.H{
		"Title":         "Categories",
		"Categories":    rows,
		"AllCategories": all,
		"Crumbs": []breadcrumbItem{
			{Label: "Home", URL: "/"},
			{Label: "Categories"},
		},
	})
}

// NewCategoryForm renders an empty category form fragment.
func NewCategoryForm(c *gin.Context) {
	svc := categoryService()
	all, _ := svc.List()

	shared.RenderFragment(c, formTemplateFiles, "category-form", gin.H{
		"Category":         nil,
		"Errors":           map[string]string{},
		"AllCategories":    all,
		"SelectedParentID": uint(0),
		"CancelURL":        "/categories/new-form-cancel",
		"CancelTarget":     "#category-form-container",
		"CancelSwap":       "innerHTML",
	})
}

// NewCategoryFormCancel clears the category form container.
func NewCategoryFormCancel(c *gin.Context) {
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// CreateCategory handles form submission to create a new category.
func CreateCategory(c *gin.Context) {
	name := c.PostForm("name")
	catType := coreCategory.CategoryType(c.PostForm("type"))
	description := c.PostForm("description")
	parentIDStr := c.PostForm("parent_id")

	newCat := &coreCategory.Category{
		Name:        name,
		Type:        catType,
		Description: description,
	}
	if parentIDStr != "" {
		pid, err := strconv.ParseUint(parentIDStr, 10, 32)
		if err == nil && pid > 0 {
			v := uint(pid)
			newCat.ParentID = &v
		}
	}

	svc := categoryService()
	_, err := svc.Create(newCat)
	if err != nil {
		errs := map[string]string{}
		switch err {
		case coreCategory.ErrEmptyName:
			errs["name"] = "Name is required"
		case coreCategory.ErrInvalidType:
			errs["type"] = "Please select a valid type"
		case coreCategory.ErrCircularParent:
			errs["parent_id"] = "A category cannot be its own parent"
		default:
			errs["name"] = err.Error()
		}
		all, _ := svc.List()
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, formTemplateFiles, "category-form", gin.H{
			"Category":         newCat,
			"Errors":           errs,
			"AllCategories":    all,
			"SelectedParentID": selectedParentID(newCat),
		})
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Category \"%s\" created successfully", newCat.Name))
	c.Header("HX-Redirect", "/categories")
	c.Status(http.StatusOK)
}

// EditCategoryForm renders a pre-populated category form fragment for editing.
func EditCategoryForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	svc := categoryService()
	cat, err := svc.GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Category not found")
		return
	}

	all, _ := svc.List()
	shared.RenderFragment(c, formTemplateFiles, "category-form", gin.H{
		"Category":         cat,
		"Errors":           map[string]string{},
		"AllCategories":    all,
		"SelectedParentID": selectedParentID(cat),
		"CancelURL":        fmt.Sprintf("/categories/%d/edit-cancel", cat.ID),
		"CancelTarget":     fmt.Sprintf("#category-row-%d", cat.ID),
		"CancelSwap":       "outerHTML",
	})
}

// UpdateCategory handles form submission to update an existing category.
func UpdateCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	svc := categoryService()
	existing, err := svc.GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Category not found")
		return
	}

	existing.Name = c.PostForm("name")
	existing.Type = coreCategory.CategoryType(c.PostForm("type"))
	existing.Description = c.PostForm("description")

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
		errs := map[string]string{}
		switch err {
		case coreCategory.ErrEmptyName:
			errs["name"] = "Name is required"
		case coreCategory.ErrInvalidType:
			errs["type"] = "Please select a valid type"
		case coreCategory.ErrCircularParent:
			errs["parent_id"] = "A category cannot be its own parent"
		default:
			errs["name"] = err.Error()
		}
		all, _ := svc.List()
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, formTemplateFiles, "category-form", gin.H{
			"Category":         existing,
			"Errors":           errs,
			"AllCategories":    all,
			"SelectedParentID": selectedParentID(existing),
		})
		return
	}

	all, _ := svc.List()
	byID := indexByID(all)

	parentName := ""
	if existing.ParentID != nil {
		if p, ok := byID[*existing.ParentID]; ok {
			parentName = p.Name
		}
	}
	desc := existing.Description
	if len(desc) > 60 {
		desc = desc[:60] + "…"
	}

	shared.SetFlash(c, fmt.Sprintf("Category \"%s\" updated successfully", existing.Name))
	shared.RenderFragment(c, rowTemplateFiles, "category-row", gin.H{
		"ID":               existing.ID,
		"Name":             existing.Name,
		"Type":             existing.Type,
		"ShortDescription": desc,
		"ParentName":       parentName,
		"ParentID":         existing.ParentID,
	})
}

// DeleteCategory deletes a category and returns an empty fragment to remove the row.
func DeleteCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}

	svc := categoryService()
	if err := svc.Delete(uint(id)); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to delete category")
		return
	}

	shared.SetFlash(c, "Category deleted successfully")
	c.Header("Content-Type", "text/html")
	c.String(http.StatusOK, "")
}

// EditCancelCategory restores the original category row when edit is cancelled.
func EditCancelCategory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid category ID")
		return
	}
	svc := categoryService()
	cat, err := svc.GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Category not found")
		return
	}
	all, _ := svc.List()
	byID := indexByID(all)
	parentName := ""
	if cat.ParentID != nil {
		if p, ok := byID[*cat.ParentID]; ok {
			parentName = p.Name
		}
	}
	desc := cat.Description
	if len(desc) > 60 {
		desc = desc[:60] + "…"
	}
	shared.RenderFragment(c, rowTemplateFiles, "category-row", gin.H{
		"ID":               cat.ID,
		"Name":             cat.Name,
		"Type":             cat.Type,
		"ShortDescription": desc,
		"ParentName":       parentName,
		"ParentID":         cat.ParentID,
	})
}

// ConfirmDeleteCategory renders the confirm-delete fragment for a category row.
func ConfirmDeleteCategory(c *gin.Context) {
	id := c.Param("id")

	shared.RenderFragment(c, confirmDeleteTemplateFiles, "confirm-delete", gin.H{
		"DeleteURL": fmt.Sprintf("/categories/%s", id),
		"CancelURL": fmt.Sprintf("/categories/%s/confirm-delete-cancel", id),
		"TargetID":  fmt.Sprintf("category-row-%s-delete", id),
		"RowTarget": fmt.Sprintf("#category-row-%s", id),
	})
}

// ConfirmDeleteCategoryCancel restores the original delete trigger for a category row.
func ConfirmDeleteCategoryCancel(c *gin.Context) {
	id := c.Param("id")

	shared.RenderFragment(c, deleteTriggerTemplateFiles, "delete-trigger", gin.H{
		"ConfirmURL": fmt.Sprintf("/categories/%s/confirm-delete", id),
		"Target":     fmt.Sprintf("#category-row-%s-delete", id),
	})
}
