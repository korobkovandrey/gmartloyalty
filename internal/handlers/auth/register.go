//nolint:dupl // ignore
package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/pkg/logging"
	"go.uber.org/zap"
)

type registerService interface {
	Register(ctx context.Context, login, password string) (string, error)
}

func NewRegisterHandler(s registerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var form struct {
			Login    string `binding:"required"`
			Password string `binding:"required"`
		}
		if err := c.ShouldBind(&form); err != nil {
			_ = c.Error(err)
			c.Status(http.StatusBadRequest)
			return
		}
		logging.SetGinContextFields(c, zap.String("login", form.Login))
		token, err := s.Register(c, form.Login, form.Password)
		if err != nil {
			_ = c.Error(err)
			if errors.Is(err, service.ErrAlreadyExists) {
				c.Status(http.StatusConflict)
				return
			}
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Header("Authorization", token)
		c.Status(http.StatusOK)
	}
}
