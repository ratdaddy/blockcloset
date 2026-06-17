package logger

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"

	"github.com/ratdaddy/blockcloset/gantry/internal/config"
)

func Init() {
	var handler slog.Handler
	if config.LogFormat == config.LogPretty {
		handler = tint.NewHandler(os.Stdout, &tint.Options{Level: config.LogLevel})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: config.LogLevel})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)
}
