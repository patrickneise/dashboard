package weather

import (
	"context"
	"fmt"

	"github.com/patrickneise/dashboard/internal/httpx"
)

const openMeteoBaseURL = "https://api.open-meteo.com/v1/forecast"

type Client struct {
	http *httpx.Client
}

func NewClient(h *httpx.Client) *Client {
	if h == nil {
		h = httpx.New("dashboard/0.1")
	}
	return &Client{http: h}
}

func (c *Client) FetchCurrentAndHourly(ctx context.Context, lat, lon float64, hours int) (*OpenMeteoResponse, error) {
	url := fmt.Sprintf(
		"%s?latitude=%.4f&longitude=%.4f&hourly=temperature_2m&current=temperature_2m,apparent_temperature,wind_speed_10m&timezone=auto&forecast_hours=%d&temperature_unit=fahrenheit",
		openMeteoBaseURL,
		lat,
		lon,
		hours,
	)

	var data OpenMeteoResponse
	if err := c.http.GetJSON(ctx, url, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
