package slides

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

// GetDziMetadata returns the viewer-initialization data for a tiled slide.
// GET /api/slides/:id/dzi
func GetDziMetadata(c *gin.Context) {
	slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slide id"})
		return
	}

	svc := services.NewSlideService(repositories.NewSlideRepository())
	sl, err := svc.GetByID(uint(slideID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "slide not found"})
		return
	}

	if !sl.IsTiled() {
		c.JSON(http.StatusNotFound, gin.H{"error": "slide has not been tiled yet"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"dzi_url":            sl.DziURL,
		"base_magnification": sl.BaseMagnification,
		"microns_per_pixel":  sl.MicronsPerPixel,
		"tile_size":          256,
		"home_viewport":      sl.HomeViewport,
	})
}

// SetHomeViewport saves the current viewport position as the curated starting view.
// PATCH /api/slides/:id/home-viewport  body: {"x": float, "y": float, "zoom": float}
func SetHomeViewport(c *gin.Context) {
	slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slide id"})
		return
	}

	var vp slide.ViewportPosition
	if err := json.NewDecoder(c.Request.Body).Decode(&vp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body must be {x, y, zoom}"})
		return
	}

	svc := services.NewSlideService(repositories.NewSlideRepository())
	sl, err := svc.GetByID(uint(slideID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "slide not found"})
		return
	}

	sl.HomeViewport = &vp
	if err := svc.Update(uint(slideID), sl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
