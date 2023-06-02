package pack

import (
	"fmt"
	service_manager "github.com/adlternative/tinygithub/pkg/manager"
	"github.com/adlternative/tinygithub/pkg/service/protocol"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
)

func checkRequest(req *protocol.PostReceiveRequest) error {
	if req.OldOid == "" || req.NewOid == "" || req.RefName == "" {
		return fmt.Errorf("empty request fields")
	}
	return nil
}

func PostReceive(manager *service_manager.ServiceManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req protocol.PostReceiveRequest
		if err := c.BindJSON(&req); err != nil {
			log.WithError(err).Errorf("failed to parse JSON data")
			c.AbortWithStatusJSON(http.StatusBadRequest, &protocol.PostReceiveError{Error: "Invalid JSON data"})
			return
		}
		if err := checkRequest(&req); err != nil {
			log.WithError(err).Errorf("invalid PostReceiveRequest data")
			c.AbortWithStatusJSON(http.StatusBadRequest, &protocol.PostReceiveError{Error: err.Error()})
			return
		}

		log.Println("Received data:", req)
		c.JSON(http.StatusOK, &protocol.PostReceiveResponse{Message: "Data received"})
	}
}
