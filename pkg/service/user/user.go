package user

import (
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
)

// Home user's home page
func Home(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {

		// url
		userName := c.Param("username")
		var user model.User
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
func UserInfoV2(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User
		userName := c.Param("username")
		tab := c.Query("tab")
		if tab == "repositories" {
			if err := db.Preload("Repositories").Where("name = ?", userName).First(&user).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": err.Error(),
				})
				return
			}
		} else {
			if err := db.Where("name = ?", userName).First(&user).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{
					"error": err.Error(),
				})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"user": user,
		})
	}
}
