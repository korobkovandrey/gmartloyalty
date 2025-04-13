package service

import (
	"context"
	"fmt"
)

type BalanceStore interface {
	Accrual(ctx context.Context, userID int64) (float64, error)
	Withdrawn(ctx context.Context, userID int64) (float64, error)
}

type Balance struct {
	r BalanceStore
}

func NewBalance(r BalanceStore) *Balance {
	return &Balance{r: r}
}

type BalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

func (s *Balance) Balance(ctx context.Context, userID int64) (*BalanceResponse, error) {
	a, err := s.r.Accrual(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accrual: %w", err)
	}
	w, err := s.r.Withdrawn(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get withdrawn: %w", err)
	}
	return &BalanceResponse{
		Current:   a - w,
		Withdrawn: w,
	}, nil
}
