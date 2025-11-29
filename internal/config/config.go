package config

import (
	"errors"
	"os"
	"strconv"
	"time"
)

type Env string

const (
	EnvDev  Env = "dev"
	EnvProd Env = "prod"
)

type Config struct {
	Env  Env
	Addr string

	// Weather defaults (v0)
	WeatherLat   float64
	WeatherLon   float64
	WeatherHours int

	// Widget caching defaults (v0)
	WidgetTTL time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Env:          EnvDev,
		Addr:         ":8080",
		WeatherLat:   38.947654,
		WeatherLon:   -76.476169,
		WeatherHours: 6,
		WidgetTTL:    5 * time.Minute,
	}

	if v := os.Getenv("APP_ENV"); v != "" {
		cfg.Env = Env(v)
	}

	if v := os.Getenv("ADDR"); v != "" {
		cfg.Addr = v
	}

	if v := os.Getenv("DASHBOARD_LAT"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return Config{}, errors.New("invalid DASHBOARD_LAT")
		}
		cfg.WeatherLat = f
	}

	if v := os.Getenv("DASHBOARD_LON"); v != "" {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return Config{}, errors.New("invalid DASHBOARD_LON")
		}
		cfg.WeatherLon = f
	}

	if v := os.Getenv("WEATHER_HOURS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 || n > 72 {
			return Config{}, errors.New("invalid WEATHER_HOURS")
		}
		cfg.WeatherHours = n
	}

	if v := os.Getenv("WIDGET_TTL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, errors.New("invalid WIDGET_TTL")
		}
		cfg.WidgetTTL = d
	}

	return cfg, nil
}
