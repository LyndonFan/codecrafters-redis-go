package logger

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	LOG_LEVEL_DEBUG string = "debug"
	LOG_LEVEL_INFO  string = "verbose"
	LOG_LEVEL_WARN  string = "notice"
	LOG_LEVEL_ERROR string = "warning"
)

var levelMapper = map[string]slog.Level{
	LOG_LEVEL_DEBUG: slog.LevelDebug,
	LOG_LEVEL_INFO:  slog.LevelInfo,
	LOG_LEVEL_WARN:  slog.LevelWarn,
	LOG_LEVEL_ERROR: slog.LevelError,
}

func NewLogger(port int, logLevel string) *slog.Logger {
	minLevel := levelMapper[logLevel]
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: minLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != slog.MessageKey {
				return a
			}
			msg := a.Value.Any()
			a.Value = slog.StringValue(fmt.Sprintf("[localhost:%04d] %v", port, msg))
			return a
		},
	})
	return slog.New(handler)
}
