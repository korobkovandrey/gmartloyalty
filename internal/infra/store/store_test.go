package store

import (
	"context"
	"testing"

	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
)

func TestNewStoreIntegration(t *testing.T) {
	TestWithStore(t, nil)
}

func TestStoreIntegration_Balance(t *testing.T) {
	TestWithStore(t, func(t *testing.T, s *Store) {
		dbtest.SeedUsers(t, s.DB)
		dbtest.SeedOrders(t, s.DB)
		dbtest.SeedWithdraw(t, s.DB)
		balance, err := s.Balance(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Equal(t, 100.5+60.3-10.2-20.3, balance)
		_, err = s.DB.ExecContext(context.TODO(), `DROP TABLE withdrawals;`)
		assert.NoError(t, err)
		_, err = s.Balance(context.TODO(), dbtest.UserID1)
		assert.Error(t, err)
		_, err = s.DB.ExecContext(context.TODO(), `DROP TABLE orders;`)
		assert.NoError(t, err)
		_, err = s.Balance(context.TODO(), dbtest.UserID1)
		assert.Error(t, err)
	})
}
