package orders

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/handlers/orders/mocks"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		userID         int64
		requestBody    string
		mockSetup      func(*mocks.MockorderCreator, *mocks.MockAccrualPusher)
		wantStatusCode int
	}{
		{
			name:        "success",
			userID:      1,
			requestBody: dbtest.ProcessedOrderNum,
			mockSetup: func(oc *mocks.MockorderCreator, ap *mocks.MockAccrualPusher) {
				oc.EXPECT().
					Push(gomock.Any(), gomock.Eq(int64(1)), gomock.Eq(dbtest.ProcessedOrderNum)).
					Return(nil).
					Times(1)
				ap.EXPECT().
					PushOrder(gomock.Eq(dbtest.ProcessedOrderNum)).
					Times(1)
			},
			wantStatusCode: http.StatusAccepted,
		},
		{
			name:        "unauthorized",
			userID:      0,
			requestBody: dbtest.ProcessedOrderNum,
			mockSetup: func(oc *mocks.MockorderCreator, ap *mocks.MockAccrualPusher) {
				oc.EXPECT().
					Push(gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(0)
				ap.EXPECT().
					PushOrder(gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusUnauthorized,
		},
		{
			name:        "invalid",
			userID:      1,
			requestBody: "invalid",
			mockSetup: func(oc *mocks.MockorderCreator, ap *mocks.MockAccrualPusher) {
				oc.EXPECT().
					Push(gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(0)
				ap.EXPECT().
					PushOrder(gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:        "empty",
			userID:      1,
			requestBody: "",
			mockSetup: func(oc *mocks.MockorderCreator, ap *mocks.MockAccrualPusher) {
				oc.EXPECT().
					Push(gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(0)
				ap.EXPECT().
					PushOrder(gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:        "invalid order number",
			userID:      1,
			requestBody: "123",
			mockSetup: func(oc *mocks.MockorderCreator, ap *mocks.MockAccrualPusher) {
				oc.EXPECT().
					Push(gomock.Any(), gomock.Any(), gomock.Any()).
					MaxTimes(0)
				ap.EXPECT().
					PushOrder(gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:        "order already exists",
			userID:      1,
			requestBody: dbtest.ProcessedOrderNum,
			mockSetup: func(oc *mocks.MockorderCreator, ap *mocks.MockAccrualPusher) {
				oc.EXPECT().
					Push(gomock.Any(), gomock.Eq(int64(1)), gomock.Eq(dbtest.ProcessedOrderNum)).
					Return(service.ErrAlreadyExists).
					Times(1)
				ap.EXPECT().
					PushOrder(gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusConflict,
		},
		{
			name:        "order already exists some",
			userID:      1,
			requestBody: dbtest.ProcessedOrderNum,
			mockSetup: func(oc *mocks.MockorderCreator, ap *mocks.MockAccrualPusher) {
				oc.EXPECT().
					Push(gomock.Any(), gomock.Eq(int64(1)), gomock.Eq(dbtest.ProcessedOrderNum)).
					Return(service.ErrAlreadyExistsSomeUser).
					Times(1)
				ap.EXPECT().
					PushOrder(gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusOK,
		},
		{
			name:        "service error",
			userID:      1,
			requestBody: dbtest.ProcessedOrderNum,
			mockSetup: func(oc *mocks.MockorderCreator, ap *mocks.MockAccrualPusher) {
				oc.EXPECT().
					Push(gomock.Any(), gomock.Eq(int64(1)), gomock.Eq(dbtest.ProcessedOrderNum)).
					Return(errors.New("database error")).
					Times(1)
				ap.EXPECT().
					PushOrder(gomock.Any()).
					MaxTimes(0)
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body io.Reader
			if tt.requestBody != "" {
				body = strings.NewReader(tt.requestBody)
			} else {
				body = http.NoBody
			}
			mockOrderCreator := mocks.NewMockorderCreator(ctrl)
			mockAccrualPusher := mocks.NewMockAccrualPusher(ctrl)
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodPost, "/orders", body)
			if tt.userID > 0 {
				ctx.Set("userID", tt.userID)
			}
			tt.mockSetup(mockOrderCreator, mockAccrualPusher)
			NewCreateHandler(mockOrderCreator, mockAccrualPusher)(ctx)
			assert.Equal(t, tt.wantStatusCode, ctx.Writer.Status())
		})
	}
}
