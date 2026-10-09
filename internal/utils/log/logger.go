package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/google/wire"
)

var (
	// Logger is the singleton logger instance.
	Logger *zap.Logger
)

// SetLogger sets the global logger instance.
func SetLogger(logger *zap.Logger) {
	Logger = logger
}

// Fatal logs a fatal message and exits the program.
func Fatal(msg string, fields ...zap.Field) {
	if Logger != nil {
		Logger.Fatal(msg, fields...)
	} else {
		// Fallback to default logger if not initialized
		logger, _ := NewLogger()
		logger.Fatal(msg, fields...)
	}
}

// WireSet holds the Wire providers for logger.
var WireSet = wire.NewSet(
	NewLogger,
)

// NewLogger creates a new logger instance.
func NewLogger() (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return logger, nil
}
