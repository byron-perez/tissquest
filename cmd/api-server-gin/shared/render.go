package shared

import (
	"html/template"
	"net/http"

	"github.com/gin-gonic/gin"
)

const baseLayout = "web/templates/layouts/base.html"

// IsHTMX returns true if the request was made by HTMX.
func IsHTMX(c *gin.Context) bool {
	return c.GetHeader("HX-Request") == "true"
}

// RenderPage renders a full page or an HTMX fragment depending on the request.
// For HTMX requests it executes only templateName from templateFiles.
// For full-page requests it parses base.html + templateFiles and executes "base".
func RenderPage(c *gin.Context, templateFiles []string, templateName string, data gin.H) {
	c.Header("Content-Type", "text/html")

	if IsHTMX(c) {
		tmpl := template.Must(template.ParseFiles(templateFiles...))
		if err := tmpl.ExecuteTemplate(c.Writer, templateName, data); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
		}
		return
	}

	files := append([]string{baseLayout}, templateFiles...)
	tmpl := template.Must(template.ParseFiles(files...))
	if err := tmpl.ExecuteTemplate(c.Writer, "base", data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

// RenderFragment always executes only the named template (no base layout).
func RenderFragment(c *gin.Context, templateFiles []string, templateName string, data gin.H) {
	c.Header("Content-Type", "text/html")

	tmpl := template.Must(template.ParseFiles(templateFiles...))
	if err := tmpl.ExecuteTemplate(c.Writer, templateName, data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

// RenderError renders an error response. For HTMX requests it returns plain text;
// otherwise it renders a full error page.
func RenderError(c *gin.Context, status int, message string) {
	c.Header("Content-Type", "text/html")
	c.Writer.WriteHeader(status)

	if IsHTMX(c) {
		c.String(status, message)
		return
	}

	tmpl := template.Must(template.ParseFiles(
		baseLayout,
		"web/templates/error.html",
	))
	if err := tmpl.ExecuteTemplate(c.Writer, "base", gin.H{"error": message}); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}

// AppendFragment parses templateFiles and appends the named template to the
// already-started response body. Use this for OOB swaps after the primary swap.
func AppendFragment(c *gin.Context, templateFiles []string, templateName string, data gin.H) {
	tmpl := template.Must(template.ParseFiles(templateFiles...))
	if err := tmpl.ExecuteTemplate(c.Writer, templateName, data); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
}
