package repo

import (
	"errors"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

func Create(db *model.DBEngine, store *storage.Storage) gin.HandlerFunc {
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
		_, err := store.CreateRepository(c, sessionUserName, repoName)
		if err != nil {
			tx.Rollback()

			err2 := store.RemoveRepository(c, sessionUserName, repoName)
			if err2 != nil {
				log.WithError(err2).Errorf("repo create rollback failed")
			}
			log.WithError(err).Errorf("store CreateRepository failed")
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}

		if err = tx.Create(&model.Repository{
			UserID: sessionUserID,
			Name:   repoName,
			Desc:   description,
		}).Error; err != nil {
			tx.Rollback()

			err2 := store.RemoveRepository(c, sessionUserName, repoName)
			if err2 != nil {
				log.WithError(err2).Errorf("repo create rollback failed")
			}

			log.WithError(err).Errorf("db CreateRepository failed")
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}

		if err = tx.Commit().Error; err != nil {
			tx.Rollback()

			err2 := store.RemoveRepository(c, sessionUserName, repoName)
			if err2 != nil {
				log.WithError(err2).Errorf("repo create rollback failed")
			}

			log.WithError(err).Errorf("txn commit failed")
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, "/"+sessionUserName+"/"+repoName)
	}
}
