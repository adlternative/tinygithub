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
		if err := db.Where("name = ?", userName).First(&user).Error; err != nil {
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
