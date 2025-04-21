package orders

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
)

//go:generate mockgen -source=list.go -destination=mocks/mock_orderlister.go -package=mocks

type orderLister interface {
	List(ctx context.Context, userID int64) ([]service.OrderResponse, error)
}

func NewListHandler(s orderLister) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt64("userID")
		if userID == 0 {
			c.Status(http.StatusUnauthorized)
			return
		}
		l, err := s.List(c, userID)
		if err != nil {
			_ = c.Error(err)
			c.Status(http.StatusInternalServerError)
			return
		}
		if len(l) == 0 {
			c.Header("Content-Type", "application/json")
			c.Status(http.StatusNoContent)
			return
		}
		c.JSON(http.StatusOK, l)
	}
}
