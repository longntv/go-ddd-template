package zap

import (
	"github.com/google/wire"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/longntv/go-ddd-template/internal/config"
	"github.com/longntv/go-ddd-template/internal/utils/log"
)

// WireSet holds the Wire providers for zap logger.
var WireSet = wire.NewSet(
	log.WireSet,
)

// New creates a new logger instance based on the configuration.
func New(cfg *config.Config) (*zap.Logger, error) {
	var zapConfig zap.Config

	// Configure logger based on environment
	if cfg.App.Env == "production" {
		zapConfig = zap.NewProductionConfig()
	} else {
		zapConfig = zap.NewDevelopmentConfig()
	}

	// Set log level based on configuration
	level := zapcore.InfoLevel
	switch cfg.App.LogLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	}
	zapConfig.Level = zap.NewAtomicLevelAt(level)

	// Use ISO8601 time format
	zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	return zapConfig.Build()
}
