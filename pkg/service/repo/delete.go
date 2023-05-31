package repo

import (
	"fmt"
	service_manager "github.com/adlternative/tinygithub/pkg/manager"
	"github.com/adlternative/tinygithub/pkg/model"
	"github.com/adlternative/tinygithub/pkg/storage"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"net/http"
	"os"
)

type DeleteRequest struct {
	Owner string `json:"owner"`
	Repo  string `json:"repo"`
}

// Delete repo record from database and git storage
// 1. backup .git -> .git.backup
// 2. delete .git
// 2. delete db record
// 3. commit txn
// 4. remove backup
//
// TODO(adl): If the database crashes while deleting
// repository records, we need to recover the Git repository
// data again during the database recovery process.
// This process should be delegated to an external stable
// storage to record the task, such as Redis or message queues.
func Delete(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Info("repo.Delete called")
		var req DeleteRequest
		if err := c.BindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		store := manager.Storage()
		storeRepo, err := store.GetRepository(req.Owner, req.Repo)
		if err != nil {
			return
		}
		backUpPath := storeRepo.Path() + ".backup"

		db := manager.DBEngine()
		// remove from db
		err = db.Transaction(func(txn *gorm.DB) error {
			// Find the user by name
			var user model.User
			if err = txn.Where("name = ?", req.Owner).First(&user).Error; err != nil {
				return err
			}
			// Find the repository by name and user ID
			var repo model.Repository
			if err = txn.Unscoped().Where("user_id = ? and name = ?", user.ID, req.Repo).First(&repo).Error; err != nil {
				return err
			}

			if err = storage.BackUp(storeRepo.Path(), backUpPath); err != nil {
				return err
			}
			if err = os.RemoveAll(storeRepo.Path()); err != nil {
				return fmt.Errorf("remove git repo failed: %w", err)
			}

			if err = txn.Unscoped().Delete(&repo).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.WithError(err).Errorf("delete repository txn commit failed")

			if err := storage.Restore(storeRepo.Path(), backUpPath); err != nil {
				log.WithError(err).Errorf("restore repository failed")
			}
			err = os.RemoveAll(backUpPath)
			if err != nil {
				log.WithError(err).Errorf("remove backupPath %s failed", backUpPath)
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete repository"})
			return
		}

		err = os.RemoveAll(backUpPath)
		if err != nil {
			log.WithError(err).Errorf("remove backupPath %s failed", backUpPath)
		}

		c.JSON(http.StatusOK, gin.H{"message": "Repository deleted successfully"})
	}
}

type DeleteRequestV2 struct {
	RepoName string `json:"reponame"`
}

func DeleteV2(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req DeleteRequestV2
		if err := c.BindJSON(&req); err != nil {
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

		// remove from db
		db := manager.DBEngine()
		err := db.Transaction(func(txn *gorm.DB) error {
			var (
				user model.User
				err  error
			)

			// Find the user by name
			user.ID = userID

			if err = txn.First(&user).Error; err != nil {
				return err
			}
			// Find the repository by name and user ID
			var repo model.Repository
			if err = txn.Unscoped().Where("user_id = ? and name = ?", user.ID, req.RepoName).First(&repo).Error; err != nil {
				return err
			}

			// remove git storage
			store := manager.Storage()
			if err = store.RemoveRepository(c, userName, req.RepoName); err != nil {
				return err
			}

			// remove from db
			if err = txn.Unscoped().Delete(&repo).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.WithError(err).Errorf("delete repository txn commit failed")

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete repository"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("repository %s deleted successfully", req.RepoName)})
	}
}
