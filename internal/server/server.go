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
	"github.com/korobkovandrey/gmartloyalty/internal/handlers/auth"
	"github.com/korobkovandrey/gmartloyalty/internal/handlers/orders"
	"github.com/korobkovandrey/gmartloyalty/internal/infra/store"
	"github.com/korobkovandrey/gmartloyalty/internal/middleware"
	"github.com/korobkovandrey/gmartloyalty/internal/service"
	"github.com/korobkovandrey/gmartloyalty/pkg/logging"
	"go.uber.org/zap/zapcore"
)

func Launch(ctx context.Context, cfg *config.Config, l *logging.ZapLogger, s *store.Store) error {
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
	//nolint:gocritic // ignore
	// r.Use(ginzap.RecoveryWithZap(l.Logger(), true))

	authRoutes(r.Group("/api/user"), &service.AuthConfig{
		Secret:   cfg.JWTKey,
		LifeTime: time.Duration(cfg.JWTLifeHours) * time.Hour,
	}, s)

	rAuth := r.Group("/")
	rAuth.Use(middleware.UserAuthJWT([]byte(cfg.JWTKey), service.NewUserFinder(s)))

	rAuth.POST("/api/user/orders", orders.NewCreateHandler(service.NewOrderCreator(s)))
	rAuth.GET("/api/user/orders", orders.NewListHandler(service.NewOrderLister(s)))

	if err := r.RunWithContext(ctx); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("failed to run router: %v", err)
	}
	return nil
}

func authRoutes(r gin.IRouter, cfg *service.AuthConfig, store service.UserAuthStore) {
	s := service.NewAuth(cfg, store)
	r.POST("/login", auth.NewLoginHandler(s))
	r.POST("/register", auth.NewRegisterHandler(s))
}
