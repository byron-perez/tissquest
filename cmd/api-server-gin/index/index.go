package index

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
	"mcba/tissquest/internal/persistence/repositories"
	"mcba/tissquest/internal/services"
)

func GetIndex(c *gin.Context) {
	slideSvc := services.NewSlideService(nil, repositories.NewSlideRepository())
	randomSlides, _ := slideSvc.GetRandomTiledDisplaySlides(3)

	data := gin.H{
		"Title":        "Tissquest — Histology Library",
		"RandomSlides": randomSlides,
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
