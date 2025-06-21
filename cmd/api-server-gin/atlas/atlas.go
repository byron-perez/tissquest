package atlas

import (
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/persistence/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func ListAtlases(c *gin.Context) {
	repo := repositories.NewGormAtlasRepository(c.MustGet("db").(*gorm.DB))
	atlases, err := repo.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.HTML(http.StatusOK, "main-menu.html", gin.H{
		"title":   "Atlas List",
		"Atlases": atlases,
	})
}

func CreateAtlas(c *gin.Context) {
	var newAtlas atlas.Atlas
	if err := c.ShouldBindJSON(&newAtlas); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	repo := repositories.NewGormAtlasRepository(c.MustGet("db").(*gorm.DB))
	id := repo.Save(&newAtlas)

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func GetAtlas(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	repo := repositories.NewGormAtlasRepository(c.MustGet("db").(*gorm.DB))
	atlas, err := repo.Retrieve(uint(id))
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

	repo := repositories.NewGormAtlasRepository(c.MustGet("db").(*gorm.DB))
	err = repo.Update(uint(id), &updatedAtlas)
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

	repo := repositories.NewGormAtlasRepository(c.MustGet("db").(*gorm.DB))
	err = repo.Delete(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Atlas deleted successfully"})
}
