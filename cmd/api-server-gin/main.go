package main

import (
	"fmt"
	"html/template"
	"log"
	"mcba/tissquest/cmd/api-server-gin/atlas"
	"mcba/tissquest/cmd/api-server-gin/index"
	"mcba/tissquest/cmd/api-server-gin/tissue_records"
	"mcba/tissquest/internal/persistence/migration"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func loadTemplates(templatesDir string) (*template.Template, error) {
	// First, load the base template
	baseTemplate := filepath.Join(templatesDir, "layouts", "base.html")
	templ, err := template.New("base.html").ParseFiles(baseTemplate)
	if err != nil {
		return nil, fmt.Errorf("error parsing base template: %v", err)
	}

	// Then, walk through the templates directory and parse all other templates
	err = filepath.Walk(templatesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".html" && path != baseTemplate {
			// Get the relative path from the templatesDir
			relPath, err := filepath.Rel(templatesDir, path)
			if err != nil {
				return err
			}
			// Use the relative path as the template name
			_, err = templ.New(filepath.ToSlash(relPath)).ParseFiles(path)
			if err != nil {
				log.Printf("Error parsing template %s: %v", path, err)
				return err
			}
			log.Printf("Loaded template: %s", relPath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return templ, nil
}

func setupRouter() (*gin.Engine, error) {
	r := gin.Default()

	// Load HTML templates
	templatesDir := "./web/templates"
	templ, err := loadTemplates(templatesDir)
	if err != nil {
		return nil, err
	}
	// Serve static files
	r.Static("/static", "./web/static")
	r.SetHTMLTemplate(templ)

	// Setup routes
	r.GET("/", index.GetIndex)
	r.GET("/tissue_records", tissue_records.ListTissueRecords)
	r.POST("/tissue_records", tissue_records.CreateTissueRecord)
	r.GET("/atlases", atlas.ListAtlases)

	return r, nil
}

func main() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Print current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current working directory: %v", err)
	} else {
		log.Printf("Current working directory: %s", cwd)
	}

	// Run migrations
	migration.RunMigration()

	// Setup and run the server
	r, err := setupRouter()
	if err != nil {
		log.Fatalf("Failed to set up router: %v", err)
	}
	r.Run(":8080")
}
