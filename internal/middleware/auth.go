package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/korobkovandrey/gmartloyalty/db/query"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/pkg/logging"
	"go.uber.org/zap"
)

//go:generate mockgen -source=auth.go -destination=mocks/mock_auth.go -package=mocks

type finder interface {
	Find(context.Context, int64) (*query.User, error)
}

func UserAuthJWT(secret []byte, s finder) gin.HandlerFunc {
	return func(c *gin.Context) {
		const prefix = "Bearer "
		ah := c.Request.Header.Get("Authorization")
		if ah == "" || !strings.HasPrefix(ah, prefix) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(ah[len(prefix):], &claims,
			func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return secret, nil
			})
		if err != nil {
			_ = c.Error(err)
		}
		if err != nil || !token.Valid {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userID, ok := claims["userID"].(float64)
		if !ok {
			_ = c.Error(fmt.Errorf("invalid userID: %v", claims))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userIDInt64 := int64(userID)
		if _, err = s.Find(c, userIDInt64); err != nil {
			_ = c.Error(err)
			if errors.Is(err, service.ErrNotFound) {
				c.AbortWithStatus(http.StatusUnauthorized)
			} else {
				c.AbortWithStatus(http.StatusInternalServerError)
			}
			return
		}
		c.Set("userID", userIDInt64)
		logging.SetGinContextFields(c, zap.Int64("userID", userIDInt64))
		c.Next()
	}
}
