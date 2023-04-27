package router

import (
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/service"
	"github.com/adlternative/tinygithub/pkg/service/auth"
	"github.com/adlternative/tinygithub/pkg/service/home"
	"github.com/adlternative/tinygithub/pkg/service/user"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
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

func Run(store *storage.Storage, dbEngine *model.DBEngine) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(Logger(),
		gin.Recovery(),
		sessions.Sessions("tinygithub-session", cookie.NewStore([]byte("secret"))))
	r.LoadHTMLGlob("pkg/template/*")

	r.GET("/", home.Page)

	gitRepoGroup := r.Group("/:username/:reponame")
	{
		gitRepoGroup.GET("/info/refs", service.InfoRefs(store))
		gitRepoGroup.POST("/git-upload-pack", service.UploadPack(store))
		gitRepoGroup.POST("/git-receive-pack", service.ReceivePack(store))
	}

	r.GET("/:username", user.Home(dbEngine))

	authGroup := r.Group("/user")
	{

		registerGroup := authGroup.Group("/register")
		{
			registerGroup.GET("", auth.RegisterPage)
			registerGroup.POST("", auth.Register(dbEngine))
		}

		loginGroup := authGroup.Group("/login")
		{
			loginGroup.GET("", auth.LoginPage)
			loginGroup.POST("", auth.Login(dbEngine))
		}

		authGroup.GET("/logout", auth.Logout)
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
