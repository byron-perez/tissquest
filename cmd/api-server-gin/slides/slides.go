package slides

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	corestorage "mcba/tissquest/internal/core/storage"
	"mcba/tissquest/internal/services"
)

func UploadSlideImage(storage corestorage.ImageStorage) gin.HandlerFunc {
	svc := services.NewSlideService(storage, nil)
	return func(c *gin.Context) {
		slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slide id"})
			return
		}

		file, header, err := c.Request.FormFile("image")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "image file required"})
			return
		}
		defer file.Close()

		url, err := svc.UploadImage(uint(slideID), file, header)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"url": url})
	}
}
