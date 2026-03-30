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

	atlasService := services.NewAtlasService(repositories.NewPostgresAtlasRepository())
	atlasData, err := atlasService.GetAtlas(uint(id))
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Atlas not found"})
		return
	}

	trService := services.NewTissueRecordService(repositories.NewGormTissueRecordRepository())
	categories, _ := repositories.NewMemoryCategoryRepository().List()
	categorizedData := make(map[string][]CategoryWithRecords)

	for _, cat := range categories {
		var tissueRecords []tissuerecord.TissueRecord
		for _, recordID := range cat.TissueRecordIDs {
			if record, status := trService.GetByID(recordID); status != 0 {
				tissueRecords = append(tissueRecords, record)
			}
		}
		categoryType := string(cat.Type)
		categorizedData[categoryType] = append(categorizedData[categoryType], CategoryWithRecords{
			Category:      cat,
			TissueRecords: tissueRecords,
		})
	}

	c.HTML(http.StatusOK, "base.html", gin.H{
		"title": atlasData.Name,
		"data": AtlasViewData{
			Atlas:      *atlasData,
			Categories: categorizedData,
		},
	})
}
