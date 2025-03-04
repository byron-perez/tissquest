package main

import (
	"log"
	"mcba/tissquest/cmd/api-server-gin/index"
	"mcba/tissquest/internal/core/slide"
	"mcba/tissquest/internal/core/tissuerecord"
	"mcba/tissquest/internal/persistence/migration"
	"mcba/tissquest/internal/persistence/repositories"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

func getTissueRecordById(c *gin.Context) {
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

func createTissueRecord(c *gin.Context) {
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

func updateTissueRecord(c *gin.Context) {
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

func deleteTissueRecord(c *gin.Context) {
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

func setupRouter() *gin.Engine {
	gin.DisableConsoleColor()
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")

	router.GET("/", index.GetIndex)

	router.GET("/tissue_records/:id", getTissueRecordById)
	router.POST("/tissue_records", createTissueRecord)
	router.PUT("/tissue_records/:id", updateTissueRecord)
	router.DELETE("/tissue_records/:id", deleteTissueRecord)

	return router
}

func main() {
	// load .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// setup database
	migration.RunMigration()

	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8000")
}
