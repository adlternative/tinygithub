package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.WithFields(log.Fields{
			"URL": c.Request.URL,
		}).Debug("new http request")
		c.Next()
	}
}

func Run() error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(Logger(), gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello TinyGithub!")
	})
	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return err
	}

	if err = r.Run(); err != nil {
		return err
	}
	return nil
}
