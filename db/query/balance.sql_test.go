package query

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
)

func TestQueriesIntegration_Withdrawn(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.Withdrawn(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, result)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		dbtest.SeedOrders(t, sqlDB)
		dbtest.SeedWithdraw(t, sqlDB)
		result, err = q.Withdrawn(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Equal(t, 10.2+20.3, result)
	})
}

func TestQueriesIntegration_Accrual(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.Accrual(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Equal(t, 0.0, result)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		dbtest.SeedOrders(t, sqlDB)
		dbtest.SeedWithdraw(t, sqlDB)
		result, err = q.Accrual(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Equal(t, 100.5+60.3, result)
	})
}
