package auth

import (
	service_manager "github.com/adlternative/tinygithub/pkg/manager"
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
func Login(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User

		account := c.PostForm("account")
		if account == "" {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{"error": "invalid account"})
			return
		}
		db := manager.DBEngine()
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

type loginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

// LoginV2 check if user's name, email exists and password right
func LoginV2(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		contentType := c.GetHeader("Content-Type")
		switch contentType {
		case "application/json":
			var req loginRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err})
			}
			var user model.User

			if req.Account == "" {
				log.Error("invalid account")
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account"})
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account"})
				return
			}
			db := manager.DBEngine()
			db.Where("name = ? OR email = ?", req.Account, req.Account).First(&user)
			if user.ID == 0 {
				c.JSON(http.StatusNotFound, gin.H{"error": "invalid username or email"})
				return
			}

			var password model.Password
			db.Where("user_id = ?", user.ID).First(&password)

			if password.ID == 0 || !cryto.CheckPasswordHash(req.Password, password.Password) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
				return
			}

			session := sessions.Default(c)
			session.Set("username", user.Name)
			session.Set("user_id", user.ID)

			err := session.Save()
			if err != nil {
				log.WithError(err).Error("session save failed when login")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "session save failed when login"})
				return
			}

			log.WithFields(
				log.Fields{
					"username": user.Name,
					"user_id":  user.ID,
				}).Info("login success")

			c.JSON(http.StatusOK, gin.H{"message": "login success"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request format"})
		}
	}
}
