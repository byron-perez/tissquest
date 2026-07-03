package slides

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"mcba/tissquest/internal/persistence/migration"
	"mcba/tissquest/internal/persistence/repositories"
)

// ListAnnotations returns all annotations for a slide as a JSON array.
// GET /api/slides/:id/annotations
// Annotorious calls this on viewer init to load existing annotations.
func ListAnnotations(c *gin.Context) {
	slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slide id"})
		return
	}

	db, err := repositories.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	var models []migration.AnnotationModel
	if err := db.Where("slide_id = ?", slideID).Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return the raw annotation JSON objects as a JSON array
	annotations := make([]json.RawMessage, 0, len(models))
	for _, m := range models {
		annotations = append(annotations, json.RawMessage(m.AnnotationJSON))
	}
	c.JSON(http.StatusOK, annotations)
}

// CreateAnnotation persists a new annotation created by Annotorious.
// POST /api/slides/:id/annotations
func CreateAnnotation(c *gin.Context) {
	slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slide id"})
		return
	}

	var body json.RawMessage
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid annotation body"})
		return
	}

	// Extract the Annotorious ID from the JSON
	var meta struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &meta); err != nil || meta.ID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "annotation must have an id field"})
		return
	}

	db, err := repositories.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	model := migration.AnnotationModel{
		SlideID:        uint(slideID),
		AnnotoriousID:  meta.ID,
		AnnotationJSON: string(body),
	}
	if err := db.Create(&model).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, body)
}

// UpdateAnnotation replaces an existing annotation (Annotorious fires this on edit).
// PUT /api/slides/:id/annotations/:annotationID
func UpdateAnnotation(c *gin.Context) {
	slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slide id"})
		return
	}
	annotoriousID := c.Param("annotationID")

	var body json.RawMessage
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid annotation body"})
		return
	}

	db, err := repositories.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if err := db.Model(&migration.AnnotationModel{}).
		Where("slide_id = ? AND annotorious_id = ?", slideID, annotoriousID).
		Update("annotation_json", string(body)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, body)
}

// DeleteAnnotation removes an annotation (Annotorious fires this on delete).
// DELETE /api/slides/:id/annotations/:annotationID
func DeleteAnnotation(c *gin.Context) {
	slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slide id"})
		return
	}
	annotoriousID := c.Param("annotationID")

	db, err := repositories.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	if err := db.Where("slide_id = ? AND annotorious_id = ?", slideID, annotoriousID).
		Delete(&migration.AnnotationModel{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

// BatchSaveRequest is the payload for the batch save endpoint.
type BatchSaveRequest struct {
	Created    []json.RawMessage `json:"created"`
	Updated    []json.RawMessage `json:"updated"`
	DeletedIDs []string          `json:"deleted_ids"`
}

// BatchSaveAnnotations applies all pending annotation changes in a single DB transaction.
// POST /api/slides/:id/annotations/batch
func BatchSaveAnnotations(c *gin.Context) {
	slideID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid slide id"})
		return
	}

	var req BatchSaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	// Validate all created entries have an id before entering the transaction
	type idOnly struct {
		ID string `json:"id"`
	}
	for i, raw := range req.Created {
		var meta idOnly
		if err := json.Unmarshal(raw, &meta); err != nil || meta.ID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "created[" + strconv.Itoa(i) + "] must have an id field"})
			return
		}
	}

	db, err := repositories.GetDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "db error"})
		return
	}

	txErr := db.Transaction(func(tx *gorm.DB) error {
		// INSERT new annotations
		for _, raw := range req.Created {
			var meta idOnly
			json.Unmarshal(raw, &meta) // already validated above
			model := migration.AnnotationModel{
				SlideID:        uint(slideID),
				AnnotoriousID:  meta.ID,
				AnnotationJSON: string(raw),
			}
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		}

		// UPDATE existing annotations
		for _, raw := range req.Updated {
			var meta idOnly
			if err := json.Unmarshal(raw, &meta); err != nil || meta.ID == "" {
				continue // skip malformed entries; log warning
			}
			result := tx.Model(&migration.AnnotationModel{}).
				Where("slide_id = ? AND annotorious_id = ?", uint(slideID), meta.ID).
				Update("annotation_json", string(raw))
			if result.Error != nil {
				return result.Error
			}
			// RowsAffected == 0 is a no-op (unknown ID) — acceptable per design
		}

		// SOFT-DELETE by annotorious_id
		for _, aid := range req.DeletedIDs {
			result := tx.Where("slide_id = ? AND annotorious_id = ?", uint(slideID), aid).
				Delete(&migration.AnnotationModel{})
			if result.Error != nil {
				return result.Error
			}
		}

		return nil
	})

	if txErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": txErr.Error()})
		return
	}

	// Return the full reconciled annotation list so the client can commitDiff
	var models []migration.AnnotationModel
	if err := db.Where("slide_id = ?", slideID).Find(&models).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load annotations after save"})
		return
	}
	result := make([]json.RawMessage, 0, len(models))
	for _, m := range models {
		result = append(result, json.RawMessage(m.AnnotationJSON))
	}
	c.JSON(http.StatusOK, result)
}
