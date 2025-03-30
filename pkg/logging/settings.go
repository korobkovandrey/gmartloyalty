package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type settings struct {
	config *zap.Config
	opts   []zap.Option
}

func defaultSettings(level zap.AtomicLevel, outputPaths []string) *settings {
	config := &zap.Config{
		Level:       level,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "msg",
			LevelKey:       "level",
			TimeKey:        "ts",
			NameKey:        "logger",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.EpochTimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      outputPaths,
		ErrorOutputPaths: outputPaths,
	}

	return &settings{
		config: config,
		opts: []zap.Option{
			zap.AddCallerSkip(1),
		},
	}
}
