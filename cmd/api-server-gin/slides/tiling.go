package slides

import (
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"mcba/tissquest/cmd/api-server-gin/shared"
	"mcba/tissquest/cmd/tile-pipeline/tiling"
	"mcba/tissquest/internal/persistence/repositories"
	persistencestorage "mcba/tissquest/internal/persistence/storage"
	"mcba/tissquest/internal/services"
)

// TileSlide runs the tiling pipeline for a single slide and refreshes the gallery card.
// POST /slides/:id/tile
// Expects the S3 storage to be injected (same instance as the main server uses).
func TileSlide(s3 *persistencestorage.S3Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
		if err != nil {
			shared.RenderError(c, http.StatusBadRequest, "Invalid slide id")
			return
		}

		svc := services.NewSlideService(s3, repositories.NewSlideRepository())
		sl, err := svc.GetByID(uint(slideID))
		if err != nil {
			shared.RenderError(c, http.StatusNotFound, "Slide not found")
			return
		}

		if sl.ImageKey == "" {
			shared.RenderError(c, http.StatusBadRequest, "Slide has no source image yet — upload an image first")
			return
		}

		// Use the slide's own magnification as base; default mpp to a sensible value
		// if not yet set (can be updated later via the CLI with a precise value).
		baseMag := sl.Magnification
		if baseMag == 0 {
			baseMag = 40
		}
		mpp := sl.MicronsPerPixel
		if mpp == 0 {
			mpp = 0.25 // reasonable default for a 40× objective
		}

		dziURL, err := tiling.Run(tiling.Params{
			SlideID:           uint(slideID),
			BaseMagnification: baseMag,
			MicronsPerPixel:   mpp,
			S3:                s3,
			Region:            os.Getenv("AWS_REGION"),
			Bucket:            os.Getenv("S3_BUCKET"),
		})
		if err != nil {
			shared.RenderError(c, http.StatusInternalServerError, "Tiling failed: "+err.Error())
			return
		}

		if err := svc.SetDziMetadata(uint(slideID), dziURL, baseMag, mpp); err != nil {
			shared.RenderError(c, http.StatusInternalServerError, "Failed to save DZI metadata: "+err.Error())
			return
		}

		shared.SetFlash(c, "Slide tiled successfully — viewer is now available")
		renderGallery(c, sl.TissueRecordID)
	}
}
