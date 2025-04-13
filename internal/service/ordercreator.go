package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/korobkovandrey/gmartloyalty/db/query"
)

type OrderCreateStore interface {
	GetOrderByNumber(ctx context.Context, number string) (query.Order, error)
	CreateOrder(ctx context.Context, arg query.CreateOrderParams) (int64, error)
}

type OrderCreator struct {
	r OrderCreateStore
}

func NewOrderCreator(r OrderCreateStore) *OrderCreator {
	return &OrderCreator{r: r}
}

func (s *OrderCreator) Push(ctx context.Context, userID int64, orderNumber string) error {
	o, err := s.r.GetOrderByNumber(ctx, orderNumber)
	if err == nil {
		if o.UserID == userID {
			err = ErrAlreadyExistsSomeUser
		} else {
			err = ErrAlreadyExists
		}
		return fmt.Errorf("failed to create order: %w, orderNumber: %s, userID: %d", err, orderNumber, userID)
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to check order: %w, orderNumber: %s, userID: %d", err, orderNumber, userID)
	}
	_, err = s.r.CreateOrder(ctx, query.CreateOrderParams{
		Number:     orderNumber,
		UserID:     userID,
		UploadedAt: time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to create order: %w, orderNumber: %s, userID: %d", err, orderNumber, userID)
	}
	return nil
}
