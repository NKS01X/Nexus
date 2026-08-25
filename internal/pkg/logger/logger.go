package logger

import (
	"log/slog"
	"os"
)

// New creates a new structured logger.
func New(level string) *slog.Logger {
	var lvl slog.Level
	switch level {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: lvl,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {

			if a.Key == "password" || a.Key == "secret" || a.Key == "token" ||
				a.Key == "key_secret" || a.Key == "card_number" || a.Key == "cvv" {
				return slog.String(a.Key, "[REDACTED]")
			}
			return a
		},
	}

	return slog.New(slog.NewJSONHandler(os.Stdout, opts))
}
