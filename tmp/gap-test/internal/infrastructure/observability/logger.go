// @kthulu:infrastructure:logger
package observability

import (
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}
