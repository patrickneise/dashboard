package weather

type openMeteoResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`

	Timezone       string `json:"timezone"`
	TimezoneAbbrev string `json:"timezone_abbreviation"`

	Current struct {
		Time                string  `json:"time"`
		Temperature2m       float64 `json:"temperature_2m"`
		ApparentTemperature float64 `json:"apparent_temperature"`
		WindSpeed10m        float64 `json:"wind_speed_10m"`
	} `json:"current"`

	Hourly struct {
		Time          []string  `json:"time"`
		Temperature2m []float64 `json:"temperature_2m"`
	} `json:"hourly"`
}
