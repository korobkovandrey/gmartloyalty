package orders

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/pkg/logging"
	"github.com/korobkovandrey/gmartloyalty/pkg/luhn"
	"go.uber.org/zap"
)

//go:generate mockgen -source=create.go -destination=mocks/mock_ordercreator.go -package=mocks
type orderCreator interface {
	Push(ctx context.Context, userID int64, orderNumber string) error
}

type AccrualPusher interface {
	PushOrder(orderNum string)
}

func NewCreateHandler(s orderCreator, a AccrualPusher) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		if userID == 0 {
			c.Status(http.StatusUnauthorized)
			return
		}
		body, err := c.GetRawData()
		if err != nil {
			_ = c.Error(err)
			c.Status(http.StatusBadRequest)
			return
		}
		orderNumber := strings.TrimSpace(string(body))
		if orderNumber == "" {
			c.Status(http.StatusBadRequest)
			return
		}
		logging.SetGinContextFields(c, zap.String("orderNumber", orderNumber))
		if !luhn.CheckString(orderNumber) {
			c.Status(http.StatusUnprocessableEntity)
			return
		}
		if err := s.Push(c, userID, orderNumber); err != nil {
			if errors.Is(err, service.ErrAlreadyExistsSomeUser) {
				c.Status(http.StatusOK)
			} else if errors.Is(err, service.ErrAlreadyExists) {
				c.Status(http.StatusConflict)
			} else {
				_ = c.Error(err)
				c.Status(http.StatusInternalServerError)
			}
			return
		}
		go a.PushOrder(orderNumber)
		c.Status(http.StatusAccepted)
	}
}
