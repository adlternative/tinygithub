package repo

import (
	"errors"
	"fmt"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/adlternative/tinygithub/pkg/utils"
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

type CreateRequest struct {
	RepoName    string `json:"repoName"`
	Description string `json:"description"`
}

func checkRequest(req *CreateRequest) error {
	if !utils.IsAlphanumeric(req.RepoName) {
		return fmt.Errorf("invalid reponame %s, only alpha allow", req.RepoName)
	}

	return nil
}

var RepoAlreadyExistsError = errors.New("repo already exists")

func CreateV2(db *model.DBEngine, store *storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateRequest

		err := c.BindJSON(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid create repo request"})
			return
		}
		// check request
		err = checkRequest(&req)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		session := sessions.Default(c)
		userName, ok := session.Get("username").(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "you cannot create repo because you are not login"})
			return
		}
		userID, ok := session.Get("user_id").(uint)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "you cannot create repo because you are not login"})
			return
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			var user model.User
			user.ID = userID

			// check user exists
			if err := tx.First(&user).Error; err != nil {
				return err
			}

			// check repo not exists
			var existingRepository model.Repository
			if err := tx.Where("user_id = ? AND name = ?", userID, req.RepoName).First(&existingRepository).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			if existingRepository.ID != 0 {
				return RepoAlreadyExistsError
			}
			// git init
			_, err := store.CreateRepository(c, userName, req.RepoName)
			if err != nil {
				log.WithError(err).Errorf("store CreateRepository failed")

				err2 := store.RemoveRepository(c, userName, req.RepoName)
				if err2 != nil {
					log.WithError(err2).Errorf("store RemoveRepository failed when rollback")
				}
				return err
			}

			// db create record
			if err := tx.Create(&model.Repository{
				UserID: userID,
				Name:   req.RepoName,
				Desc:   req.Description,
			}).Error; err != nil {
				log.WithError(err).Errorf("db CreateRepository record failed")
				err2 := store.RemoveRepository(c, userName, req.RepoName)
				if err2 != nil {
					log.WithError(err2).Errorf("store RemoveRepository failed when rollback")
				}
				return err
			}
			return nil
		})

		if err != nil {
			log.WithError(err).Errorf("CreateRepo failed")
			if errors.Is(err, RepoAlreadyExistsError) {
				c.JSON(http.StatusConflict, err.Error())
				return
			}

			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "create repository success",
			"repoName": req.RepoName,
		})
	}
}
