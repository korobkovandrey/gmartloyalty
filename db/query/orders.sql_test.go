package query

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
)

func TestQueriesIntegration_CreateOrder(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.CreateOrder(context.TODO(), CreateOrderParams{
			UserID:     dbtest.UserID1,
			Number:     dbtest.ProcessedOrderNum,
			UploadedAt: time.Now(),
		})
		assert.Error(t, err)
		assert.Equal(t, int64(0), result)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		result, err = q.CreateOrder(context.TODO(), CreateOrderParams{
			UserID:     dbtest.UserID1,
			Number:     dbtest.ProcessedOrderNum,
			UploadedAt: time.Now(),
		})
		assert.NoError(t, err)
		assert.Equal(t, int64(2), result)
		result, err = q.CreateOrder(context.TODO(), CreateOrderParams{
			UserID:     dbtest.UserID1,
			Number:     dbtest.ProcessedOrderNum,
			UploadedAt: time.Now(),
		})
		assert.Error(t, err)
		assert.Equal(t, int64(0), result)
	})
}

func TestQueriesIntegration_GetOrderByID(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.GetOrderByID(context.TODO(), 1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Equal(t, Order{}, result)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		dbtest.SeedOrders(t, sqlDB)
		result, err = q.GetOrderByID(context.TODO(), 1)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, int64(1), result.UserID)
		assert.Equal(t, dbtest.ProcessedOrderNum, result.Number)
		assert.Equal(t, TOrderStatusPROCESSED, result.Status)
		assert.Equal(t, 100.5, result.Accrual)
		result, err = q.GetOrderByID(context.TODO(), 10)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Equal(t, Order{}, result)
	})
}

func TestQueriesIntegration_GetOrderByNumber(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.GetOrderByNumber(context.TODO(), dbtest.ProcessedOrderNum)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Equal(t, Order{}, result)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		dbtest.SeedOrders(t, sqlDB)
		result, err = q.GetOrderByNumber(context.TODO(), dbtest.ProcessedOrderNum)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), result.ID)
		assert.Equal(t, int64(1), result.UserID)
		assert.Equal(t, dbtest.ProcessedOrderNum, result.Number)
		assert.Equal(t, TOrderStatusPROCESSED, result.Status)
		assert.Equal(t, 100.5, result.Accrual)
		result, err = q.GetOrderByNumber(context.TODO(), "invalid_num")
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Equal(t, Order{}, result)
	})
}

func TestQueriesIntegration_GetOrdersForUserID(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.GetOrdersForUserID(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Len(t, result, 0)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		dbtest.SeedOrders(t, sqlDB)
		result, err = q.GetOrdersForUserID(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Len(t, result, 5)
		result, err = q.GetOrdersForUserID(context.TODO(), dbtest.UserID2)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		_, err = q.db.ExecContext(context.TODO(), `DROP TABLE orders;`)
		assert.NoError(t, err)
		_, err = q.GetOrdersForUserID(context.TODO(), dbtest.UserID1)
		assert.Error(t, err)
	})
}

func TestQueriesIntegration_GetOrdersNotProcessed(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		result, err := q.GetOrdersNotProcessed(context.TODO())
		assert.NoError(t, err)
		assert.Len(t, result, 0)
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		dbtest.SeedOrders(t, sqlDB)
		result, err = q.GetOrdersNotProcessed(context.TODO())
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		_, err = q.db.ExecContext(context.TODO(), `DROP TABLE orders;`)
		assert.NoError(t, err)
		_, err = q.GetOrdersNotProcessed(context.TODO())
		assert.Error(t, err)
	})
}

func TestQueriesIntegration_SetOrderStatus(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		dbtest.SeedOrders(t, sqlDB)
		err := q.SetOrderStatus(context.TODO(), SetOrderStatusParams{
			Status: TOrderStatusNEW,
			ID:     1,
		})
		assert.NoError(t, err)
		result, err := q.GetOrderByID(context.TODO(), 1)
		assert.NoError(t, err)
		assert.Equal(t, TOrderStatusNEW, result.Status)
	})
}

func TestQueriesIntegration_SetOrderStatusAndAccrual(t *testing.T) {
	testWithQueries(t, func(t *testing.T, q *Queries) {
		sqlDB := q.db.(*sql.DB)
		dbtest.SeedUsers(t, sqlDB)
		dbtest.SeedOrders(t, sqlDB)
		err := q.SetOrderStatusAndAccrual(context.TODO(), SetOrderStatusAndAccrualParams{
			Status:  TOrderStatusNEW,
			ID:      1,
			Accrual: 300,
		})
		assert.NoError(t, err)
		result, err := q.GetOrderByID(context.TODO(), 1)
		assert.NoError(t, err)
		assert.Equal(t, TOrderStatusNEW, result.Status)
		assert.Equal(t, float64(300), result.Accrual)
	})
}
