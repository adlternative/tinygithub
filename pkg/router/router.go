package router

import (
	"fmt"
	"github.com/adlternative/tinygithub/pkg/config"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/service/auth"
	"github.com/adlternative/tinygithub/pkg/service/blob"
	"github.com/adlternative/tinygithub/pkg/service/branches"
	"github.com/adlternative/tinygithub/pkg/service/home"
	"github.com/adlternative/tinygithub/pkg/service/pack"
	"github.com/adlternative/tinygithub/pkg/service/repo"
	"github.com/adlternative/tinygithub/pkg/service/tags"
	"github.com/adlternative/tinygithub/pkg/service/tree"
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

func DefaultCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:8080")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Next()
	}
}

func DefaultOptions(c *gin.Context) {
	c.Status(http.StatusNoContent)
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

	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusNotFound, "404.html", nil)
	})
	r.Use(DefaultCORS())

	// default options handle
	r.OPTIONS("/*path", DefaultOptions)

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
	r.StaticFile("/favicon.icon", staticResourcePath+"/favicon.icon")

	r.GET("/", home.Page(dbEngine))

	apiGroup := r.Group("/api")
	{
		v2Group := apiGroup.Group("/v2")
		{
			v2AuthGroup := v2Group.Group("/auth")
			{
				v2AuthGroup.POST("/login", auth.LoginV2(dbEngine))
				v2AuthGroup.POST("/register", auth.RegisterV2(dbEngine))
				v2AuthGroup.GET("/logout", auth.LogoutV2(dbEngine))

			}
			v2UserGroup := v2Group.Group("/users")
			{
				v2UserGroup.GET("/current", user.CurrentUserInfo(dbEngine))

				v2UserNameGroup := v2UserGroup.Group("/:username")
				{
					v2UserNameGroup.GET("", user.UserInfoV2(dbEngine))
				}
			}

			v2UserNameGroup := v2Group.Group("/:username")
			{
				v2RepoGroup := v2UserNameGroup.Group("/:reponame")
				{
					v2RepoGroup.GET("", repo.ShowRepo(dbEngine, store))

					branchesGroup := v2RepoGroup.Group("/branches")
					{
						branchesGroup.GET("", branches.Show(dbEngine, store))
					}
					tagsGroup := v2RepoGroup.Group("/tags")
					{
						tagsGroup.GET("", tags.Show(dbEngine, store))
					}
					treeGroup := v2RepoGroup.Group("/tree")
					{
						treeGroup.GET("", tree.Show(dbEngine, store))
					}
					blobGroup := v2RepoGroup.Group("/blob")
					{
						blobGroup.GET("", blob.ShowV2(dbEngine, store))
					}
				}
			}

		}
	}

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

	gitRepoGroup := r.Group("/:username/:reponame")
	{
		gitRepoGroup.GET("", repo.Home(dbEngine, store))
		gitRepoGroup.GET("/tree/*treepath", repo.Home(dbEngine, store))
		gitRepoGroup.GET("/blob/*blobpath", blob.Show(dbEngine, store))

		gitRepoGroup.GET("/info/refs", pack.InfoRefs(store))
		gitRepoGroup.POST("/git-upload-pack", pack.UploadPack(store))
		gitRepoGroup.POST("/git-receive-pack", pack.ReceivePack(store))
	}

	r.GET("/:username", user.Home(dbEngine))

	repoGroup := r.Group("/repos")
	{
		repoGroup.GET("/new", repo.CreatePage)
		repoGroup.POST("/new", repo.Create(dbEngine, store))
		repoGroup.POST("/delete", repo.Delete(dbEngine, store))

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
