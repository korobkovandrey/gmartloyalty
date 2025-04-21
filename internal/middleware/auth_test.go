package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/korobkovandrey/gmartloyalty/db/query"
	"github.com/korobkovandrey/gmartloyalty/internal/middleware/mocks"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestUserAuthJWT(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const secret = "secret"

	validToken1WithExp, err := makeBearerToken(t, secret, 1, time.Now().Add(time.Hour).Unix())
	require.NoError(t, err)
	validToken1Expired, err := makeBearerToken(t, secret, 1, time.Now().Add(-time.Hour).Unix())
	require.NoError(t, err)

	tests := []struct {
		name           string
		headers        map[string]string
		mockSetup      func(*mocks.Mockfinder)
		wantStatusCode int
		wantCtxUserID  int64
	}{
		{
			name: "401 without Authorisation header",
			mockSetup: func(s *mocks.Mockfinder) {
				s.EXPECT().Find(gomock.Any(), gomock.Any()).Times(0)
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "401 valid token, user not exists",
			headers: map[string]string{
				"Authorization": validToken1WithExp,
			},
			mockSetup: func(s *mocks.Mockfinder) {
				s.EXPECT().Find(gomock.Any(), int64(1)).Return(nil, service.ErrNotFound)
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "ok valid token, user exists",
			headers: map[string]string{
				"Authorization": validToken1WithExp,
			},
			mockSetup: func(s *mocks.Mockfinder) {
				s.EXPECT().Find(gomock.Any(), int64(1)).Return(&query.User{
					ID: 1,
				}, nil)
			},
			wantStatusCode: http.StatusOK,
			wantCtxUserID:  1,
		},
		{
			name: "401 expired token",
			headers: map[string]string{
				"Authorization": validToken1Expired,
			},
			mockSetup: func(s *mocks.Mockfinder) {
				s.EXPECT().Find(gomock.Any(), gomock.Any()).Times(0)
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name: "500 service error",
			headers: map[string]string{
				"Authorization": validToken1WithExp,
			},
			mockSetup: func(s *mocks.Mockfinder) {
				s.EXPECT().Find(gomock.Any(), int64(1)).Return(nil, errors.New("service error"))
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := mocks.NewMockfinder(ctrl)
			tt.mockSetup(s)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			for k, h := range tt.headers {
				ctx.Request.Header.Set(k, h)
			}
			UserAuthJWT([]byte(secret), s)(ctx)
			assert.Equal(t, tt.wantStatusCode, ctx.Writer.Status())
			assert.Equal(t, tt.wantCtxUserID, ctx.GetInt64("userID"))

			/* variant with engine
			_, engine := gin.CreateTestContext(w)
			engine.Use(UserAuthJWT([]byte(secret), s))
			engine.GET("/", func(c *gin.Context) {
				c.Status(http.StatusOK)
				assert.Equal(t, tt.wantCtxUserID, c.GetInt64("userID"))
			})
			engine.ServeHTTP(w, r)
			assert.Equal(t, tt.wantStatusCode, w.Result().StatusCode)
			*/
		})
	}
}

func makeBearerToken(t *testing.T, secret string, id, exp int64) (string, error) {
	t.Helper()
	claims := jwt.MapClaims{
		"userID": id,
	}
	if exp > 0 {
		claims["exp"] = exp
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tok, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to create signed token: %w", err)
	}
	return "Bearer " + tok, nil
}
