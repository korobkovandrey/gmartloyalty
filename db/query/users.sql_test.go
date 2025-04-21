package query

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
)

func TestQueriesIntegration_CreateUser(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.CreateUser(context.TODO(), CreateUserParams{
			Login:     "user1",
			Password:  "",
			CreatedAt: time.Now(),
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result)
		result, err = q.CreateUser(context.TODO(), CreateUserParams{
			Login:     "user1",
			Password:  "",
			CreatedAt: time.Now(),
		})
		assert.Error(t, err)
		assert.Equal(t, int64(0), result)
	})
}

func TestQueriesIntegration_GetUserById(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.GetUserById(context.TODO(), 1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Equal(t, User{}, result)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		result, err = q.GetUserById(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Equal(t, int64(dbtest.UserID1), result.ID)
		assert.Equal(t, "user1", result.Login)
	})
}

func TestQueriesIntegration_GetUserByLogin(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.GetUserByLogin(context.TODO(), "user1")
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Equal(t, User{}, result)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		result, err = q.GetUserByLogin(context.TODO(), "user1")
		assert.NoError(t, err)
		assert.Equal(t, int64(dbtest.UserID1), result.ID)
		assert.Equal(t, "user1", result.Login)
	})
}
