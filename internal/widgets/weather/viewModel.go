package weather

import "time"

type WidgetViewModel struct {
	LocationName string
	UpdatedAt    string
	CurrentTemp  float64
	FeelsLike    float64
	WindSpeedKph float64
	NextHours    []HourForecast

	IsStale bool
	StaleBy string
}

type HourForecast struct {
	Label string // e.g. "14:00"
	Temp  float64
	Time  time.Time
}
