package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gin-contrib/graceful"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/korobkovandrey/gmartloyalty/internal/config"
	"github.com/korobkovandrey/gmartloyalty/internal/handlers"
	"github.com/korobkovandrey/gmartloyalty/internal/handlers/auth"
	"github.com/korobkovandrey/gmartloyalty/internal/handlers/orders"
	"github.com/korobkovandrey/gmartloyalty/internal/infra/store"
	"github.com/korobkovandrey/gmartloyalty/internal/middleware"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/pkg/logging"
	"go.uber.org/zap/zapcore"
)

func Launch(ctx context.Context, cfg *config.Config, l *logging.ZapLogger, s *store.Store, a orders.AccrualPusher) error {
	r, err := graceful.New(gin.New(), graceful.WithAddr(cfg.RunAddress))
	if err != nil {
		return fmt.Errorf("failed to create router: %v", err)
	}
	defer r.Close()
	zapLoggerMiddleware := ginzap.GinzapWithConfig(l.Logger(), &ginzap.Config{
		TimeFormat:   time.RFC3339,
		UTC:          true,
		DefaultLevel: zapcore.InfoLevel,
		Context: func(c *gin.Context) []zapcore.Field {
			return logging.GetContextFields(c)
		},
	})
	r.Use(zapLoggerMiddleware)
	r.Use(ginzap.RecoveryWithZap(l.Logger(), true))

	authService := service.NewAuth(&service.AuthConfig{
		Secret:   cfg.JWTKey,
		LifeTime: time.Duration(cfg.JWTLifeHours) * time.Hour,
	}, s)
	r.POST("/api/user/login", auth.NewLoginHandler(authService))
	r.POST("/api/user/register", auth.NewRegisterHandler(authService))

	r.Group("/").
		Use(middleware.UserAuthJWT([]byte(cfg.JWTKey), service.NewUserFinder(s))).
		POST("/api/user/orders", orders.NewCreateHandler(service.NewOrderCreator(s), a)).
		GET("/api/user/orders", orders.NewListHandler(service.NewOrderLister(s))).
		GET("/api/user/balance", handlers.NewBalanceHandler(service.NewBalance(s))).
		POST("/api/user/balance/withdraw", handlers.NewWithdrawHandler(service.NewWithdraw(s))).
		GET("/api/user/withdrawals", handlers.NewWithdrawalsHandler(service.NewWithdrawals(s)))

	if err := r.RunWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("failed to run router: %v", err)
	}
	return nil
}
