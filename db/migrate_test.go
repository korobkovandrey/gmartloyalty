package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/korobkovandrey/gmartloyalty/test/dbtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB_MigrateIntegration(t *testing.T) {
	dsn, resource, pool := dbtest.MakeDockerPoolDB(t)
	defer func() {
		assert.NoError(t, resource.Close())
	}()
	require.NoError(t, resource.Expire(60))

	require.NoError(t, pool.Retry(func() error {
		return Migrate(dsn)
	}))
	db, err := sql.Open("pgx", dsn)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, db.Close())
	}()
	ctx := context.TODO()
	require.NoError(t, db.PingContext(ctx))
	tables := []string{"users", "orders", "withdrawals"}
	for _, table := range tables {
		var exists bool
		err = db.QueryRowContext(ctx,
			fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '%s')", table)).
			Scan(&exists)
		require.NoError(t, err)
		assert.True(t, exists)
	}
}
