package auth

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/service/auth/cryto"
)

func isAlphanumeric(s string) bool {
	re := regexp.MustCompile("^[a-zA-Z0-9]+$")
	return re.MatchString(s)
}

func isReserved(s string) bool {
	switch s {
	case "user":
		return true
	default:
		return false
	}
}

func RegisterPage(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", nil)
}

// Register take the account and password to create a new user
func Register(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user model.User

		if err := c.ShouldBind(&user); err != nil || !isAlphanumeric(user.Name) {
			c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": "invalid username or email"})
			return
		}
		if isReserved(user.Name) {
			c.HTML(http.StatusBadRequest, "register.html", gin.H{"error": fmt.Sprintf("The username \"%s\" is reserved", user.Name)})
			return
		}

		tx := db.Begin()
		code, err := registerTx(c, tx, &user)
		if err != nil {
			tx.Rollback()
			switch code {
			case http.StatusBadRequest:
				c.HTML(code, "register.html", gin.H{"error": err.Error()})
			default:
				fallthrough
			case http.StatusInternalServerError:
				c.HTML(code, "500.html", gin.H{"error": err.Error()})
			}
			return
		} else {
			if err = tx.Commit().Error; err != nil {
				tx.Rollback()
				c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
				return
			}
		}

		c.Redirect(http.StatusFound, "/"+user.Name)
	}
}

func registerTx(c *gin.Context, tx *gorm.DB, user *model.User) (int, error) {
	var existedUser model.User

	err := tx.Where("name = ?", user.Name).First(&existedUser).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		tx.Rollback()
		return http.StatusInternalServerError, err
	} else if err == nil {
		tx.Rollback()
		return http.StatusBadRequest, err
	}

	err = tx.Where("email = ?", user.Email).First(&existedUser).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return http.StatusInternalServerError, err
	} else if err == nil {
		return http.StatusBadRequest, err
	}

	if err = tx.Create(&user).Error; err != nil {
		return http.StatusInternalServerError, err
	}

	var password model.Password
	password.UserID = user.ID

	passwordPlainText := c.PostForm("password")

	hashPassword, err := cryto.HashPassword(passwordPlainText)
	if err != nil {
		return http.StatusInternalServerError, err
	}
	password.Password = hashPassword

	if err := tx.Create(&password).Error; err != nil {
		return http.StatusInternalServerError, err
	}

	// 设置 session
	session := sessions.Default(c)
	session.Set("username", user.Name)
	session.Set("user_id", user.ID)
	err = session.Save()
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return http.StatusFound, nil
}
