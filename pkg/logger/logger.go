package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"CONVERDA/pkg/setting"

	"github.com/natefinch/lumberjack"
)

type Logger struct {
	*slog.Logger
}

func NewLogger(config setting.LoggerSetting) *Logger {
	var level slog.Level
	switch config.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	hook := &lumberjack.Logger{
		Filename:   config.LogFileName,
		MaxSize:    config.MaxSize, // megabytes
		MaxBackups: config.MaxBackups,
		MaxAge:     config.MaxAge,   // days
		Compress:   config.Compress, // disabled by default
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				// Use ISO8601-like format
				return slog.Attr{
					Key:   "time",
					Value: slog.StringValue(a.Value.Time().Format(time.RFC3339)),
				}
			}
			return a
		},
	}

	// Multi-writer for stdout and file
	multiWriter := io.MultiWriter(os.Stdout, hook)
	handler := slog.NewJSONHandler(multiWriter, opts)

	return &Logger{slog.New(handler)}
}

// Info logs with info level (compatibility helper if needed, but slog has it)
func (l *Logger) Info(msg string, args ...any) {
	l.Logger.Info(msg, args...)
}

// Error logs with error level
func (l *Logger) Error(msg string, args ...any) {
	l.Logger.Error(msg, args...)
}

// Warn logs with warn level
func (l *Logger) Warn(msg string, args ...any) {
	l.Logger.Warn(msg, args...)
}

// Debug logs with debug level
func (l *Logger) Debug(msg string, args ...any) {
	l.Logger.Debug(msg, args...)
}

// Fatal logs with error level and exits
func (l *Logger) Fatal(msg string, args ...any) {
	l.Logger.Error(msg, args...)
	os.Exit(1)
}

// With wraps slog.With
func (l *Logger) With(args ...any) *Logger {
	return &Logger{l.Logger.With(args...)}
}

// WithContext wraps for future context-based logging if needed
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// For now just return self, can be expanded to extract trace IDs etc
	return l
}
