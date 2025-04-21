package query

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
)

func TestQueriesIntegration_CreateWithdraw(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.CreateWithdraw(context.TODO(), CreateWithdrawParams{
			UserID:      dbtest.UserID1,
			Order:       dbtest.ProcessedOrderNum,
			Sum:         10,
			ProcessedAt: time.Now(),
		})
		assert.Error(t, err)
		assert.Equal(t, int64(0), result)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		result, err = q.CreateWithdraw(context.TODO(), CreateWithdrawParams{
			UserID:      dbtest.UserID1,
			Order:       dbtest.ProcessedOrderNum,
			Sum:         10,
			ProcessedAt: time.Now(),
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), result)
	})
}

func TestQueriesIntegration_GetWithdrawalsForUserID(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.GetWithdrawalsForUserID(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Len(t, result, 0)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		dbtest.SeedWithdraw(t, sqlDB)
		result, err = q.GetWithdrawalsForUserID(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		result, err = q.GetWithdrawalsForUserID(context.TODO(), dbtest.UserID2)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		_, err = q.db.ExecContext(context.TODO(), `DROP TABLE withdrawals;`)
		assert.NoError(t, err)
		_, err = q.GetWithdrawalsForUserID(context.TODO(), dbtest.UserID1)
		assert.Error(t, err)
	})
}
