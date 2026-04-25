package slides

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcba/tissquest/cmd/api-server-gin/shared"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

// dziViewModel is the data passed to the viewer template.
type dziViewModel struct {
	DziURL            string
	BaseMagnification int
	// MicronsPerPixelJS is pre-formatted for safe inline JS (always uses "." decimal separator).
	MicronsPerPixelJS string
	HomeViewportJSON  string
}

func buildDziViewModel(sl *slide.Slide) dziViewModel {
	vpJSON := "null"
	if sl.HomeViewport != nil {
		if b, err := json.Marshal(sl.HomeViewport); err == nil {
			vpJSON = string(b)
		}
	}
	mppJS := strconv.FormatFloat(sl.MicronsPerPixel, 'f', 6, 64)
	return dziViewModel{
		DziURL:            sl.DziURL,
		BaseMagnification: sl.BaseMagnification,
		MicronsPerPixelJS: mppJS,
		HomeViewportJSON:  vpJSON,
	}
}

var viewerTemplateFiles = []string{
	"web/templates/pages/slide_viewer.html",
}

// ViewSlide renders the full-screen virtual microscope viewer for a slide.
// GET /slides/:id/viewer
func ViewSlide(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		shared.RenderError(c, http.StatusBadRequest, "invalid slide id")
		return
	}

	svc := services.NewSlideService(nil, repositories.NewSlideRepository())
	sl, err := svc.GetByID(uint(id))
	if err != nil {
		shared.RenderError(c, http.StatusNotFound, "slide not found")
		return
	}

	// Resolve static image URL for the fallback path
	staticURL := ""
	displaySlides, _ := svc.ListDisplayByTissueRecord(sl.TissueRecordID, slide.ImageSizePreview)
	for _, ds := range displaySlides {
		if ds.ID == sl.ID {
			staticURL = ds.ImageUrl
			break
		}
	}

	shared.RenderPage(c, viewerTemplateFiles, "content", gin.H{
		"Title":          sl.Name + " — Virtual Microscope",
		"Slide":          sl,
		"DziMetadata":    buildDziViewModel(sl),
		"StaticImageURL": staticURL,
		"Objectives":     []int{4, 10, 40},
	})
}
