package app

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/patrickneise/dashboard/internal/config"
	"github.com/patrickneise/dashboard/internal/httpx"
	"github.com/patrickneise/dashboard/internal/logging"
	"github.com/patrickneise/dashboard/internal/server"
	"github.com/patrickneise/dashboard/internal/widgetkit"
	"github.com/patrickneise/dashboard/internal/widgets/hn"
	"github.com/patrickneise/dashboard/internal/widgets/weather"
)

type App struct {
	Router http.Handler
}

func Build(cfg config.Config, log *slog.Logger) (*App, error) {
	// Shared HTTP client for all public API widgets
	sharedHTTP := httpx.New("dashboard/0.1 (+https://github.com/patrickneise/dashboard)")

	// Widget handlers
	weatherWidget := weather.NewWidgetHandler(weather.Options{
		Lat:    cfg.WeatherLat,
		Lon:    cfg.WeatherLon,
		Hours:  cfg.WeatherHours,
		TTL:    cfg.WidgetTTL,
		Log:    log,
		Client: weather.NewClient(sharedHTTP),
	})

	hnWidget := hn.NewWidgetHandler(hn.Options{
		Count:  10,
		TTL:    cfg.WidgetTTL,
		Log:    log,
		Client: hn.NewClient(sharedHTTP),
	})

	// Registry
	reg := widgetkit.NewRegistry()
	reg.MustAdd(widgetkit.Spec{Key: "weather", Title: "Weather", Handler: weatherWidget})
	reg.MustAdd(widgetkit.Spec{Key: "hn", Title: "Hacker News", Handler: hnWidget})

	// Router + middleware
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	r.Use(server.SecurityHeaders)
	r.Use(logging.RequestLogger(log))

	// Routes
	server.RegisterRoutes(r, reg)

	return &App{Router: r}, nil
}
