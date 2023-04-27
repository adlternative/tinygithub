package router

import (
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/service/auth"
	"github.com/adlternative/tinygithub/pkg/service/home"
	"github.com/adlternative/tinygithub/pkg/service/pack"
	"github.com/adlternative/tinygithub/pkg/service/repo"
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
		gitRepoGroup.GET("", repo.Home(dbEngine, store))
		gitRepoGroup.GET("/info/refs", pack.InfoRefs(store))
		gitRepoGroup.POST("/git-upload-pack", pack.UploadPack(store))
		gitRepoGroup.POST("/git-receive-pack", pack.ReceivePack(store))
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

	repoGroup := r.Group("/repos")
	{
		repoGroup.GET("/new", repo.CreatePage)
		repoGroup.POST("/new", repo.Create(dbEngine, store))
		//repoGroup.Get("/:id", repo.Get(dbEngine))
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
