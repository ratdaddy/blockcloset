package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"

	"github.com/ratdaddy/blockcloset/flatbed/internal/config"
)

func Init() {
	var handler slog.Handler
	if config.LogFormat == config.LogPretty {
		handler = tint.NewHandler(os.Stdout, &tint.Options{})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
