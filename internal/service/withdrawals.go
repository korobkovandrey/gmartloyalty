package service

import (
	"context"
	"fmt"
	"time"

	"github.com/korobkovandrey/gmartloyalty/db/query"
)

type WithdrawalsStore interface {
	GetWithdrawalsForUserID(ctx context.Context, userID int64) ([]query.Withdrawal, error)
}

type Withdrawals struct {
	r WithdrawalsStore
}

func NewWithdrawals(r WithdrawalsStore) *Withdrawals {
	return &Withdrawals{r: r}
}

type WithdrawResponse struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (s *Withdrawals) List(ctx context.Context, userID int64) ([]WithdrawResponse, error) {
	l, err := s.r.GetWithdrawalsForUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawals: %w", err)
	}
	result := make([]WithdrawResponse, len(l))
	for i, w := range l {
		result[i] = WithdrawResponse{
			Order:       w.Order,
			Sum:         w.Sum,
			ProcessedAt: w.ProcessedAt,
		}
	}
	return result, nil
}
