package main

import (
	"log"
	"mcba/tissquest/cmd/api-server-gin/atlas"
	"mcba/tissquest/cmd/api-server-gin/index"
	"mcba/tissquest/cmd/api-server-gin/tissue_records"
	"mcba/tissquest/internal/persistence/migration"
	"path/filepath"

	"github.com/gin-contrib/multitemplate"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

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
	r := gin.Default()

	// Set up HTML rendering using loadTemplates
	r.HTMLRender = loadTemplates("web/templates")

	// Serve static files
	r.Static("/static", "./web/static")

	// Routes
	r.GET("/", index.GetIndex)
	r.GET("/menu", index.GetMainMenu)

	// TissueRecord routes
	r.GET("/tissue_records", tissue_records.ListTissueRecords)
	r.POST("/tissue_records", tissue_records.CreateTissueRecord)
	r.GET("/tissue_records/:id", tissue_records.GetTissueRecordById)
	r.PUT("/tissue_records/:id", tissue_records.UpdateTissueRecord)
	r.DELETE("/tissue_records/:id", tissue_records.DeleteTissueRecord)

	// Atlas routes
	r.GET("/atlases", atlas.ListAtlases)
	r.POST("/atlases", atlas.CreateAtlas)
	r.GET("/atlases/:id", atlas.GetAtlas)
	r.PUT("/atlases/:id", atlas.UpdateAtlas)
	r.DELETE("/atlases/:id", atlas.DeleteAtlas)

	return r
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
