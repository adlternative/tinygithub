package user

import (
	service_manager "github.com/adlternative/tinygithub/pkg/manager"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Home user's home page
func Home(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {

		// url
		userName := c.Param("username")
		var user model.User

		db := manager.DBEngine()
		if err := db.Preload("Repositories").Where("name = ?", userName).First(&user).Error; err != nil {
			c.HTML(http.StatusNotFound, "404.html", nil)
			return
		}

		session := sessions.Default(c)
		sessionUserName, ok := session.Get("username").(string)
		if !ok {
			c.HTML(http.StatusUnauthorized, "401.html", nil)
			return
		}

		if userName != sessionUserName {
			c.HTML(http.StatusUnauthorized, "401.html", nil)
			return
		}

		c.HTML(http.StatusOK, "user.html", gin.H{"user": user})
	}
}

// UserInfoV2 show user information
func UserInfoV2(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		db := manager.DBEngine()

		userName := c.Param("username")
		tab := c.Query("tab")
		if tab == "repositories" {
			if err := db.Preload("Repositories").Where("name = ?", userName).First(&user).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"repositories": user.Repositories,
			})
		} else {
			if err := db.Where("name = ?", userName).First(&user).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"name":  user.Name,
				"email": user.Email,
			})
		}

	}
}

func CurrentUserInfo(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)

		userName, ok1 := session.Get("username").(string)
		userID, ok2 := session.Get("user_id").(uint)
		var user model.User
		if !ok1 || !ok2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unknown current user"})
			return
		}
		user.Name = userName
		user.ID = userID

		db := manager.DBEngine()
		if err := db.First(&user).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "not such user in the server"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user": user,
		})
	}
}
