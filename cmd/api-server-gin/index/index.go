package index

import (
	"html/template"
	"mcba/tissquest/internal/core/collection"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetIndex(c *gin.Context) {
	collections, err := fetchCollections()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collections"})
		return
	}

	var featured *collection.Collection
	if len(collections) > 0 {
		featured = &collections[0]
	}

	data := gin.H{
		"Title":       "Tissquest",
		"Collections": collections,
		"Featured":    featured,
	}

	tmpl := template.Must(template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/index.html",
		"web/templates/includes/main-menu.html",
	))
	c.Header("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(c.Writer, "base", data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

func fetchCollections() ([]collection.Collection, error) {
	svc := services.NewCollectionService(repositories.NewCollectionRepository(), nil)
	return svc.ListCollections()
}
