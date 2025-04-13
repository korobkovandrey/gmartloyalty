package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/korobkovandrey/gmartloyalty/pkg/logging"
	"go.uber.org/zap"
)

func UserAuthJWT(secret []byte) gin.HandlerFunc {
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
		c.Set("userID", userIDInt64)
		logging.SetGinContextFields(c, zap.Int64("userID", userIDInt64))
		c.Next()
	}
}
