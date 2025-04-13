package logging

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type key string

const zapFieldsKey key = "zapFields"

type ZapFields map[string]zap.Field

func (zf ZapFields) Append(fields ...zap.Field) ZapFields {
	zfCopy := make(ZapFields)
	for k, v := range zf {
		zfCopy[k] = v
	}
	for _, f := range fields {
		zfCopy[f.Key] = f
	}
	return zfCopy
}

type ZapLogger struct {
	logger *zap.Logger
	level  zap.AtomicLevel
}

// NewZapLogger returns a new ZapLogger configured with the provided options.
func NewZapLogger(level zapcore.Level, outputPaths []string) (*ZapLogger, error) {
	atomic := zap.NewAtomicLevelAt(level)
	s := defaultSettings(atomic, outputPaths)
	l, err := s.config.Build(s.opts...)
	if err != nil {
		return nil, fmt.Errorf("NewZapLogger: %w", err)
	}
	return &ZapLogger{
		logger: l,
		level:  atomic,
	}, nil
}

func (z *ZapLogger) Sync() {
	_ = z.logger.Sync()
}

func SetGinContextFields(ctx *gin.Context, fields ...zap.Field) {
	ctxFields, _ := ctx.Value(zapFieldsKey).(ZapFields)
	if ctxFields == nil {
		ctxFields = make(ZapFields)
	}
	ctx.Set(string(zapFieldsKey), ctxFields.Append(fields...))
}

func GetContextFields(ctx context.Context, fields ...zap.Field) []zap.Field {
	fs := make(ZapFields)
	ctxFields, ok := ctx.Value(string(zapFieldsKey)).(ZapFields)
	if ok {
		fs = ctxFields
	}
	fs = fs.Append(fields...)
	res := make([]zap.Field, len(fs))
	i := 0
	for _, f := range fs {
		res[i] = f
		i++
	}
	return res
}

func (z *ZapLogger) InfoCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Info(msg, GetContextFields(ctx, fields...)...)
}

func (z *ZapLogger) DebugCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Debug(msg, GetContextFields(ctx, fields...)...)
}

func (z *ZapLogger) WarnCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Warn(msg, GetContextFields(ctx, fields...)...)
}

func (z *ZapLogger) ErrorCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Error(msg, GetContextFields(ctx, fields...)...)
}

func (z *ZapLogger) FatalCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Fatal(msg, GetContextFields(ctx, fields...)...)
}

func (z *ZapLogger) PanicCtx(ctx context.Context, msg string, fields ...zap.Field) {
	z.logger.Panic(msg, GetContextFields(ctx, fields...)...)
}

func (z *ZapLogger) SetLevel(level zapcore.Level) {
	z.level.SetLevel(level)
}

func (z *ZapLogger) Std() *log.Logger {
	return zap.NewStdLog(z.logger)
}

func (z *ZapLogger) Logger() *zap.Logger {
	return z.logger
}
