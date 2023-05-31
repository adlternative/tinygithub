package tags

import (
	"fmt"
	"github.com/adlternative/tinygithub/pkg/git/tags"
	service_manager "github.com/adlternative/tinygithub/pkg/manager"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Show(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		userName := c.Param("username")
		repoName := c.Param("reponame")

		var user model.User
		db := manager.DBEngine()
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
		store := manager.Storage()
		repo, err := store.GetRepository(userName, repoName)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Errorf("no such repository"),
			})
			return
		}
		allTags, err := tags.GetAllTags(c, repo.Path())
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"error": fmt.Sprintf("GetAllBranch failed with %v", err),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"tags": allTags,
		})
		return
	}
}
