package index

import (
	"html/template"
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetIndex(c *gin.Context) {
	atlases, err := fetchAtlases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch atlases"})
		return
	}

	var featured *atlas.Atlas
	if len(atlases) > 0 {
		featured = &atlases[0]
	}

	data := gin.H{
		"Title":         "Tissquest",
		"Atlases":       atlases,
		"FeaturedAtlas": featured,
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/index.html",
		"web/templates/includes/main-menu.html",
	))
	c.Header("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(c.Writer, "base", data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func fetchAtlases() ([]atlas.Atlas, error) {
	repo := repositories.NewAtlasRepository()
	service := services.NewAtlasService(repo)
	return service.ListAtlases()
}
