package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/korobkovandrey/gmartloyalty/db/query"
	"github.com/korobkovandrey/gmartloyalty/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestWithdrawals_List(t *testing.T) {
	ctx := context.Background()
	userID := int64(123)
	processedAt := time.Now()
	withdrawals := []query.Withdrawal{
		{
			Order:       "order1",
			Sum:         100.50,
			ProcessedAt: processedAt,
		},
		{
			Order:       "order2",
			Sum:         200.75,
			ProcessedAt: processedAt,
		},
	}

	expectedResponse := []WithdrawResponse{
		{
			Order:       "order1",
			Sum:         100.50,
			ProcessedAt: processedAt,
		},
		{
			Order:       "order2",
			Sum:         200.75,
			ProcessedAt: processedAt,
		},
	}

	t.Run("successful withdrawal list retrieval", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockStore := mocks.NewMockWithdrawalsStore(ctrl)
		mockStore.EXPECT().
			GetWithdrawalsForUserID(ctx, userID).
			Return(withdrawals, nil).
			Times(1)
		s := NewWithdrawals(mockStore)
		result, err := s.List(ctx, userID)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, result)
	})

	t.Run("error from store", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockStore := mocks.NewMockWithdrawalsStore(ctrl)
		expectedErr := errors.New("database error")
		mockStore.EXPECT().
			GetWithdrawalsForUserID(ctx, userID).
			Return(nil, expectedErr).
			Times(1)
		s := NewWithdrawals(mockStore)
		result, err := s.List(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get withdrawals: database error")
		assert.Nil(t, result)
	})

	t.Run("empty withdrawals list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		mockStore := mocks.NewMockWithdrawalsStore(ctrl)
		mockStore.EXPECT().
			GetWithdrawalsForUserID(ctx, userID).
			Return([]query.Withdrawal{}, nil).
			Times(1)
		s := NewWithdrawals(mockStore)
		result, err := s.List(ctx, userID)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}
