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

// TestOrdersLister_List tests the List method of the OrdersLister struct.
func TestOrdersLister_List(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)
	uploadedAt := time.Now().UTC()

	t.Run("successful order listing", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockOrderListStore(ctrl)
		orders := []query.Order{
			{
				Number:     "12345",
				Status:     query.TOrderStatusPROCESSED,
				Accrual:    100.50,
				UploadedAt: uploadedAt,
			},
			{
				Number:     "67890",
				Status:     query.TOrderStatusNEW,
				Accrual:    0.0,
				UploadedAt: uploadedAt,
			},
		}
		mockStore.EXPECT().
			GetOrdersForUserID(ctx, userID).
			Return(orders, nil)

		lister := NewOrderLister(mockStore)
		response, err := lister.List(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response, 2)
		assert.Equal(t, "12345", response[0].Number)
		assert.Equal(t, query.TOrderStatusPROCESSED, response[0].Status)
		assert.Equal(t, 100.50, response[0].Accrual)
		assert.Equal(t, uploadedAt, response[0].UploadedAt)
		assert.Equal(t, "67890", response[1].Number)
		assert.Equal(t, query.TOrderStatusNEW, response[1].Status)
		assert.Equal(t, 0.0, response[1].Accrual)
		assert.Equal(t, uploadedAt, response[1].UploadedAt)
	})

	t.Run("empty order list", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockOrderListStore(ctrl)
		mockStore.EXPECT().
			GetOrdersForUserID(ctx, userID).
			Return([]query.Order{}, nil)

		lister := NewOrderLister(mockStore)
		response, err := lister.List(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Empty(t, response)
	})

	t.Run("error retrieving orders", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storeErr := errors.New("database error")
		mockStore := mocks.NewMockOrderListStore(ctrl)
		mockStore.EXPECT().
			GetOrdersForUserID(ctx, userID).
			Return(nil, storeErr)

		lister := NewOrderLister(mockStore)
		response, err := lister.List(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), storeErr.Error())
		assert.Nil(t, response)
	})
}
