package logger

import (
	"log/slog"
	"os"
)

// New creates a new structured logger.
func New(level string, outputPath string) *slog.Logger {
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

	var out *os.File
	if outputPath != "" {
		var err error
		out, err = os.OpenFile(outputPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			out = os.Stdout
		}
	} else {
		out = os.Stdout
	}

	return slog.New(slog.NewJSONHandler(out, opts))
}
