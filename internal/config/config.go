package config

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	RunAddress           string   `mapstructure:"run_address"`
	DatabaseURI          string   `mapstructure:"database_uri"`
	AccrualSystemAddress string   `mapstructure:"accrual_system_address"`
	JWTKey               string   `mapstructure:"jwt_key"`
	JWTLifeHours         int16    `mapstructure:"jwt_life_hours"`
	LogLevel             int8     `mapstructure:"log_level"`
	LogOutput            []string `mapstructure:"log_output"`
}

func NewConfig() (cfg *Config, err error) {
	conf := viper.New()
	conf.AutomaticEnv()
	conf.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	cfg = &Config{
		RunAddress:           ":8090",
		DatabaseURI:          "",
		AccrualSystemAddress: "http://localhost:8080",
		JWTKey:               "secret",
		JWTLifeHours:         24,
		LogLevel:             -1,
		LogOutput:            []string{"stderr"},
	}
	conf.SetDefault("run_address", cfg.RunAddress)
	pflag.String("a", cfg.RunAddress, "address to run HTTP server")
	_ = conf.BindPFlag("run_address", pflag.Lookup("a"))

	conf.SetDefault("database_uri", cfg.DatabaseURI)
	pflag.String("d", cfg.DatabaseURI, "URI to database")
	_ = conf.BindPFlag("database_uri", pflag.Lookup("d"))

	conf.SetDefault("accrual_system_address", cfg.AccrualSystemAddress)
	pflag.String("r", cfg.AccrualSystemAddress, "accrual system address")
	_ = conf.BindPFlag("accrual_system_address", pflag.Lookup("r"))

	conf.SetDefault("jwt_key", cfg.JWTKey)
	pflag.String("j", cfg.JWTKey, "JWT secret key")
	_ = conf.BindPFlag("jwt_key", pflag.Lookup("j"))

	conf.SetDefault("jwt_life_hours", cfg.JWTLifeHours)
	pflag.Int16("h", cfg.JWTLifeHours, "JWT life hours")
	_ = conf.BindPFlag("jwt_life_hours", pflag.Lookup("h"))

	conf.SetDefault("log_level", cfg.LogLevel)
	pflag.Int8("log-level", cfg.LogLevel, "log level: -1 debug, 0 info, 1 warn, 2 error, 3 dPanic, 4 panic, 5 fatal")
	_ = conf.BindPFlag("log_level", pflag.Lookup("log-level"))

	conf.SetDefault("log_output", cfg.LogOutput)
	pflag.StringSlice("log-output", cfg.LogOutput, "comma separated log output paths")
	_ = conf.BindPFlag("log_output", pflag.Lookup("log-output"))

	if err = conf.ReadInConfig(); err != nil && !errors.As(err, &viper.ConfigFileNotFoundError{}) {
		return cfg, fmt.Errorf("failed to read config: %w", err)
	}
	pflag.Parse()
	if err = conf.Unmarshal(cfg); err != nil {
		return cfg, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if cfg.LogLevel < int8(zap.DebugLevel) || cfg.LogLevel > int8(zap.FatalLevel) {
		return nil, fmt.Errorf("invalid log level: %d", cfg.LogLevel)
	}
	if cfg.DatabaseURI == "" {
		return nil, errors.New("database URI is empty")
	}
	if cfg.JWTKey == "" {
		return nil, errors.New("JWT key is empty")
	}
	return cfg, nil
}
