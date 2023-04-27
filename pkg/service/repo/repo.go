package repo

import (
	"errors"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
)

func CreatePage(c *gin.Context) {
	session := sessions.Default(c)
	sessionUserName := session.Get("username").(string)
	if sessionUserName == "" {
		c.Redirect(http.StatusFound, "/user/login")
		return
	}
	c.HTML(http.StatusOK, "create_repo.html", nil)
}

func Create(db *model.DBEngine) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		sessionUserName := session.Get("username").(string)
		if sessionUserName == "" {
			c.Redirect(http.StatusFound, "/user/login")
			return
		}
		sessionUserID := session.Get("user_id").(uint)

		repoName := c.PostForm("reponame")
		description := c.PostForm("description")

		tx := db.Begin()

		// 判断仓库是否已经存在
		var existingRepository model.Repository
		if err := tx.Where("user_id = ? AND name = ?", sessionUserID, repoName).First(&existingRepository).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}
		if existingRepository.ID != 0 {
			tx.Rollback()
			c.HTML(http.StatusConflict, "create_repo.html", gin.H{"error": fmt.Errorf("repository %s existed", existingRepository.Name)})
		}

		// git init
		if err := tx.Create(&model.Repository{
			UserID: sessionUserID,
			Name:   repoName,
			Desc:   description,
		}).Error; err != nil {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}

		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, "/"+sessionUserName+"/"+repoName)
	}
}
