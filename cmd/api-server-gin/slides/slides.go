package slides

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcba/tissquest/cmd/api-server-gin/shared"
	corestorage "mcba/tissquest/internal/core/storage"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

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
		url, err := svc.UploadImage(uint(slideID), file, header)
		if err != nil {
			shared.RenderError(c, http.StatusInternalServerError, err.Error())
			return
		}

		// Update the slide URL in the database
		sl, err := svc.GetByID(uint(slideID))
		if err != nil {
			shared.RenderError(c, http.StatusInternalServerError, "Slide not found after upload")
			return
		}
		sl.Url = url
		if err := svc.Update(uint(slideID), sl); err != nil {
			shared.RenderError(c, http.StatusInternalServerError, "Failed to save image URL")
			return
		}

		// Return refreshed gallery fragment
		renderGallery(c, sl.TissueRecordID)
	}
}
