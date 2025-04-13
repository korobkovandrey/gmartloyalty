package orders

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/pkg/logging"
	"go.uber.org/zap"
)

type orderCreator interface {
	Push(ctx context.Context, userID int64, orderNumber string) error
}

func NewCreateHandler(s orderCreator) gin.HandlerFunc {
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
		if !checkLuhn(orderNumber) {
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
		c.Status(http.StatusAccepted)
	}
}

func checkLuhn(s string) bool {
	number, err := strconv.Atoi(s)
	if err != nil {
		return false
	}
	var sum int
	for i := 0; number > 0; i++ {
		cur := number % 10
		if i%2 == 0 {
			sum += cur
			number /= 10
			continue
		}
		cur *= 2
		//nolint:mnd // ignore
		if cur > 9 {
			cur -= 9
		}
		sum += cur
		number /= 10
	}
	return sum%10 == 0
}
