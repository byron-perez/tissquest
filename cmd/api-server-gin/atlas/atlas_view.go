package atlas

import (
	"html/template"
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// TissueRecordCard is the view model for a tissue record card in the atlas view.
// It carries the resolved thumbnail URL so the template stays simple.
type TissueRecordCard struct {
	tissuerecord.TissueRecord
	ThumbUrl string // best available low-res image for the first slide
}

type AtlasViewData struct {
	Atlas         atlas.Atlas
	TissueRecords []TissueRecordCard
}

func renderError(c *gin.Context, status int, message string) {
	tmpl := template.Must(template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/error.html",
	))
	c.Header("Content-Type", "text/html")
	c.Writer.WriteHeader(status)
	if err := tmpl.ExecuteTemplate(c.Writer, "error", gin.H{"error": message}); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func ViewAtlas(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		renderError(c, http.StatusBadRequest, "Invalid atlas ID")
		return
	}

	atlasService := services.NewAtlasService(repositories.NewAtlasRepository())
	atlasData, err := atlasService.GetAtlas(uint(id))
	if err != nil {
		renderError(c, http.StatusNotFound, "Atlas not found")
		return
	}

	trService := services.NewTissueRecordService(repositories.NewTissueRecordRepository())
	slideService := services.NewSlideService(nil, repositories.NewSlideRepository())

	var cards []TissueRecordCard
	for _, recordID := range atlasData.TissueRecords {
		if record, status := trService.GetByID(recordID); status != 0 {
			card := TissueRecordCard{TissueRecord: record}
			// Resolve the low-res thumbnail for the first slide
			displaySlides, err := slideService.ListDisplayByTissueRecord(record.ID, slide.ImageSizeThumb)
			if err == nil && len(displaySlides) > 0 {
				card.ThumbUrl = displaySlides[0].ImageUrl
			}
			cards = append(cards, card)
		}
	}

	data := gin.H{
		"Title": atlasData.Name,
		"Crumbs": []map[string]string{
			{"Label": "Home", "URL": "/"},
			{"Label": "Atlases", "URL": "/atlases"},
			{"Label": atlasData.Name},
		},
		"data": AtlasViewData{
			Atlas:         *atlasData,
			TissueRecords: cards,
		},
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/atlas_view.html",
		"web/templates/includes/main-menu.html",
		"web/templates/includes/breadcrumb.html",
	))
	c.Header("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(c.Writer, "base", data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
