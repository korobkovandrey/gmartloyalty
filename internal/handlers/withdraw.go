package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/pkg/luhn"
)

//go:generate mockgen -source=withdraw.go -destination=mocks/mock_withdraw.go -package=mocks

type withdraw interface {
	Create(ctx context.Context, userID int64, order string, sum float64) error
}

func NewWithdrawHandler(s withdraw) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		if userID == 0 {
			c.Status(http.StatusUnauthorized)
			return
		}
		var form struct {
			Order string  `binding:"required"`
			Sum   float64 `binding:"required"`
		}
		if err := c.ShouldBind(&form); err != nil {
			_ = c.Error(err)
			c.Status(http.StatusBadRequest)
			return
		}
		if !luhn.CheckString(form.Order) {
			c.Status(http.StatusUnprocessableEntity)
			return
		}
		if err := s.Create(c, userID, form.Order, form.Sum); err != nil {
			_ = c.Error(err)
			if errors.Is(err, service.ErrNotEnoughBalance) {
				c.Status(http.StatusConflict)
			} else {
				c.Status(http.StatusInternalServerError)
			}
			return
		}
		c.Status(http.StatusOK)
	}
}
