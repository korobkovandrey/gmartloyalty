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

func TestWithdraw_Create(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	order := "12345"
	sum := 50.25

	t.Run("successful withdrawal creation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockWithdrawStore(ctrl)
		mockStore.EXPECT().
			Balance(ctx, userID).
			Return(100.50, nil)
		mockStore.EXPECT().
			CreateWithdraw(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, params query.CreateWithdrawParams) (int64, error) {
				assert.Equal(t, userID, params.UserID)
				assert.Equal(t, order, params.Order)
				assert.Equal(t, sum, params.Sum)
				assert.WithinDuration(t, time.Now().UTC(), params.ProcessedAt, 1*time.Second)
				return 1, nil
			})

		withdraw := NewWithdraw(mockStore)
		err := withdraw.Create(ctx, userID, order, sum)

		assert.NoError(t, err)
	})

	t.Run("insufficient balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockWithdrawStore(ctrl)
		mockStore.EXPECT().
			Balance(ctx, userID).
			Return(30.00, nil)

		withdraw := NewWithdraw(mockStore)
		err := withdraw.Create(ctx, userID, order, sum)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotEnoughBalance))
	})

	t.Run("error retrieving balance", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		balanceErr := errors.New("database error")
		mockStore := mocks.NewMockWithdrawStore(ctrl)
		mockStore.EXPECT().
			Balance(ctx, userID).
			Return(0.0, balanceErr)

		withdraw := NewWithdraw(mockStore)
		err := withdraw.Create(ctx, userID, order, sum)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), balanceErr.Error())
	})

	t.Run("error creating withdrawal", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		createErr := errors.New("database error")
		mockStore := mocks.NewMockWithdrawStore(ctrl)
		mockStore.EXPECT().
			Balance(ctx, userID).
			Return(100.50, nil)
		mockStore.EXPECT().
			CreateWithdraw(ctx, gomock.Any()).
			Return(int64(0), createErr)

		withdraw := NewWithdraw(mockStore)
		err := withdraw.Create(ctx, userID, order, sum)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), createErr.Error())
	})
}
