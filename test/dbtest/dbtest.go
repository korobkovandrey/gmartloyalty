package dbtest

import (
	"testing"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"
)

func MakeDockerPoolDB(t *testing.T) (dsn string, resource *dockertest.Resource, pool *dockertest.Pool) {
	t.Helper()
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)
	require.NoError(t, pool.Client.Ping())
	resource, err = pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "17.4-alpine",
		Env: []string{
			"POSTGRES_PASSWORD=password",
			"POSTGRES_USER=postgres",
			"POSTGRES_DB=postgres",
			"listen_addresses = '*'",
		},
	}, func(config *docker.HostConfig) {
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	require.NoError(t, err)
	return "postgres://postgres:password@" + resource.GetHostPort("5432/tcp") + "/postgres?sslmode=disable", resource, pool
}
