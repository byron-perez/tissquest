package slides

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcba/tissquest/cmd/api-server-gin/shared"
	"mcba/tissquest/internal/core/slide"
	corestorage "mcba/tissquest/internal/core/storage"
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

	svc := services.NewSlideService(nil, repositories.NewSlideRepository())
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
		"home_viewport":      sl.HomeViewport, // nil is serialised as JSON null
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

	svc := services.NewSlideService(nil, repositories.NewSlideRepository())
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

// SetImageVariant is called by the Lambda function after it generates a size variant.
// PATCH /slides/:id/images/:size  body: { "url": "https://..." }
func SetImageVariant(c *gin.Context) {
	slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slide id"})
		return
	}

	size := slide.ImageSize(c.Param("size"))
	switch size {
	case slide.ImageSizeOriginal, slide.ImageSizeThumb, slide.ImageSizePreview:
		// valid
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid size, use: original, low, medium"})
		return
	}

	var body struct {
		Url string `json:"url" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url is required"})
		return
	}

	svc := services.NewSlideService(nil, repositories.NewSlideRepository())
	if err := svc.SetImageVariant(uint(slideID), size, body.Url); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func UploadSlideImage(storage corestorage.ImageStorage) gin.HandlerFunc {
	return func(c *gin.Context) {
		slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			shared.RenderError(c, http.StatusBadRequest, "Invalid slide id")
			return
		}

		file, header, err := c.Request.FormFile("image")
		if err != nil {
			shared.RenderError(c, http.StatusBadRequest, "Image file required")
			return
		}
		defer file.Close()

		svc := services.NewSlideService(storage, repositories.NewSlideRepository())
		if _, err := svc.UploadImage(uint(slideID), file, header); err != nil {
			shared.RenderError(c, http.StatusInternalServerError, err.Error())
			return
		}

		sl, err := svc.GetByID(uint(slideID))
		if err != nil {
			shared.RenderError(c, http.StatusInternalServerError, "Slide not found after upload")
			return
		}

		renderGallery(c, sl.TissueRecordID)
	}
}
