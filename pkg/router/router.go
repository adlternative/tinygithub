package router

import (
	"fmt"
	"github.com/adlternative/tinygithub/pkg/config"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/service/auth"
	"github.com/adlternative/tinygithub/pkg/service/blob"
	"github.com/adlternative/tinygithub/pkg/service/home"
	"github.com/adlternative/tinygithub/pkg/service/pack"
	"github.com/adlternative/tinygithub/pkg/service/repo"
	"github.com/adlternative/tinygithub/pkg/service/user"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

func isDirectory(path string) bool {
	// 将相对路径转换为绝对路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// 检查文件或目录是否存在
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return false
	}

	// 检查是否是目录
	return info.IsDir()
}

func Run(store *storage.Storage, dbEngine *model.DBEngine) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	sessionSecret := viper.GetString(config.SessionSecret)
	if sessionSecret == "" {
		return fmt.Errorf("empty session secret")
	}
	sessionMiddleWare := sessions.Sessions("tinygithub-session", cookie.NewStore([]byte(sessionSecret)))

	r.Use(Logger(), gin.Recovery(), sessionMiddleWare)
	r.LoadHTMLGlob("pkg/template/*")

	staticResourcePath := viper.GetString(config.StaticResourcePath)

	if !isDirectory(staticResourcePath) {
		return fmt.Errorf("staticResourcePath %s is not a directory", staticResourcePath)
	}

	r.Static("/static", staticResourcePath)

	r.GET("/", home.Page)

	gitRepoGroup := r.Group("/:username/:reponame")
	gitRepoGroup.Use(func(c *gin.Context) {
		// 获取 reponame
		reponame := c.Param("reponame")

		// 检查 reponame 是否以 .git 结尾
		if strings.HasSuffix(reponame, ".git") {
			// 将请求重定向到 /:username/repo
			c.Redirect(http.StatusMovedPermanently, "/"+c.Param("username")+"/"+reponame[:len(reponame)-4])
			return
		}

		c.Next()
	})
	{
		gitRepoGroup.GET("", repo.Home(dbEngine, store))
		gitRepoGroup.GET("/tree/*treepath", repo.Home(dbEngine, store))
		gitRepoGroup.GET("/blob/*blobpath", blob.Show(dbEngine, store))

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

	if err = r.Run(fmt.Sprintf("%s:%s", viper.GetString(config.ServerIp), viper.GetString(config.ServerPort))); err != nil {
		return err
	}
	return nil
}
