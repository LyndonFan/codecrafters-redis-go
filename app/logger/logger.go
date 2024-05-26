package logger

import (
	"fmt"
	"log/slog"
	"os"
)

func NewLogger(port int) *slog.Logger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
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
