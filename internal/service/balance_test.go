package service

import (
	"context"
	"errors"
	"testing"

	"github.com/korobkovandrey/gmartloyalty/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestBalance_Balance(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)

	t.Run("successful balance retrieval", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockBalanceStore(ctrl)
		mockStore.EXPECT().
			Accrual(ctx, userID).
			Return(100.50, nil)
		mockStore.EXPECT().
			Withdrawn(ctx, userID).
			Return(30.25, nil)

		balanceService := NewBalance(mockStore)
		response, err := balanceService.Balance(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 70.25, response.Current) // 100.50 - 30.25
		assert.Equal(t, 30.25, response.Withdrawn)
	})

	t.Run("accrual error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		accrualErr := errors.New("database error")
		mockStore := mocks.NewMockBalanceStore(ctrl)
		mockStore.EXPECT().
			Accrual(ctx, userID).
			Return(0.0, accrualErr)

		balanceService := NewBalance(mockStore)
		response, err := balanceService.Balance(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), accrualErr.Error())
		assert.Nil(t, response)
	})

	t.Run("withdrawn error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		withdrawnErr := errors.New("database error")
		mockStore := mocks.NewMockBalanceStore(ctrl)
		mockStore.EXPECT().
			Accrual(ctx, userID).
			Return(100.50, nil)
		mockStore.EXPECT().
			Withdrawn(ctx, userID).
			Return(0.0, withdrawnErr)

		balanceService := NewBalance(mockStore)
		response, err := balanceService.Balance(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), withdrawnErr.Error())
		assert.Nil(t, response)
	})

	t.Run("zero balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockBalanceStore(ctrl)
		mockStore.EXPECT().
			Accrual(ctx, userID).
			Return(0.0, nil)
		mockStore.EXPECT().
			Withdrawn(ctx, userID).
			Return(0.0, nil)

		balanceService := NewBalance(mockStore)
		response, err := balanceService.Balance(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 0.0, response.Current)
		assert.Equal(t, 0.0, response.Withdrawn)
	})
}
