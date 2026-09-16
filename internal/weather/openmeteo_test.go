package weather_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/weather"
)

func TestOpenMeteoClient_FetchForecast_ParsesHourlyReadings(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"latitude": -22.9,
			"longitude": -47.06,
			"hourly": {
				"time": ["2026-08-30T00:00", "2026-08-30T01:00"],
				"temperature_2m": [36.5, 37.0],
				"relative_humidity_2m": [25.0, 20.0],
				"precipitation": [0.0, 0.0],
				"wind_speed_10m": [8.5, 9.2]
			}
		}`))
	}))
	defer server.Close()

	client := weather.NewOpenMeteoClient(weather.WithBaseURL(server.URL))

	readings, err := client.FetchForecast(context.Background(), weather.Coordinates{Lat: -22.9, Lon: -47.06})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(readings) != 2 {
		t.Fatalf("expected 2 readings, got %d", len(readings))
	}

	first := readings[0]
	if first.TempAvgC != 36.5 {
		t.Errorf("expected TempAvgC 36.5, got %v", first.TempAvgC)
	}
	if first.HumidityPct != 25.0 {
		t.Errorf("expected HumidityPct 25.0, got %v", first.HumidityPct)
	}
	if first.Time.IsZero() {
		t.Errorf("expected non-zero Time")
	}
}

func TestOpenMeteoClient_FetchForecast_NonOKStatus_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := weather.NewOpenMeteoClient(weather.WithBaseURL(server.URL))

	_, err := client.FetchForecast(context.Background(), weather.Coordinates{Lat: -22.9, Lon: -47.06})
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}
