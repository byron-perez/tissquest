package main

import (
	"log"
	"mcba/tissquest/cmd/api-server-gin/index"
	tissuerecords "mcba/tissquest/cmd/api-server-gin/tissue_records"
	"mcba/tissquest/internal/persistence/migration"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	gin.DisableConsoleColor()
	router := gin.Default()
	router.LoadHTMLGlob("web/templates/*")

	router.GET("/", index.GetIndex)

	router.GET("/tissue_records/:id", tissuerecords.GetTissueRecordById)
	router.POST("/tissue_records", tissuerecords.CreateTissueRecord)
	router.PUT("/tissue_records/:id", tissuerecords.UpdateTissueRecord)
	router.DELETE("/tissue_records/:id", tissuerecords.DeleteTissueRecord)

	// include static folder
	router.Static("/static", "web/static")

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
	r.Run("localhost:8000")
}
