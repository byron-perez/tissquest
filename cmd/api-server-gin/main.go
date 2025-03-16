package main

import (
	"fmt"
	"log"
	"mcba/tissquest/cmd/api-server-gin/index"
	tissuerecords "mcba/tissquest/cmd/api-server-gin/tissue_records"
	"mcba/tissquest/internal/persistence/migration"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func setupRouter() *gin.Engine {
	gin.DisableConsoleColor()
	router := gin.Default()

	router.Static("/static", "web/static")
	router.LoadHTMLGlob("web/templates/*")

	router.GET("/", index.GetIndex)

	router.GET("/tissue_records/:id", tissuerecords.GetTissueRecordById)
	router.POST("/tissue_records", tissuerecords.CreateTissueRecord)
	router.PUT("/tissue_records/:id", tissuerecords.UpdateTissueRecord)
	router.DELETE("/tissue_records/:id", tissuerecords.DeleteTissueRecord)

	return router
}

func main() {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Walk(ex, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		fmt.Printf("dir: %v: name: %s\n", info.IsDir(), path)
		return nil
	})
	fmt.Println(exPath)
	// load .env
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// setup database
	migration.RunMigration()

	r := setupRouter()
	// Listen and Server in 0.0.0.0:8080
	r.Run(":8080")
}
