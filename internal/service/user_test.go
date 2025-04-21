package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/korobkovandrey/gmartloyalty/db/query"
	"github.com/korobkovandrey/gmartloyalty/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUserFinder_Find(t *testing.T) {
	ctx := context.Background()
	userID := int64(1)

	t.Run("successful user retrieval", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockUserFinderStore(ctrl)
		expectedUser := query.User{
			ID:       userID,
			Login:    "testuser",
			Password: "hashedpassword",
		}
		mockStore.EXPECT().
			GetUserById(ctx, userID).
			Return(expectedUser, nil)

		finder := NewUserFinder(mockStore)
		user, err := finder.Find(ctx, userID)

		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser, *user)
	})

	t.Run("user not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockStore := mocks.NewMockUserFinderStore(ctrl)
		mockStore.EXPECT().
			GetUserById(ctx, userID).
			Return(query.User{}, sql.ErrNoRows)

		finder := NewUserFinder(mockStore)
		user, err := finder.Find(ctx, userID)

		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrNotFound))
		assert.Nil(t, user)
	})

	t.Run("error retrieving user", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		storeErr := errors.New("database error")
		mockStore := mocks.NewMockUserFinderStore(ctrl)
		mockStore.EXPECT().
			GetUserById(ctx, userID).
			Return(query.User{}, storeErr)

		finder := NewUserFinder(mockStore)
		user, err := finder.Find(ctx, userID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), storeErr.Error())
		assert.Nil(t, user)
	})
}
