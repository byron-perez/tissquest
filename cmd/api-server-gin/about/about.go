package about

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetAbout(c *gin.Context) {
	tmpl := template.Must(template.ParseFiles(
		"web/templates/layouts/base.html",
		"web/templates/pages/about.html",
		"web/templates/includes/main-menu.html",
	))
	c.Header("Content-Type", "text/html")
	if err := tmpl.ExecuteTemplate(c.Writer, "base", gin.H{
		"Title": "Acerca de — TissQuest",
	}); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
