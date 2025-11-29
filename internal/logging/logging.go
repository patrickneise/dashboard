package logging

import (
	"log/slog"
	"os"
)

type Mode string

const (
	ModeDev  Mode = "dev"
	ModeProd Mode = "prod"
)

func New(mode Mode) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		// AddSource is handy in dev, noisy in prod.
		AddSource: mode == ModeDev,
	}

	var h slog.Handler
	switch mode {
	case ModeDev:
		h = slog.NewTextHandler(os.Stdout, opts)
	default:
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(h)
}
