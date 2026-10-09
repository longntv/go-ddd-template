package log

import (
	"context"

	"go.uber.org/zap"
)

// Logger is an AWS SDK logger.
type Logger struct {
	logger *zap.Logger
}

// NewLogger creates a new AWS SDK logger.
func NewLogger(logger *zap.Logger) *Logger {
	return &Logger{
		logger: logger,
	}
}

// Log is called for each SDK request/response.
func (l *Logger) Log(ctx context.Context, msg string) {
	l.logger.Info(msg)
}
