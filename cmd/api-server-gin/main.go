package main

import (
	"log"
	"mcba/tissquest/cmd/api-server-gin/index"
	tissuerecords "mcba/tissquest/cmd/api-server-gin/tissue_records"
	"mcba/tissquest/internal/persistence/migration"
	"path/filepath"
	"text/template"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var templates map[string]*template.Template

type TemplateConfig struct {
	TemplateLayoutPath  string
	TemplateIncludePath string
}

func loadTemplates(templatesDir string) multitemplate.Renderer {

	r := multitemplate.NewRenderer()

	layouts, err := filepath.Glob(templatesDir + "/layouts/*.html")

	if err != nil {
		panic(err.Error())
	}

	includes, err := filepath.Glob(templatesDir + "/includes/*.html")
	if err != nil {
		panic(err.Error())
	}

	// Generate our templates map from our layouts/ and includes/ directories
	for _, include := range includes {
		layoutCopy := make([]string, len(layouts))
		copy(layoutCopy, layouts)
		files := append(layoutCopy, include)
		r.AddFromFiles(filepath.Base(include), files...)
	}
	return r
}

func setupRouter() *gin.Engine {
	gin.DisableConsoleColor()
	router := gin.Default()
	router.HTMLRender = loadTemplates("web/templates")

	router.Static("/static", "web/static")

	router.GET("/", index.GetIndex)
	router.GET("/main-menu", index.GetMainMenu)
	router.GET("/tissue_records/:id", tissuerecords.GetTissueRecordById)
	router.POST("/tissue_records", tissuerecords.CreateTissueRecord)
	router.PUT("/tissue_records/:id", tissuerecords.UpdateTissueRecord)
	router.DELETE("/tissue_records/:id", tissuerecords.DeleteTissueRecord)

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
	r.Run(":8080")
}
