package router

import (
	"net/http"

	"github.com/adlternative/tinygithub/pkg/service"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.WithFields(log.Fields{
			"URL":    c.Request.URL,
			"Method": c.Request.Method,
		}).Debug("new http request")
		c.Next()
	}
}

func Run(store *storage.Storage) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(Logger(), gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello TinyGithub!")
	})
	r.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "Test TinyGithub!")
	})
	r.GET("/:username/:reponame/info/refs", func(c *gin.Context) {
		userName := c.Param("username")
		// check user exist
		repoName := c.Param("reponame")
		// check repo exist
		serviceName := c.Query("service")

		err := service.InfoRefs(c, store, userName, repoName, serviceName)
		if err != nil {
			log.WithError(err).Errorf("info refs failed")
			c.String(http.StatusInternalServerError, "get info refs failed with %s", err)
			return
		}
	})
	r.POST("/:username/:reponame/git-upload-pack", func(c *gin.Context) {
		userName := c.Param("username")
		// check user exist
		repoName := c.Param("reponame")
		// check repo exist
		err := service.UploadPack(c, store, userName, repoName)
		if err != nil {
			log.WithError(err).Errorf("upload-pack failed")
			c.String(http.StatusInternalServerError, "upload-pack failed with %s", err)
			return
		}
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
