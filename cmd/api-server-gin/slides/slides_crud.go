package slides

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"mcba/tissquest/cmd/api-server-gin/shared"
	coreSlide "mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

func trService() *services.TissueRecordService {
	return services.NewTissueRecordService(repositories.NewTissueRecordRepository())
}

var (
	slideGalleryTemplateFiles = []string{
		"web/templates/includes/slide_gallery.html",
	}
	slideSummaryTemplateFiles = []string{
		"web/templates/includes/workspace_summary.html",
	}
	slideFormTemplateFiles = []string{
		"web/templates/pages/slide_form.html",
	}
	slideConfirmDeleteTemplateFiles = []string{
		"web/templates/includes/confirm-delete.html",
	}
	slideDeleteTriggerTemplateFiles = []string{
		"web/templates/includes/delete-trigger.html",
	}
)

func slideService() *services.SlideService {
	return services.NewSlideService(nil, repositories.NewSlideRepository())
}

func renderGallery(c *gin.Context, tissueRecordID uint) {
	svc := slideService()
	slides, err := svc.ListDisplayByTissueRecord(tissueRecordID, coreSlide.ImageSizeThumb)
	if err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to load slides")
		return
	}
	shared.RenderFragment(c, slideGalleryTemplateFiles, "slide-gallery", gin.H{
		"Slides":         slides,
		"TissueRecordID": tissueRecordID,
	})
	// OOB: update the summary bar with the new slide count
	tr, _ := trService().GetByID(tissueRecordID)

	shared.AppendFragment(c, slideSummaryTemplateFiles, "workspace-summary", gin.H{
		"TissueRecord": tr,
		"SlideCount":   len(slides),
	})
}

// ListSlides renders the slide gallery fragment for a tissue record.
func ListSlides(c *gin.Context) {
	trID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid tissue record ID")
		return
	}
	renderGallery(c, uint(trID))
}

// NewSlideForm renders an empty slide form fragment.
func NewSlideForm(c *gin.Context) {
	trID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid tissue record ID")
		return
	}
	shared.RenderFragment(c, slideFormTemplateFiles, "slide-form", gin.H{
		"Slide":          nil,
		"TissueRecordID": uint(trID),
		"Errors":         map[string]string{},
	})
}

// CreateSlide handles form submission to create a new slide for a tissue record.
func CreateSlide(c *gin.Context) {
	trID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid tissue record ID")
		return
	}

	mag, _ := strconv.Atoi(c.PostForm("magnification"))

	sl := &coreSlide.Slide{
		Name:          c.PostForm("name"),
		Magnification: mag,
		Preparation: coreSlide.Preparation{
			Staining:        c.PostForm("staining"),
			InclusionMethod: c.PostForm("inclusion_method"),
			Reagents:        c.PostForm("reagents"),
			Protocol:        c.PostForm("protocol"),
			Notes:           c.PostForm("notes"),
		},
	}

	svc := slideService()
	_, createErr := svc.Create(uint(trID), sl)
	if createErr != nil {
		errors := map[string]string{}
		switch createErr {
		case coreSlide.ErrEmptyName:
			errors["name"] = "Name is required"
		case coreSlide.ErrInvalidMagnification:
			errors["magnification"] = "Magnification must be a positive number"
		default:
			errors["name"] = createErr.Error()
		}
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, slideFormTemplateFiles, "slide-form", gin.H{
			"Slide":          sl,
			"TissueRecordID": uint(trID),
			"Errors":         errors,
		})
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Slide \"%s\" created successfully", sl.Name))
	renderGallery(c, uint(trID))
}

// EditSlideForm renders a pre-populated slide form fragment for editing.
func EditSlideForm(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid slide ID")
		return
	}

	svc := slideService()
	sl, err := svc.GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Slide not found")
		return
	}

	shared.RenderFragment(c, slideFormTemplateFiles, "slide-form", gin.H{
		"Slide":          sl,
		"TissueRecordID": sl.TissueRecordID,
		"Errors":         map[string]string{},
	})
}

// UpdateSlide handles form submission to update an existing slide.
func UpdateSlide(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid slide ID")
		return
	}

	svc := slideService()
	existing, err := svc.GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "Slide not found")
		return
	}

	mag, _ := strconv.Atoi(c.PostForm("magnification"))

	updated := &coreSlide.Slide{
		Name:          c.PostForm("name"),
		Magnification: mag,
		Preparation: coreSlide.Preparation{
			Staining:        c.PostForm("staining"),
			InclusionMethod: c.PostForm("inclusion_method"),
			Reagents:        c.PostForm("reagents"),
			Protocol:        c.PostForm("protocol"),
			Notes:           c.PostForm("notes"),
		},
	}

	if updateErr := svc.Update(uint(id), updated); updateErr != nil {
		errors := map[string]string{}
		switch updateErr {
		case coreSlide.ErrEmptyName:
			errors["name"] = "Name is required"
		case coreSlide.ErrInvalidMagnification:
			errors["magnification"] = "Magnification must be a positive number"
		default:
			errors["name"] = updateErr.Error()
		}
		c.Status(http.StatusUnprocessableEntity)
		shared.RenderFragment(c, slideFormTemplateFiles, "slide-form", gin.H{
			"Slide":          existing,
			"TissueRecordID": existing.TissueRecordID,
			"Errors":         errors,
		})
		return
	}

	shared.SetFlash(c, fmt.Sprintf("Slide \"%s\" updated successfully", updated.Name))
	renderGallery(c, existing.TissueRecordID)
}

// DeleteSlide deletes a slide and returns an empty response so HTMX removes the card.
func DeleteSlide(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "Invalid slide ID")
		return
	}

	svc := slideService()
	if _, err := svc.GetByID(uint(id)); err != nil {
		shared.RenderError(c, http.StatusNotFound, "Slide not found")
		return
	}

	if err := svc.Delete(uint(id)); err != nil {
		shared.RenderError(c, http.StatusInternalServerError, "Failed to delete slide")
		return
	}

	shared.SetFlash(c, "Slide deleted successfully")
	c.Status(http.StatusOK) // empty body — HTMX replaces the card with nothing
}

// ConfirmDeleteSlide renders the confirm-delete fragment for a slide.
func ConfirmDeleteSlide(c *gin.Context) {
	id := c.Param("id")
	shared.RenderFragment(c, slideConfirmDeleteTemplateFiles, "confirm-delete", gin.H{
		"DeleteURL": fmt.Sprintf("/slides/%s", id),
		"CancelURL": fmt.Sprintf("/slides/%s/confirm-delete-cancel", id),
		"TargetID":  fmt.Sprintf("slide-%s-delete", id),
		"RowTarget": fmt.Sprintf("#slide-card-%s", id),
		"RowSwap":   "outerHTML",
	})
}

// ConfirmDeleteSlideCancel restores the original delete trigger for a slide.
func ConfirmDeleteSlideCancel(c *gin.Context) {
	id := c.Param("id")
	shared.RenderFragment(c, slideDeleteTriggerTemplateFiles, "delete-trigger", gin.H{
		"ConfirmURL": fmt.Sprintf("/slides/%s/confirm-delete", id),
		"Target":     fmt.Sprintf("#slide-%s-delete", id),
	})
}
