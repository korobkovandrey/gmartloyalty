package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConfig(t *testing.T) {
	err := os.Setenv("RUN_ADDRESS", "test_RunAddress")
	require.NoError(t, err)
	err = os.Setenv("DATABASE_URI", "test_DatabaseURI")
	require.NoError(t, err)
	err = os.Setenv("ACCRUAL_SYSTEM_ADDRESS", "test_AccrualSystemAddress")
	require.NoError(t, err)
	err = os.Setenv("JWT_KEY", "test_JWTKey")
	require.NoError(t, err)
	err = os.Setenv("JWT_LIFE_HOURS", "18")
	require.NoError(t, err)
	err = os.Setenv("LOG_LEVEL", "2")
	require.NoError(t, err)
	err = os.Setenv("LOG_OUTPUT", "stderr,logfile.log")
	require.NoError(t, err)
	cfg, err := NewConfig()
	require.NoError(t, err)
	assert.Equal(t, &Config{
		RunAddress:           "test_RunAddress",
		DatabaseURI:          "test_DatabaseURI",
		AccrualSystemAddress: "test_AccrualSystemAddress",
		JWTKey:               "test_JWTKey",
		JWTLifeHours:         18,
		LogLevel:             2,
		LogOutput:            []string{"stderr", "logfile.log"},
	}, cfg)
}
