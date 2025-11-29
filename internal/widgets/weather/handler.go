package weather

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/patrickneise/dashboard/internal/cache"
	"github.com/patrickneise/dashboard/internal/templates"
	"github.com/patrickneise/dashboard/internal/widget"
)

type Options struct {
	Lat          float64
	Lon          float64
	Hours        int
	LocationName string
	TTL          time.Duration

	Client *Client
	Log    *slog.Logger
}

func NewWidgetHander(opts Options) http.Handler {
	c := opts.Client

	ttl := opts.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	location := opts.LocationName
	if location == "" {
		location = "Weather"
	}

	return widget.Handler[WidgetViewModel]{
		Name:  "weather",
		TTL:   ttl,
		Cache: &cache.TTL[WidgetViewModel]{},
		Log:   opts.Log,

		Fetch: func(ctx context.Context) (WidgetViewModel, error) {
			resp, err := c.FetchCurrentAndHourly(ctx, opts.Lat, opts.Lon, opts.Hours)
			if err != nil {
				var zero WidgetViewModel
				return zero, err
			}
			vm := toViewModel(resp)
			vm.LocationName = location
			return vm, nil
		},

		Render: func(vm WidgetViewModel) templ.Component {
			return WeatherWidgetView(vm)
		},

		Error: func(_ error) templ.Component {
			return templates.WidgetError("Weather", "/widgets/weather")
		},

		MarkStale: func(vm WidgetViewModel, staleBy time.Duration) WidgetViewModel {
			vm.IsStale = true
			vm.StaleBy = staleBy.Round(time.Second).String()
			return vm
		},
	}
}

func toViewModel(api *openMeteoResponse) WidgetViewModel {
	updatedAt := api.Current.Time
	// Try to parse the ISO8601 time; if it fails, just use the string
	if t, err := time.Parse(time.RFC3339, api.Current.Time); err == nil {
		updatedAt = t.Format("15:04") // local time HH:MM
	}

	// Build next-hours forecast (first few entries)
	var next []HourForecast
	max := len(api.Hourly.Time)
	if max > 6 {
		max = 6
	}
	for i := 0; i < max; i++ {
		t, err := time.Parse(time.RFC3339, api.Hourly.Time[i])
		label := api.Hourly.Time[i]
		if err == nil {
			label = t.Format("15:04")
		}
		next = append(next, HourForecast{
			Label: label,
			Temp:  api.Hourly.Temperature2m[i],
			Time:  t,
		})
	}

	return WidgetViewModel{
		LocationName: "New York, NY", // later: make dynamic
		UpdatedAt:    updatedAt,
		CurrentTemp:  api.Current.Temperature2m,
		FeelsLike:    api.Current.ApparentTemperature,
		WindSpeedKph: api.Current.WindSpeed10m, // already in km/h
		NextHours:    next,
	}
}
