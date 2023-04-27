package auth

import (
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/service/auth/cryto"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func LoginPage(c *gin.Context) {
	c.HTML(http.StatusOK, "login.html", nil)
}

// Login check if user's name, email exists and password right
func Login(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User

		account := c.PostForm("account")
		if account == "" {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "invalid account"})
			return
		}

		db.Where("name = ? OR email = ?", account, account).First(&user)
		if user.ID == 0 {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "invalid username or email"})
			return
		}

		var password model.Password
		db.Where("user_id = ?", user.ID).First(&password)

		if password.ID == 0 || !cryto.CheckPasswordHash(c.PostForm("password"), password.Password) {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "invalid password"})
			return
		}

		session := sessions.Default(c)
		session.Set("username", user.Name)
		session.Set("user_id", user.ID)

		err := session.Save()
		if err != nil {
			log.WithError(err).Error("session save failed when login")
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}

		log.WithFields(
			log.Fields{
				"username": user.Name,
				"user_id":  user.ID,
			}).Info("login success")

		c.Redirect(http.StatusFound, "/"+user.Name)
	}
}
