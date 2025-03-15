package tissuerecords

import (
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/repositories"
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

func GetTissueRecordById(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	parsedId := uint(id)

	if err != nil {
		// ... handle error TODO
		panic(err)
	}

	gorm_repository := repositories.NewGormTissueRecordRepository()
	tissrecord := &tissuerecord.TissueRecord{}
	tissrecord.ConfigureTissueRecord(gorm_repository)

	foundTissueRecord, status_code := tissrecord.GetById(parsedId)

	if status_code == 0 {
		panic("Not found record")
	}

	c.IndentedJSON(http.StatusOK, foundTissueRecord)
}

func CreateTissueRecord(c *gin.Context) {
	var body TissueRecordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	tissrecord := tissuerecord.TissueRecord{
		Name:           body.Name,
		Notes:          body.Notes,
		Taxonomicclass: body.Taxonomicclass,
		Slides:         []slide.Slide{},
	}

	gorm_repository := repositories.NewGormTissueRecordRepository()
	tissrecord.ConfigureTissueRecord(gorm_repository)

	newRecordId := tissrecord.Save()

	c.IndentedJSON(http.StatusOK, newRecordId)
}

func UpdateTissueRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	parsedId := uint(id)

	if err != nil {
		// ... handle error TODO
		panic(err)
	}
	var body TissueRecordBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.Error(err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// map fields
	tissrecordForUpdate := tissuerecord.TissueRecord{
		Name:           body.Name,
		Notes:          body.Notes,
		Taxonomicclass: body.Taxonomicclass,
		Slides:         []slide.Slide{},
	}

	tissrecord := tissuerecord.TissueRecord{}
	gorm_repository := repositories.NewGormTissueRecordRepository()
	tissrecord.ConfigureTissueRecord(gorm_repository)

	tissrecord.Update(parsedId, tissrecordForUpdate)

	c.IndentedJSON(http.StatusOK, tissrecordForUpdate)
}

func DeleteTissueRecord(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	parsedId := uint(id)

	if err != nil {
		// ... handle error TODO
		panic(err)
	}

	tissrecord := tissuerecord.TissueRecord{}
	gorm_repository := repositories.NewGormTissueRecordRepository()
	tissrecord.ConfigureTissueRecord(gorm_repository)

	tissrecord.Delete(parsedId)

	c.IndentedJSON(http.StatusOK, tissrecord)
}
