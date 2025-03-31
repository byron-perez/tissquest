package index

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func GetIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "Tissquest",
	})
}

func GetMainMenu(c *gin.Context) {
	c.HTML(http.StatusOK, "main-menu.html", gin.H{
		"title": "Menu",
	})
}
