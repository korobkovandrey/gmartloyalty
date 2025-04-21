package orders

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/handlers/orders/mocks"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		userID         int64
		requestBody    string
		mockSetup      func(*mocks.MockorderLister)
		wantStatusCode int
		wantBodyJSON   string
	}{
		{
			name:        "success",
			userID:      1,
			requestBody: "",
			mockSetup: func(m *mocks.MockorderLister) {
				uploadedAt := time.Now().UTC().Round(time.Second)
				m.EXPECT().
					List(gomock.Any(), gomock.Eq(int64(1))).
					Return([]service.OrderResponse{
						{
							Number:     dbtest.ProcessedOrderNum,
							Status:     "PROCESSED",
							Accrual:    100.50,
							UploadedAt: uploadedAt,
						},
					}, nil).
					Times(1)
			},
			wantStatusCode: http.StatusOK,
			wantBodyJSON:   `[{"number":"` + dbtest.ProcessedOrderNum + `","status":"PROCESSED","accrual":100.5,"uploaded_at":"%s"}]`,
		},
		{
			name:        "empty orders",
			userID:      1,
			requestBody: "",
			mockSetup: func(m *mocks.MockorderLister) {
				m.EXPECT().
					List(gomock.Any(), gomock.Eq(int64(1))).
					Return([]service.OrderResponse{}, nil).
					Times(1)
			},
			wantStatusCode: http.StatusNoContent,
			wantBodyJSON:   "",
		},
		{
			name:        "unauthorized",
			userID:      0,
			requestBody: "",
			mockSetup: func(m *mocks.MockorderLister) {
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
			requestBody: "",
			mockSetup: func(m *mocks.MockorderLister) {
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
			if tt.requestBody != "" {
				body = strings.NewReader(tt.requestBody)
			} else {
				body = http.NoBody
			}
			mockOrderLister := mocks.NewMockorderLister(ctrl)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/orders", body)
			if tt.userID > 0 {
				ctx.Set("userID", tt.userID)
			}
			tt.mockSetup(mockOrderLister)
			NewListHandler(mockOrderLister)(ctx)
			assert.Equal(t, tt.wantStatusCode, ctx.Writer.Status(), "unexpected status code")
			if tt.wantBodyJSON != "" {
				if strings.Contains(tt.wantBodyJSON, "%s") {
					uploadedAt := time.Now().UTC().Round(time.Second).Format(time.RFC3339)
					expectedJSON := fmt.Sprintf(tt.wantBodyJSON, uploadedAt)
					assert.JSONEq(t, expectedJSON, w.Body.String(), "unexpected response body")
				} else {
					assert.JSONEq(t, tt.wantBodyJSON, w.Body.String(), "unexpected response body")
				}
			} else {
				assert.Empty(t, w.Body.String(), "expected empty response body")
			}
		})
	}
}
