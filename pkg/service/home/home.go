package home

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

func Page(c *gin.Context) {
	c.HTML(http.StatusOK, "home.html", nil)
}
