package auth

import (
	service_manager "github.com/adlternative/tinygithub/pkg/manager"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// Logout user exit
func Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	err := session.Save()
	if err != nil {
		log.WithError(err).Error("session save failed when logout")
		c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, "/")
}

// LogoutV2 do user exit
func LogoutV2(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userName, ok1 := session.Get("username").(string)
		userID, ok2 := session.Get("user_id").(uint)

		var user model.User
		if !ok1 || !ok2 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "You are not logged in yet, so you cannot log out"})
			return
		}
		user.Name = userName
		user.ID = userID

		db := manager.DBEngine()
		if err := db.First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "The information of your session may be incorrect"})
			return
		}
		session.Clear()
		err := session.Save()
		if err != nil {
			log.WithError(err).Error("session save failed when logout")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "session save failed when logout"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "logout success"})
	}
}
