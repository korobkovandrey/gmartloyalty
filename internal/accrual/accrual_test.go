package accrual

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/korobkovandrey/gmartloyalty/db/query"
	"github.com/korobkovandrey/gmartloyalty/internal/accrual/mocks"
	"github.com/korobkovandrey/gmartloyalty/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zapcore"
)

func TestAccrual(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	cfg := &Config{
		AccrualSystemAddress: "",
		JobsSize:             10,
		DeferJobsSize:        10,
		NumWorkers:           1,
		MaxAttempts:          3,
		AttemptTimeout:       100 * time.Millisecond,
	}
	logger, err := logging.NewZapLogger(zapcore.FatalLevel, []string{"stderr"})
	require.NoError(t, err)

	tests := []struct {
		name            string
		orderID         int64
		orderNum        string
		accrualResponse string
		wantAccrual     float64
		wantOrderStatus query.TOrderStatus
	}{
		{
			name:            "successful order processing - PROCESSED",
			orderID:         1,
			orderNum:        "3228087",
			accrualResponse: `{"order":"3228087","status":"PROCESSED","accrual":100.50}`,
			wantAccrual:     100.50,
			wantOrderStatus: query.TOrderStatusPROCESSED,
		},
		{
			name:            "successful order processing - INVALID",
			orderID:         1,
			orderNum:        "3228087",
			accrualResponse: `{"order":"3228087","status":"INVALID"}`,
			wantAccrual:     0,
			wantOrderStatus: query.TOrderStatusINVALID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/orders/"+tt.orderNum, r.URL.Path)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(tt.accrualResponse))
			}))
			defer server.Close()
			cfg.AccrualSystemAddress = server.URL
			mockStore := mocks.NewMockorderStore(ctrl)
			order := query.Order{
				ID:     tt.orderID,
				Number: tt.orderNum,
				Status: query.TOrderStatusNEW,
			}
			ctx := context.TODO()
			mockStore.EXPECT().GetOrderByNumber(ctx, gomock.Eq("3228087")).Return(order, nil)
			mockStore.EXPECT().SetOrderStatus(ctx, gomock.Eq(query.SetOrderStatusParams{
				ID:     tt.orderID,
				Status: query.TOrderStatusPROCESSING,
			})).Return(nil)
			mockStore.EXPECT().SetOrderStatusAndAccrual(ctx, gomock.Eq(query.SetOrderStatusAndAccrualParams{
				ID:      tt.orderID,
				Status:  tt.wantOrderStatus,
				Accrual: tt.wantAccrual,
			})).Return(nil)
			accrual := NewAccrual(logger, mockStore, cfg)
			accrual.Run(ctx)
			accrual.PushOrder(tt.orderNum)
			time.Sleep(time.Second)
			accrual.Close()
		})
	}

	t.Run("order already processed", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("HTTP server should not be called")
		}))
		defer server.Close()
		cfg.AccrualSystemAddress = server.URL
		mockStore := mocks.NewMockorderStore(ctrl)
		order := query.Order{
			ID:     1,
			Number: "3228087",
			Status: query.TOrderStatusPROCESSED,
		}
		ctx := context.TODO()
		mockStore.EXPECT().GetOrderByNumber(ctx, "3228087").Return(order, nil)
		accrual := NewAccrual(logger, mockStore, cfg)
		accrual.Run(ctx)
		accrual.PushOrder("3228087")
		time.Sleep(time.Second)
		accrual.Close()
	})

	t.Run("getOrderToProcess error - no rows", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("HTTP server should not be called")
		}))
		defer server.Close()
		cfg.AccrualSystemAddress = server.URL
		mockStore := mocks.NewMockorderStore(ctrl)
		ctx := context.TODO()
		mockStore.EXPECT().GetOrderByNumber(ctx, "3228087").Return(query.Order{}, sql.ErrNoRows)
		accrual := NewAccrual(logger, mockStore, cfg)
		accrual.Run(ctx)
		accrual.PushOrder("3228087")
		time.Sleep(time.Second)
		accrual.Close()
	})
}
