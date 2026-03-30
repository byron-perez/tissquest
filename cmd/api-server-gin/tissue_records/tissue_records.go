package tissue_records

import (
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TissueRecordBody struct {
	Name           string `json:"name" binding:"required"`
	Notes          string `json:"notes" binding:"required"`
	Taxonomicclass string `json:"taxonomic_class" binding:"required"`
}

type TissueRecordBodyUpdate struct {
	Name           string `json:"name" binding:"required"`
	Notes          string `json:"notes" binding:"required"`
	Taxonomicclass string `json:"taxonomic_class" binding:"required"`
}

func newService() *services.TissueRecordService {
	return services.NewTissueRecordService(repositories.NewGormTissueRecordRepository())
}

func GetTissueRecordById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	record, status := newService().GetByID(uint(id))
	if status == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	c.IndentedJSON(http.StatusOK, record)
}

func CreateTissueRecord(c *gin.Context) {
	var body TissueRecordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	tr := tissuerecord.TissueRecord{
		Name:           body.Name,
		Notes:          body.Notes,
		Taxonomicclass: body.Taxonomicclass,
		Slides:         []slide.Slide{},
	}

	newID := newService().Create(&tr)
	c.IndentedJSON(http.StatusOK, newID)
}

func UpdateTissueRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var body TissueRecordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	tr := tissuerecord.TissueRecord{
		Name:           body.Name,
		Notes:          body.Notes,
		Taxonomicclass: body.Taxonomicclass,
		Slides:         []slide.Slide{},
	}

	newService().Update(uint(id), &tr)
	c.IndentedJSON(http.StatusOK, tr)
}

func DeleteTissueRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	newService().Delete(uint(id))
	c.Status(http.StatusNoContent)
}

func ListTissueRecords(c *gin.Context) {
	limit, page := 10, 1

	if v, err := strconv.Atoi(c.Query("limit")); err == nil {
		limit = v
	}
	if v, err := strconv.Atoi(c.Query("page")); err == nil {
		page = v
	}

	records, total, err := newService().List(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve tissue records"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  records,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
