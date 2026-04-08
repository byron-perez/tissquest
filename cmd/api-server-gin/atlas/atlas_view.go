package atlas

import (
	"html/template"
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AtlasViewData struct {
	Atlas         atlas.Atlas
	TissueRecords []tissuerecord.TissueRecord
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
	var tissueRecords []tissuerecord.TissueRecord
	for _, recordID := range atlasData.TissueRecords {
		if record, status := trService.GetByID(recordID); status != 0 {
			tissueRecords = append(tissueRecords, record)
		}
	}

	data := gin.H{
		"Title": atlasData.Name,
		"data": AtlasViewData{
			Atlas:         *atlasData,
			TissueRecords: tissueRecords,
		},
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/atlas_view.html",
	))
	c.Header("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(c.Writer, "base", data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
