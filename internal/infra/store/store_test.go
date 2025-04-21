package store

import (
	"context"
	"testing"

	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestStore(t *testing.T, pool *dockertest.Pool, dsn string) (s *Store, err error) {
	t.Helper()
	err = pool.Retry(func() (err error) {
		s, err = NewStore(context.TODO(), dsn)
		return err
	})
	return s, err
}

func testWithStore(t *testing.T, f func(t *testing.T, s *Store)) {
	t.Helper()
	dsn, resource, pool := dbtest.MakeDockerPoolDB(t)
	defer func() {
		assert.NoError(t, resource.Close())
	}()
	require.NoError(t, resource.Expire(60))
	s, err := makeTestStore(t, pool, dsn)
	defer func() {
		assert.NoError(t, s.Close())
	}()
	require.NoError(t, err)
	if f != nil {
		f(t, s)
	}
}

func TestNewStore(t *testing.T) {
	testWithStore(t, nil)
}

func TestStore_Balance(t *testing.T) {
	testWithStore(t, func(t *testing.T, s *Store) {
		dbtest.SeedUsers(t, s.db)
		dbtest.SeedOrders(t, s.db)
		dbtest.SeedWithdraw(t, s.db)
		balance, err := s.Balance(context.TODO(), dbtest.UserID1)
		assert.NoError(t, err)
		assert.Equal(t, 100.5+60.3-10.2-20.3, balance)
	})
}
