package index

import (
	"fmt"
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/persistence/repositories"
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
		"title":   "Welcome to TissQuest",
		"Atlases": atlases,
	})
}

func fetchAtlases() ([]atlas.Atlas, error) {
	repo := repositories.NewPostgresAtlasRepository()
	atlas := atlas.Atlas{}
	atlas.ConfigureAtlas(repo)
	return repo.List()
}
