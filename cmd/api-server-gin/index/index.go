package index

import (
	"fmt"
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
	//print atlases for debugging purposes
	fmt.Println(atlases)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch atlases"})
		return
	}

	c.HTML(http.StatusOK, "base.html", gin.H{
		"title":   "Tissquest",
		"Atlases": atlases,
	})
}

func fetchAtlases() ([]atlas.Atlas, error) {
	repo := repositories.NewPostgresAtlasRepository()
	service := services.NewAtlasService(repo)
	return service.ListAtlases()
}
