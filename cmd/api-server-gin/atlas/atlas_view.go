package atlas

import (
	"mcba/tissquest/internal/core/atlas"
	"mcba/tissquest/internal/core/category"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AtlasViewData struct {
	Atlas      atlas.Atlas
	Categories map[string][]CategoryWithRecords
}

type CategoryWithRecords struct {
	Category      category.Category
	TissueRecords []tissuerecord.TissueRecord
}

func ViewAtlas(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid atlas ID"})
		return
	}

	atlasService := services.NewAtlasService(repositories.NewAtlasRepository())
	atlasData, err := atlasService.GetAtlas(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Atlas not found"})
		return
	}

	trService := services.NewTissueRecordService(repositories.NewTissueRecordRepository())
	categories, _ := repositories.NewMemoryCategoryRepository().List()
	categorizedData := make(map[string][]CategoryWithRecords)

	// Create a set of tissue record IDs in this atlas for quick lookup
	atlasRecordIDs := make(map[uint]bool)
	for _, recordID := range atlasData.TissueRecords {
		atlasRecordIDs[recordID] = true
	}

	for _, cat := range categories {
		var tissueRecords []tissuerecord.TissueRecord
		for _, recordID := range cat.TissueRecordIDs {
			if atlasRecordIDs[recordID] { // Only include records that belong to this atlas
				if record, status := trService.GetByID(recordID); status != 0 {
					tissueRecords = append(tissueRecords, record)
				}
			}
		}
		if len(tissueRecords) > 0 { // Only add categories that have records in this atlas
			categoryType := string(cat.Type)
			categorizedData[categoryType] = append(categorizedData[categoryType], CategoryWithRecords{
				Category:      cat,
				TissueRecords: tissueRecords,
			})
		}
	}

	c.HTML(http.StatusOK, "includes/atlas_view.html", gin.H{
		"title": atlasData.Name,
		"data": AtlasViewData{
			Atlas:      *atlasData,
			Categories: categorizedData,
		},
	})
}
