package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/handlers/mocks"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestWithdrawals(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		userID         int64
		requestBody    interface{}
		mockSetup      func(*mocks.Mockwithdrawals)
		wantStatusCode int
		wantBodyJSON   string
	}{
		{
			name:        "success",
			userID:      1,
			requestBody: nil,
			mockSetup: func(m *mocks.Mockwithdrawals) {
				processedAt := time.Now().UTC().Round(time.Second)
				m.EXPECT().
					List(gomock.Any(), gomock.Eq(int64(1))).
					Return([]service.WithdrawResponse{
						{
							Order:       dbtest.ProcessedOrderNum,
							Sum:         50.75,
							ProcessedAt: processedAt,
						},
					}, nil).
					Times(1)
			},
			wantStatusCode: http.StatusOK,
			wantBodyJSON:   `[{"order":"` + dbtest.ProcessedOrderNum + `","sum":50.75,"processed_at":"%s"}]`,
		},
		{
			name:        "empty withdrawals",
			userID:      1,
			requestBody: nil,
			mockSetup: func(m *mocks.Mockwithdrawals) {
				m.EXPECT().
					List(gomock.Any(), gomock.Eq(int64(1))).
					Return([]service.WithdrawResponse{}, nil).
					Times(1)
			},
			wantStatusCode: http.StatusNoContent,
		},
		{
			name:        "unauthorized",
			userID:      0,
			requestBody: nil,
			mockSetup: func(m *mocks.Mockwithdrawals) {
				m.EXPECT().
					List(gomock.Any(), gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantBodyJSON:   "",
		},
		{
			name:        "service error",
			userID:      1,
			requestBody: nil,
			mockSetup: func(m *mocks.Mockwithdrawals) {
				m.EXPECT().
					List(gomock.Any(), gomock.Eq(int64(1))).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			wantStatusCode: http.StatusInternalServerError,
			wantBodyJSON:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Подготавливаем тело запроса
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
			mockWithdrawals := mocks.NewMockwithdrawals(ctrl)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/withdrawals", body)
			ctx.Request.Header.Set("Content-Type", "application/json")
			if tt.userID > 0 {
				ctx.Set("userID", tt.userID)
			}
			tt.mockSetup(mockWithdrawals)
			NewWithdrawalsHandler(mockWithdrawals)(ctx)
			assert.Equal(t, tt.wantStatusCode, ctx.Writer.Status())
			if tt.wantBodyJSON != "" {
				if strings.Contains(tt.wantBodyJSON, "%s") {
					processedAt := time.Now().UTC().Round(time.Second).Format(time.RFC3339)
					expectedJSON := fmt.Sprintf(tt.wantBodyJSON, processedAt)
					assert.JSONEq(t, expectedJSON, w.Body.String())
				} else {
					assert.JSONEq(t, tt.wantBodyJSON, w.Body.String())
				}
			} else {
				assert.Empty(t, w.Body.String())
			}
		})
	}
}
