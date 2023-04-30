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
	c.HTML(http.StatusOK, "create_repo.html", nil)
}

func Create(db *model.DBEngine, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {

		repoName := c.PostForm("reponame")
		description := c.PostForm("description")

		session := sessions.Default(c)
		userName, ok := session.Get("username").(string)
		if !ok {
			c.HTML(http.StatusUnauthorized, "401.html", nil)
			return
		}
		userID, ok := session.Get("user_id").(uint)
		if !ok {
			c.HTML(http.StatusUnauthorized, "401.html", nil)
			return
		}

		tx := db.Begin()

		// 判断仓库是否已经存在
		var existingRepository model.Repository
		if err := tx.Where("user_id = ? AND name = ?", userID, repoName).First(&existingRepository).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			tx.Rollback()
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}
		if existingRepository.ID != 0 {
			tx.Rollback()
			c.HTML(http.StatusConflict, "create_repo.html", gin.H{"error": fmt.Errorf("repository %s existed", existingRepository.Name)})
		}

		// git init
		_, err := store.CreateRepository(c, userName, repoName)
		if err != nil {
			tx.Rollback()

			err2 := store.RemoveRepository(c, userName, repoName)
			if err2 != nil {
				log.WithError(err2).Errorf("repo create rollback failed")
			}
			log.WithError(err).Errorf("store CreateRepository failed")
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}

		if err = tx.Create(&model.Repository{
			UserID: userID,
			Name:   repoName,
			Desc:   description,
		}).Error; err != nil {
			tx.Rollback()

			err2 := store.RemoveRepository(c, userName, repoName)
			if err2 != nil {
				log.WithError(err2).Errorf("repo create rollback failed")
			}

			log.WithError(err).Errorf("db CreateRepository failed")
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}

		if err = tx.Commit().Error; err != nil {
			tx.Rollback()

			err2 := store.RemoveRepository(c, userName, repoName)
			if err2 != nil {
				log.WithError(err2).Errorf("repo create rollback failed")
			}

			log.WithError(err).Errorf("txn commit failed")
			c.HTML(http.StatusInternalServerError, "500.html", gin.H{"error": err.Error()})
			return
		}

		c.Redirect(http.StatusFound, "/"+userName+"/"+repoName)
	}
}
