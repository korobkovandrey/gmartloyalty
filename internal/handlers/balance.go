package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
)

type balance interface {
	Balance(ctx context.Context, userID int64) (*service.BalanceResponse, error)
}

func NewBalanceHandler(s balance) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		if userID == 0 {
			c.Status(http.StatusUnauthorized)
			return
		}
		b, err := s.Balance(c, userID)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, b)
	}
}
