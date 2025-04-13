package service

import (
	"context"
	"fmt"
	"time"

	"github.com/korobkovandrey/gmartloyalty/db/query"
)

type OrderListStore interface {
	GetOrdersForUserID(ctx context.Context, userID int64) ([]query.Order, error)
}

type OrdersLister struct {
	r OrderListStore
}

func NewOrderLister(r OrderListStore) *OrdersLister {
	return &OrdersLister{r: r}
}

type OrderResponse struct {
	Number     string             `json:"number"`
	Status     query.TOrderStatus `json:"status"`
	Accrual    float64            `json:"accrual,omitempty"`
	UploadedAt time.Time          `json:"uploaded_at"`
}

func (s *OrdersLister) List(ctx context.Context, userID int64) ([]OrderResponse, error) {
	l, err := s.r.GetOrdersForUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}
	result := make([]OrderResponse, len(l))
	for i, o := range l {
		result[i] = OrderResponse{
			Number:     o.Number,
			Status:     o.Status,
			Accrual:    o.Accrual,
			UploadedAt: o.UploadedAt,
		}
	}
	return result, nil
}
