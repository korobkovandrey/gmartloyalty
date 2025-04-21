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

	t.Run("successful order processing - PROCESSED", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/orders/3228087", r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"order":"3228087","status":"PROCESSED","accrual":100.50}`))
		}))
		defer server.Close()
		cfg.AccrualSystemAddress = server.URL
		mockStore := mocks.NewMockorderStore(ctrl)
		order := query.Order{
			ID:     1,
			Number: "3228087",
			Status: query.TOrderStatusNEW,
		}
		ctx := context.TODO()
		mockStore.EXPECT().GetOrderByNumber(ctx, gomock.Eq("3228087")).Return(order, nil)
		mockStore.EXPECT().SetOrderStatus(ctx, gomock.Eq(query.SetOrderStatusParams{
			ID:     1,
			Status: query.TOrderStatusPROCESSING,
		})).Return(nil)
		mockStore.EXPECT().SetOrderStatusAndAccrual(ctx, gomock.Eq(query.SetOrderStatusAndAccrualParams{
			ID:      1,
			Status:  query.TOrderStatusPROCESSED,
			Accrual: 100.50,
		})).Return(nil)
		accrual := NewAccrual(logger, mockStore, cfg)
		accrual.Run(ctx)
		accrual.PushOrder("3228087")
		time.Sleep(time.Second)
		accrual.Close()
	})

	t.Run("order already processed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Создаем тестовый HTTP-сервер (не должен вызываться)
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
