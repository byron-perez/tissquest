package slides

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcba/tissquest/cmd/api-server-gin/shared"
	"mcba/tissquest/internal/core/slide"
	corestorage "mcba/tissquest/internal/core/storage"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

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
