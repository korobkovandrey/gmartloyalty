package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/korobkovandrey/gmartloyalty/db/query"
	"github.com/korobkovandrey/gmartloyalty/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOrderCreator_Push(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	orderNumber := "12345"

	t.Run("successful order creation", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockOrderCreateStore(ctrl)
		mockStore.EXPECT().
			GetOrderByNumber(ctx, orderNumber).
			Return(query.Order{}, sql.ErrNoRows)
		mockStore.EXPECT().
			CreateOrder(ctx, gomock.Any()).
			DoAndReturn(func(_ context.Context, params query.CreateOrderParams) (int64, error) {
				assert.Equal(t, orderNumber, params.Number)
				assert.Equal(t, userID, params.UserID)
				assert.WithinDuration(t, time.Now().UTC(), params.UploadedAt, 1*time.Second)
				return 1, nil
			})

		creator := NewOrderCreator(mockStore)
		err := creator.Push(ctx, userID, orderNumber)

		assert.NoError(t, err)
	})

	t.Run("order already exists for same user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockOrderCreateStore(ctrl)
		mockStore.EXPECT().
			GetOrderByNumber(ctx, orderNumber).
			Return(query.Order{Number: orderNumber, UserID: userID}, nil)

		creator := NewOrderCreator(mockStore)
		err := creator.Push(ctx, userID, orderNumber)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrAlreadyExistsSomeUser))
		assert.Contains(t, err.Error(), orderNumber)
		assert.Contains(t, err.Error(), "userID: 1")
	})

	t.Run("order already exists for different user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockOrderCreateStore(ctrl)
		mockStore.EXPECT().
			GetOrderByNumber(ctx, orderNumber).
			Return(query.Order{Number: orderNumber, UserID: 2}, nil)

		creator := NewOrderCreator(mockStore)
		err := creator.Push(ctx, userID, orderNumber)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrAlreadyExists))
		assert.Contains(t, err.Error(), orderNumber)
		assert.Contains(t, err.Error(), "userID: 1")
	})

	t.Run("error checking order", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storeErr := errors.New("database error")
		mockStore := mocks.NewMockOrderCreateStore(ctrl)
		mockStore.EXPECT().
			GetOrderByNumber(ctx, orderNumber).
			Return(query.Order{}, storeErr)

		creator := NewOrderCreator(mockStore)
		err := creator.Push(ctx, userID, orderNumber)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), storeErr.Error())
		assert.Contains(t, err.Error(), orderNumber)
		assert.Contains(t, err.Error(), "userID: 1")
	})

	t.Run("error creating order", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storeErr := errors.New("database error")
		mockStore := mocks.NewMockOrderCreateStore(ctrl)
		mockStore.EXPECT().
			GetOrderByNumber(ctx, orderNumber).
			Return(query.Order{}, sql.ErrNoRows)
		mockStore.EXPECT().
			CreateOrder(ctx, gomock.Any()).
			Return(int64(0), storeErr)

		creator := NewOrderCreator(mockStore)
		err := creator.Push(ctx, userID, orderNumber)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), storeErr.Error())
		assert.Contains(t, err.Error(), orderNumber)
		assert.Contains(t, err.Error(), "userID: 1")
	})
}
