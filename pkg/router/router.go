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

	gitRepoGroup := r.Group("/:username/:reponame")
	{
		gitRepoGroup.GET("/info/refs", service.InfoRefs(store))
		gitRepoGroup.POST("/git-upload-pack", service.UploadPack(store))
		gitRepoGroup.POST("/git-receive-pack", service.ReceivePack(store))
	}

	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return err
	}

	if err = r.Run(); err != nil {
		return err
	}
	return nil
}
