package atlas

import (
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ListAtlases(c *gin.Context) {
	repo := repositories.NewPostgresAtlasRepository()
	service := services.NewAtlasService(repo)
	atlases, err := service.ListAtlases()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"atlases": atlases,
	})
}

func CreateAtlas(c *gin.Context) {
	var newAtlas atlas.Atlas
	if err := c.ShouldBindJSON(&newAtlas); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repo := repositories.NewPostgresAtlasRepository()
	service := services.NewAtlasService(repo)
	id, err := service.CreateAtlas(&newAtlas)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetAtlas(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	repo := repositories.NewPostgresAtlasRepository()
	service := services.NewAtlasService(repo)
	atlas, err := service.GetAtlas(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Atlas not found"})
		return
	}

	c.JSON(http.StatusOK, atlas)
}

func UpdateAtlas(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var updatedAtlas atlas.Atlas
	if err := c.ShouldBindJSON(&updatedAtlas); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repo := repositories.NewPostgresAtlasRepository()
	service := services.NewAtlasService(repo)
	err = service.UpdateAtlas(uint(id), &updatedAtlas)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Atlas updated successfully"})
}

func DeleteAtlas(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	repo := repositories.NewPostgresAtlasRepository()
	service := services.NewAtlasService(repo)
	err = service.DeleteAtlas(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Atlas deleted successfully"})
}
