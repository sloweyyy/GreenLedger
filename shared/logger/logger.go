package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger wraps slog.Logger with additional functionality
type Logger struct {
	*slog.Logger
}

// New creates a new logger instance
func New(level string) *Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)

	return &Logger{Logger: logger}
}

// WithContext adds context to the logger
func (l *Logger) WithContext(ctx context.Context) *Logger {
	return &Logger{Logger: l.Logger.With("trace_id", getTraceID(ctx))}
}

// WithService adds service name to the logger
func (l *Logger) WithService(serviceName string) *Logger {
	return &Logger{Logger: l.Logger.With("service", serviceName)}
}

// WithRequestID adds request ID to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	return &Logger{Logger: l.Logger.With("request_id", requestID)}
}

// WithUserID adds user ID to the logger
func (l *Logger) WithUserID(userID string) *Logger {
	return &Logger{Logger: l.Logger.With("user_id", userID)}
}

// LogError logs an error with additional context
func (l *Logger) LogError(ctx context.Context, msg string, err error, attrs ...slog.Attr) {
	args := []any{"error", err.Error()}
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value)
	}
	l.Logger.ErrorContext(ctx, msg, args...)
}

// LogInfo logs an info message with context
func (l *Logger) LogInfo(ctx context.Context, msg string, attrs ...slog.Attr) {
	args := []any{}
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value)
	}
	l.Logger.InfoContext(ctx, msg, args...)
}

// LogDebug logs a debug message with context
func (l *Logger) LogDebug(ctx context.Context, msg string, attrs ...slog.Attr) {
	args := []any{}
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value)
	}
	l.Logger.DebugContext(ctx, msg, args...)
}

// LogWarn logs a warning message with context
func (l *Logger) LogWarn(ctx context.Context, msg string, attrs ...slog.Attr) {
	args := []any{}
	for _, attr := range attrs {
		args = append(args, attr.Key, attr.Value)
	}
	l.Logger.WarnContext(ctx, msg, args...)
}

// Helper function to extract trace ID from context
func getTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// Structured logging helpers
func String(key, value string) slog.Attr {
	return slog.String(key, value)
}

func Int(key string, value int) slog.Attr {
	return slog.Int(key, value)
}

func Float64(key string, value float64) slog.Attr {
	return slog.Float64(key, value)
}

func Bool(key string, value bool) slog.Attr {
	return slog.Bool(key, value)
}

func Any(key string, value any) slog.Attr {
	return slog.Any(key, value)
}
