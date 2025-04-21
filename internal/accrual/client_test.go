package accrual

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClient_getOrder(t *testing.T) {
	ctx := context.Background()
	order := "12345"

	t.Run("successful order retrieval", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/orders/"+order, r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"order":"12345","status":"PROCESSED","accrual":100.50}`))
		}))
		defer server.Close()
		c := newClient(server.URL)
		defer c.close()
		result, err := c.getOrder(ctx, order)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, &orderResponse{
			Order:   "12345",
			Status:  "PROCESSED",
			Accrual: 100.50,
		}, result)
	})

	t.Run("too many requests with retry-after", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Retry-After", "1")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{}`))
		}))
		defer server.Close()
		c := newClient(server.URL)
		defer c.close()
		result, err := c.getOrder(ctx, order)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, &orderResponse{
			Order:      "",
			RetryAfter: 1,
		}, result)
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"order":"12345","status":"PROCESSED","accrual":100.50}`))
		}))
		defer server.Close()
		c := newClient(server.URL)
		defer c.close()
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()
		result, err := c.getOrder(ctx, order)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get order")
		assert.Contains(t, err.Error(), context.DeadlineExceeded.Error())
		assert.Nil(t, result)
	})
}
