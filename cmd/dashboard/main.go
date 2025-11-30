package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/patrickneise/dashboard/internal/app"
	"github.com/patrickneise/dashboard/internal/config"
	"github.com/patrickneise/dashboard/internal/logging"
)

func main() {
	// Load Config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config_error", slog.Any("err", err))
		os.Exit(1)
	}

	// Setup Logging
	mode := logging.ModeDev
	if cfg.Env == "prod" {
		mode = logging.ModeProd
	}
	log := logging.New(mode)
	slog.SetDefault(log)

	// Build App
	a, err := app.Build(cfg, log)
	if err != nil {
		log.Error("app_build_failed", slog.Any("err", err))
		os.Exit(1)
	}

	// Configure Server
	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           a.Router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	// Server Start/Stop
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info("starting_server", slog.String("addr", cfg.Addr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown_signal_received")
	case err := <-errCh:
		if err != nil {
			log.Error("server_error", slog.Any("err", err))
			os.Exit(1)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown_error", slog.Any("err", err))
		_ = srv.Close()
	} else {
		log.Info("server_stopped")
	}
}
