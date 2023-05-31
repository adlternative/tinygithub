package router

import (
	"fmt"
	"github.com/adlternative/tinygithub/pkg/config"
	"github.com/adlternative/tinygithub/pkg/manager"
	"github.com/adlternative/tinygithub/pkg/service/auth"
	"github.com/adlternative/tinygithub/pkg/service/blob"
	"github.com/adlternative/tinygithub/pkg/service/branches"
	"github.com/adlternative/tinygithub/pkg/service/home"
	"github.com/adlternative/tinygithub/pkg/service/pack"
	"github.com/adlternative/tinygithub/pkg/service/repo"
	"github.com/adlternative/tinygithub/pkg/service/tags"
	"github.com/adlternative/tinygithub/pkg/service/tree"
	"github.com/adlternative/tinygithub/pkg/service/user"
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

func Run(manager *service_manager.ServiceManager) error {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(DefaultCORS())

	// default options handle
	r.OPTIONS("/*path", DefaultOptions)

	sessionSecret := viper.GetString(config.SessionSecret)
	if sessionSecret == "" {
		return fmt.Errorf("empty session secret")
	}
	sessionMiddleWare := sessions.Sessions("tinygithub-session", cookie.NewStore([]byte(sessionSecret)))

	r.Use(Logger(), gin.Recovery(), sessionMiddleWare)

	if viper.GetString(config.APIVersion) == "v1" {
		r.NoRoute(func(c *gin.Context) {
			c.HTML(http.StatusNotFound, "404.html", nil)
		})

		r.GET("/", home.Page(manager))

		htmlTemplatePath := viper.GetString(config.HtmlTemplatePath)

		if !isDirectory(htmlTemplatePath) {
			return fmt.Errorf("htmlTemplatePath %s is not a directory", htmlTemplatePath)
		}
		r.LoadHTMLGlob(fmt.Sprintf("%s/*", htmlTemplatePath))

		staticResourcePath := viper.GetString(config.StaticResourcePath)

		if !isDirectory(staticResourcePath) {
			return fmt.Errorf("staticResourcePath %s is not a directory", staticResourcePath)
		}

		r.Static("/static", staticResourcePath)
		r.StaticFile("/favicon.icon", staticResourcePath+"/favicon.icon")

		authGroup := r.Group("/user")
		{

			registerGroup := authGroup.Group("/register")
			{
				registerGroup.GET("", auth.RegisterPage)
				registerGroup.POST("", auth.Register(manager))
			}

			loginGroup := authGroup.Group("/login")
			{
				loginGroup.GET("", auth.LoginPage)
				loginGroup.POST("", auth.Login(manager))
			}

			authGroup.GET("/logout", auth.Logout)
		}

		gitRepoGroup := r.Group("/:username/:reponame")
		{
			gitRepoGroup.GET("", repo.Home(manager))
			gitRepoGroup.GET("/tree/*treepath", repo.Home(manager))
			gitRepoGroup.GET("/blob/*blobpath", blob.Show(manager))

			gitRepoGroup.GET("/info/refs", pack.InfoRefs(manager))
			gitRepoGroup.POST("/git-upload-pack", pack.UploadPack(manager))
			gitRepoGroup.POST("/git-receive-pack", pack.ReceivePack(manager))
		}

		r.GET("/:username", user.Home(manager))

		repoGroup := r.Group("/repos")
		{
			repoGroup.GET("/new", repo.CreatePage)
			repoGroup.POST("/new", repo.Create(manager))
			repoGroup.POST("/delete", repo.Delete(manager))

			//repoGroup.Get("/:id", repo.Get(manager))
		}
	} else {
		gitRepoGroup := r.Group("/:username/:reponame")
		{
			gitRepoGroup.GET("/info/refs", pack.InfoRefs(manager))
			gitRepoGroup.POST("/git-upload-pack", pack.UploadPack(manager))
			gitRepoGroup.POST("/git-receive-pack", pack.ReceivePack(manager))
		}

		apiGroup := r.Group("/api")
		{
			v2Group := apiGroup.Group("/v2")
			{
				v2AuthGroup := v2Group.Group("/auth")
				{
					v2AuthGroup.POST("/login", auth.LoginV2(manager))
					v2AuthGroup.POST("/register", auth.RegisterV2(manager))
					v2AuthGroup.GET("/logout", auth.LogoutV2(manager))

				}
				v2UserGroup := v2Group.Group("/users")
				{
					v2UserGroup.GET("/current", user.CurrentUserInfo(manager))

					v2UserNameGroup := v2UserGroup.Group("/:username")
					{
						v2UserNameGroup.GET("", user.UserInfoV2(manager))
					}
				}

				v2ReposGroup := v2Group.Group("/repos")
				{
					v2ReposGroup.POST("/new", repo.CreateV2(manager))
					v2ReposGroup.POST("/delete", repo.DeleteV2(manager))
				}

				v2UserNameGroup := v2Group.Group("/:username")
				{
					v2RepoGroup := v2UserNameGroup.Group("/:reponame")
					{
						v2RepoGroup.GET("", repo.ShowRepo(manager))

						branchesGroup := v2RepoGroup.Group("/branches")
						{
							branchesGroup.GET("", branches.Show(manager))
						}
						tagsGroup := v2RepoGroup.Group("/tags")
						{
							tagsGroup.GET("", tags.Show(manager))
						}
						treeGroup := v2RepoGroup.Group("/tree")
						{
							treeGroup.GET("", tree.Show(manager))
						}
						blobGroup := v2RepoGroup.Group("/blob")
						{
							blobGroup.GET("", blob.ShowV2(manager))
						}
					}
				}

			}
		}
	}
	err := r.SetTrustedProxies([]string{"127.0.0.1"})
	if err != nil {
		return err
	}

	serverAddr := fmt.Sprintf("%s:%s", viper.GetString(config.ServerIp), viper.GetString(config.ServerPort))
	log.Infof("tinygithub server run on %s", serverAddr)

	if err = r.Run(serverAddr); err != nil {
		return err
	}
	return nil
}
