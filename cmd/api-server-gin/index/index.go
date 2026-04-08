package index

import (
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

	c.HTML(http.StatusOK, "base.html", gin.H{
		"title":         "Tissquest",
		"Atlases":       atlases,
		"FeaturedAtlas": featured,
	})
}

func fetchAtlases() ([]atlas.Atlas, error) {
	repo := repositories.NewGormAtlasRepository()
	service := services.NewAtlasService(repo)
	return service.ListAtlases()
}
