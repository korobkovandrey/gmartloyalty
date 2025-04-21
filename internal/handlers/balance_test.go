package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/handlers/mocks"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBalance(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	userID1 := new(int64)
	*userID1 = 1
	tests := []struct {
		name           string
		userID         int64
		serviceBalance *service.BalanceResponse
		serviceErr     error
		wantStatusCode int
		wantBodyJSON   string
	}{
		{
			name:           "success",
			userID:         1,
			serviceBalance: &service.BalanceResponse{Current: 100.50, Withdrawn: 20.25},
			wantStatusCode: http.StatusOK,
			wantBodyJSON:   `{"current":100.5,"withdrawn":20.25}`,
		},
		{
			name:           "unauthorized",
			userID:         0,
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:           "service error",
			userID:         1,
			serviceBalance: nil,
			serviceErr:     errors.New("database error"),
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем мок
			mockBalance := mocks.NewMockbalance(ctrl)

			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/balance", http.NoBody)

			if tt.userID > 0 {
				ctx.Set("userID", tt.userID)
				mockBalance.EXPECT().
					Balance(gomock.Any(), gomock.Eq(tt.userID)).
					Return(tt.serviceBalance, tt.serviceErr).
					Times(1)
			} else {
				mockBalance.EXPECT().
					Balance(gomock.Any(), gomock.Any()).MaxTimes(0)
			}
			NewBalanceHandler(mockBalance)(ctx)
			assert.Equal(t, tt.wantStatusCode, ctx.Writer.Status())
			if tt.wantBodyJSON != "" {
				assert.JSONEq(t, tt.wantBodyJSON, w.Body.String())
			}
		})
	}
}
