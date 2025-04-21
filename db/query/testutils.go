package query

import (
	"database/sql"
	"testing"
	"time"

	"github.com/korobkovandrey/gmartloyalty/db"
	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeTestQueries(t *testing.T, pool *dockertest.Pool, dsn string) (q *Queries) {
	t.Helper()
	require.NoError(t, pool.Retry(func() (err error) {
		return db.Migrate(dsn)
	}))
	d, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	d.SetMaxOpenConns(5)
	d.SetMaxIdleConns(5)
	d.SetConnMaxIdleTime(time.Minute)
	d.SetConnMaxLifetime(time.Minute)
	require.NoError(t, d.Ping())
	return New(d)
}

func testWithQueries(t *testing.T, f func(t *testing.T, q *Queries)) {
	t.Helper()
	dsn, resource, pool := dbtest.MakeDockerPoolDB(t)
	defer func() {
		assert.NoError(t, resource.Close())
	}()
	//nolint:mnd // ignore
	require.NoError(t, resource.Expire(60))
	q := makeTestQueries(t, pool, dsn)
	defer func() {
		assert.NoError(t, q.db.(*sql.DB).Close())
	}()
	if f != nil {
		f(t, q)
	}
}
