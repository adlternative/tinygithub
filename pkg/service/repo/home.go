package repo

import (
	"fmt"
	"github.com/adlternative/tinygithub/pkg/config"
	gitRepo "github.com/adlternative/tinygithub/pkg/git/repo"
	"github.com/adlternative/tinygithub/pkg/git/tree"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func ShowRepos(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		userName := c.Param("username")
		var user model.User
		if err := db.Preload("Repositories").Where("name = ?", userName).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user": user,
		})
	}
}

func ShowRepo(db *model.DBEngine, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		userName := c.Param("username")
		repoName := c.Param("reponame")

		var user model.User
		if err := db.Preload("Repositories", "name = ?", repoName).Where("name = ?", userName).First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
			return
		}
		if len(user.Repositories) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Errorf("no such repository"),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"repo": user.Repositories[0],
		})
	}
}

func Home(db *model.DBEngine, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		userName := c.Param("username")
		repoName := strings.TrimSuffix(c.Param("reponame"), ".git")
		treePath := c.Param("treepath")

		if strings.HasSuffix(treePath, "/") {
			treePath = strings.TrimSuffix(treePath, "/")
		}
		if strings.HasPrefix(treePath, "/") {
			treePath = strings.TrimPrefix(treePath, "/")
		}

		var user model.User
		user.Name = userName

		if err := db.Where("name = ?", userName).Preload("Repositories", "name = ?", repoName).First(&user).Error; err != nil {
			// 处理错误
			if err == gorm.ErrRecordNotFound {
				c.HTML(http.StatusNotFound, "404.html", nil)
				return
			} else {
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
		}

		if len(user.Repositories) == 0 {
			c.HTML(http.StatusNotFound, "404.html", nil)
			return
		} else if len(user.Repositories) > 1 {
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{
				"error": "multiple repo same name",
			})
			return
		}

		repo, err := store.GetRepository(userName, repoName)
		if err != nil {
			c.HTML(http.StatusNotFound, "404.html", nil)
			return
		}

		revision := "HEAD"

		// git rev-parse
		isEmpty, _ := gitRepo.IsRepositoryEmpty(c, repo.Path())
		var entries []*tree.Entry
		if !isEmpty {
			// git ls-tree
			entries, err = tree.ParseTree(c, repo.Path(), revision, treePath)
			if err != nil {
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{
					"error": err.Error(),
				})
				return
			}
		}

		c.HTML(http.StatusOK, "repo.html", gin.H{
			"RepoName":    user.Repositories[0].Name,
			"Description": user.Repositories[0].Desc,
			"Owner":       userName,
			"DownloadURL": fmt.Sprintf("http://%s:%s/%s/%s.git", viper.GetString(config.ServerIp), viper.GetString(config.ServerPort), userName, repoName),
			"TreePath":    treePath,
			"TreeEntries": entries,
		})
	}
}
