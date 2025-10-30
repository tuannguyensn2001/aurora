package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger interface defines the logging operations
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	DebugContext(ctx interface{}, msg string, args ...any)
	InfoContext(ctx interface{}, msg string, args ...any)
	WarnContext(ctx interface{}, msg string, args ...any)
	ErrorContext(ctx interface{}, msg string, args ...any)
}

// DefaultLogger implements the Logger interface using slog
type DefaultLogger struct {
	logger *slog.Logger
}

// NewDefaultLogger creates a new default logger
func NewDefaultLogger(level slog.Level) Logger {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	return &DefaultLogger{
		logger: slog.New(slog.NewTextHandler(os.Stdout, opts)),
	}
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info logs an info message
func (l *DefaultLogger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// DebugContext logs a debug message with context
func (l *DefaultLogger) DebugContext(ctx interface{}, msg string, args ...any) {
	if context, ok := ctx.(context.Context); ok {
		l.logger.DebugContext(context, msg, args...)
	} else {
		l.logger.Debug(msg, args...)
	}
}

// InfoContext logs an info message with context
func (l *DefaultLogger) InfoContext(ctx interface{}, msg string, args ...any) {
	if context, ok := ctx.(context.Context); ok {
		l.logger.InfoContext(context, msg, args...)
	} else {
		l.logger.Info(msg, args...)
	}
}

// WarnContext logs a warning message with context
func (l *DefaultLogger) WarnContext(ctx interface{}, msg string, args ...any) {
	if context, ok := ctx.(context.Context); ok {
		l.logger.WarnContext(context, msg, args...)
	} else {
		l.logger.Warn(msg, args...)
	}
}

// ErrorContext logs an error message with context
func (l *DefaultLogger) ErrorContext(ctx interface{}, msg string, args ...any) {
	if context, ok := ctx.(context.Context); ok {
		l.logger.ErrorContext(context, msg, args...)
	} else {
		l.logger.Error(msg, args...)
	}
}
