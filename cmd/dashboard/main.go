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

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/patrickneise/dashboard/internal/config"
	"github.com/patrickneise/dashboard/internal/httpx"
	"github.com/patrickneise/dashboard/internal/logging"
	"github.com/patrickneise/dashboard/internal/server"
	"github.com/patrickneise/dashboard/internal/widgets"
	"github.com/patrickneise/dashboard/internal/widgets/hn"
	"github.com/patrickneise/dashboard/internal/widgets/weather"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config_error", slog.Any("err", err))
		os.Exit(1)
	}

	mode := logging.ModeDev
	if cfg.Env == "prod" {
		mode = logging.ModeProd
	}
	log := logging.New(mode)
	slog.SetDefault(log)

	r := chi.NewRouter()
	// Core middleware (order matters)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	// Optional but usefull
	r.Use(middleware.Compress(5))
	r.Use(middleware.Timeout(30 * time.Second))
	// Security
	r.Use(server.SecurityHeaders)
	// Request Logger
	r.Use(logging.RequestLogger(log))

	sharedHTTP := httpx.New("dashboard/0.1")
	sharedHTTP.Retries = 2
	sharedHTTP.Backoff = 250 * time.Millisecond

	weatherWidget := weather.NewWidgetHander(weather.Options{
		Lat:          cfg.WeatherLat,
		Lon:          cfg.WeatherLon,
		Hours:        cfg.WeatherHours,
		LocationName: "Annapolis, MD",
		TTL:          cfg.WidgetTTL,
		Log:          log,
		Client:       weather.NewClient(sharedHTTP),
	})

	hnWidget := hn.NewWidgetHandler(hn.Options{
		Count:  10,
		TTL:    cfg.WidgetTTL,
		Log:    log,
		Client: hn.NewClient(sharedHTTP),
	})

	reg := widgets.New()
	reg.MustAdd(widgets.Spec{
		Key:     "weather",
		Title:   "Weather",
		Handler: weatherWidget,
	})
	reg.MustAdd(widgets.Spec{
		Key:     "hn",
		Title:   "Hacker News",
		Handler: hnWidget,
	})
	server.RegisterRoutes(r, reg)

	srv := &http.Server{
		Addr:              cfg.Addr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1 MiB
	}

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
