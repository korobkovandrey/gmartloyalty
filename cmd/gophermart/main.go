package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/korobkovandrey/gmartloyalty/internal/config"
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
}
