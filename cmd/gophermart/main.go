package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/korobkovandrey/gmartloyalty/internal/config"
	"github.com/korobkovandrey/gmartloyalty/internal/infra/store"
	"github.com/korobkovandrey/gmartloyalty/internal/server"
	"github.com/korobkovandrey/gmartloyalty/pkg/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}
	l, err := logging.NewZapLogger(zapcore.Level(cfg.LogLevel), cfg.LogOutput)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Sync()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	l.InfoCtx(ctx, "", zap.Any("config", cfg))

	s, err := store.NewStore(ctx, cfg.DatabaseURI)
	if err != nil {
		l.FatalCtx(ctx, fmt.Errorf("failed to open database: %v", err).Error())
	}
	defer func() {
		if err := s.Close(); err != nil {
			l.ErrorCtx(ctx, fmt.Errorf("failed to close database: %v", err).Error())
		}
	}()

	if err := server.Launch(ctx, cfg, l, s); err != nil {
		l.FatalCtx(ctx, fmt.Errorf("failed to launch server: %v", err).Error())
	}
}
