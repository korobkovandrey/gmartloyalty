package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/handlers/mocks"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestWithdraw(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		userID         int64
		requestBody    interface{}
		mockSetup      func(*mocks.Mockwithdraw)
		wantStatusCode int
	}{
		{
			name:        "success",
			userID:      1,
			requestBody: map[string]interface{}{"order": dbtest.ProcessedOrderNum, "sum": 50.75},
			mockSetup: func(m *mocks.Mockwithdraw) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Eq(int64(1)), gomock.Eq(dbtest.ProcessedOrderNum), gomock.Eq(50.75)).
					Return(nil).
					Times(1)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "unauthorized",
			userID:      0,
			requestBody: map[string]interface{}{"order": dbtest.ProcessedOrderNum, "sum": 50.75},
			mockSetup: func(m *mocks.Mockwithdraw) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:        "invalid json",
			userID:      1,
			requestBody: "invalid json",
			mockSetup: func(m *mocks.Mockwithdraw) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:        "invalid request",
			userID:      1,
			requestBody: map[string]interface{}{"invalid": "form"},
			mockSetup: func(m *mocks.Mockwithdraw) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:        "invalid order",
			userID:      1,
			requestBody: map[string]interface{}{"order": "123", "sum": 50.75},
			mockSetup: func(m *mocks.Mockwithdraw) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:        "insufficient balance",
			userID:      1,
			requestBody: map[string]interface{}{"order": dbtest.ProcessedOrderNum, "sum": 50.75},
			mockSetup: func(m *mocks.Mockwithdraw) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Eq(int64(1)), gomock.Eq(dbtest.ProcessedOrderNum), gomock.Eq(50.75)).
					Return(service.ErrNotEnoughBalance).
					Times(1)
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name:        "service error",
			userID:      1,
			requestBody: map[string]interface{}{"order": dbtest.ProcessedOrderNum, "sum": 50.75},
			mockSetup: func(m *mocks.Mockwithdraw) {
				m.EXPECT().
					Create(gomock.Any(), gomock.Eq(int64(1)), gomock.Eq(dbtest.ProcessedOrderNum), gomock.Eq(50.75)).
					Return(errors.New("database error")).
					Times(1)
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.requestBody != nil {
				if str, ok := tt.requestBody.(string); ok {
					body = strings.NewReader(str)
				} else {
					bodyBytes, err := json.Marshal(tt.requestBody)
					require.NoError(t, err)
					body = bytes.NewReader(bodyBytes)
				}
			} else {
				body = http.NoBody
			}
			mockWithdraw := mocks.NewMockwithdraw(ctrl)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/withdraw", body)
			ctx.Request.Header.Set("Content-Type", "application/json")
			if tt.userID > 0 {
				ctx.Set("userID", tt.userID)
			}
			tt.mockSetup(mockWithdraw)
			NewWithdrawHandler(mockWithdraw)(ctx)
			assert.Equal(t, tt.wantStatusCode, ctx.Writer.Status())
		})
	}
}
