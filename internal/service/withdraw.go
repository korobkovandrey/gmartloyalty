package service

import (
	"context"
	"fmt"
	"time"

	"github.com/korobkovandrey/gmartloyalty/db/query"
)

type WithdrawStore interface {
	Balance(ctx context.Context, userID int64) (float64, error)
	CreateWithdraw(ctx context.Context, arg query.CreateWithdrawParams) (int64, error)
}

type Withdraw struct {
	r WithdrawStore
}

func NewWithdraw(r WithdrawStore) *Withdraw {
	return &Withdraw{r: r}
}

func (s *Withdraw) Create(ctx context.Context, userID int64, order string, sum float64) error {
	b, err := s.r.Balance(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get balance: %w", err)
	}
	if sum > b {
		return ErrNotEnoughBalance
	}
	_, err = s.r.CreateWithdraw(ctx, query.CreateWithdrawParams{
		UserID:      userID,
		Sum:         sum,
		Order:       order,
		ProcessedAt: time.Now().UTC(),
	})
	if err != nil {
		return fmt.Errorf("failed to create withdraw: %w", err)
	}
	return nil
}
