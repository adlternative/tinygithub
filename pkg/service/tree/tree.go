package tree

import (
	"fmt"
	gitRepo "github.com/adlternative/tinygithub/pkg/git/repo"
	"github.com/adlternative/tinygithub/pkg/git/tree"
	service_manager "github.com/adlternative/tinygithub/pkg/manager"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"strings"
)

func Show(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userName := c.Param("username")
		repoName := c.Param("reponame")
		revision := c.Query("revision")
		path := c.Query("path")

		if revision == "" {
			revision = "HEAD"
		}
		if strings.HasSuffix(path, "/") {
			path = strings.TrimSuffix(path, "/")
		}
		if strings.HasPrefix(path, "/") {
			path = strings.TrimPrefix(path, "/")
		}

		var user model.User
		user.Name = userName

		db := manager.DBEngine()
		if err := db.Where("name = ?", userName).Preload("Repositories", "name = ?", repoName).First(&user).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": fmt.Sprintf("user %s repo %s not found", userName, repoName),
				})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}

		if len(user.Repositories) == 0 {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("user %s repo %s not found", userName, repoName),
			})
			return
		}

		store := manager.Storage()
		repo, err := store.GetRepository(userName, repoName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("user %s repo %s not found", userName, repoName),
			})
			return
		}
		// git rev-parse
		isEmpty, _ := gitRepo.IsRepositoryEmpty(c, repo.Path())
		var entries []*tree.BlameTreeEntry
		if !isEmpty {
			// git ls-tree
			entries, err = tree.ParseBlameTree(c, repo.Path(), revision, path)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"revision":     revision,
			"tree_path":    path,
			"tree_entries": entries,
		})
	}
}
