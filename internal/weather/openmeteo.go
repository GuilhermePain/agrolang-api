package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
)

const defaultBaseURL = "https://api.open-meteo.com"

const hourlyParams = "temperature_2m,relative_humidity_2m,precipitation,wind_speed_10m"

type Coordinates struct {
	Lat float64
	Lon float64
}

type OpenMeteoClient struct {
	baseURL    string
	httpClient *http.Client
}

type Option func(*OpenMeteoClient)

func WithBaseURL(baseURL string) Option {
	return func(c *OpenMeteoClient) {
		c.baseURL = baseURL
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *OpenMeteoClient) {
		c.httpClient = httpClient
	}
}

func NewOpenMeteoClient(opts ...Option) *OpenMeteoClient {
	c := &OpenMeteoClient{
		baseURL:    defaultBaseURL,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

type forecastResponse struct {
	Hourly struct {
		Time               []string  `json:"time"`
		Temperature2m      []float64 `json:"temperature_2m"`
		RelativeHumidity2m []float64 `json:"relative_humidity_2m"`
		Precipitation      []float64 `json:"precipitation"`
	} `json:"hourly"`
}

func (c *OpenMeteoClient) FetchForecast(ctx context.Context, coords Coordinates) ([]risk.WeatherReading, error) {
	url := fmt.Sprintf("%s/v1/forecast?latitude=%f&longitude=%f&hourly=%s",
		c.baseURL, coords.Lat, coords.Lon, hourlyParams)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("weather: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("weather: request forecast: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather: unexpected status %d", resp.StatusCode)
	}

	var parsed forecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, fmt.Errorf("weather: decode response: %w", err)
	}

	readings := make([]risk.WeatherReading, 0, len(parsed.Hourly.Time))
	for i, ts := range parsed.Hourly.Time {
		t, err := time.Parse("2006-01-02T15:04", ts)
		if err != nil {
			return nil, fmt.Errorf("weather: parse time %q: %w", ts, err)
		}
		readings = append(readings, risk.WeatherReading{
			Time:            t,
			TempAvgC:        parsed.Hourly.Temperature2m[i],
			HumidityPct:     parsed.Hourly.RelativeHumidity2m[i],
			PrecipitationMM: parsed.Hourly.Precipitation[i],
		})
	}

	return readings, nil
}
