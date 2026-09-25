package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
)

type fakePropertyStore struct {
	byID map[string]model.Property
	all  []model.Property
}

func (f *fakePropertyStore) Create(ctx context.Context, p model.Property) (model.Property, error) {
	p.ID = "generated-id"
	return p, nil
}

func (f *fakePropertyStore) GetByID(ctx context.Context, id string) (model.Property, error) {
	p, ok := f.byID[id]
	if !ok {
		return model.Property{}, repository.ErrNotFound
	}
	return p, nil
}

func (f *fakePropertyStore) ListAll(ctx context.Context) ([]model.Property, error) {
	return f.all, nil
}

type fakePropertyAlertLister struct {
	byProperty map[string][]model.Alert
}

func (f *fakePropertyAlertLister) ListByProperty(ctx context.Context, propertyID string) ([]model.Alert, error) {
	return f.byProperty[propertyID], nil
}

type fakePropertyForecastFetcher struct {
	readings []risk.WeatherReading
	err      error
}

func (f *fakePropertyForecastFetcher) FetchForecast(ctx context.Context, coords weather.Coordinates) ([]risk.WeatherReading, error) {
	return f.readings, f.err
}

func TestPropertyHandler_Create_ReturnsCreatedProperty(t *testing.T) {
	store := &fakePropertyStore{}
	h := handler.NewPropertyHandler(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})
	mux := http.NewServeMux()
	h.Register(mux)

	body := `{"producer_id":"prod-1","latitude":-22.9,"longitude":-47.06,"crop":"corn","soil_type":"clay","crop_stage":"flowering"}`
	req := httptest.NewRequest(http.MethodPost, "/properties", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var got model.Property
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "generated-id" {
		t.Errorf("expected ID generated-id, got %q", got.ID)
	}
}

func TestPropertyHandler_GetByID_ReturnsProperty(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{
		"prop-1": {ID: "prop-1", Crop: "corn"},
	}}
	h := handler.NewPropertyHandler(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/properties/prop-1", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_GetByID_NotFound_Returns404(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{}}
	h := handler.NewPropertyHandler(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/properties/missing", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_List_ReturnsAllProperties(t *testing.T) {
	store := &fakePropertyStore{all: []model.Property{{ID: "p1"}, {ID: "p2"}}}
	h := handler.NewPropertyHandler(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/properties", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.Property
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(got))
	}
}

func TestPropertyHandler_ListAlerts_ReturnsAlertsForProperty(t *testing.T) {
	store := &fakePropertyStore{}
	alertLister := &fakePropertyAlertLister{byProperty: map[string][]model.Alert{
		"prop-1": {{ID: "a1", PropertyID: "prop-1", AlertType: "frost"}},
	}}
	h := handler.NewPropertyHandler(store, alertLister, &fakePropertyForecastFetcher{})
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/properties/prop-1/alerts", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.Alert
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].AlertType != "frost" {
		t.Fatalf("expected 1 frost alert, got %v", got)
	}
}

func TestPropertyHandler_Forecast_ReturnsReadingsForProperty(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{
		"prop-1": {ID: "prop-1", Latitude: -22.9, Longitude: -47.06},
	}}
	fetcher := &fakePropertyForecastFetcher{readings: []risk.WeatherReading{
		{Time: time.Now(), TempAvgC: 25, HumidityPct: 50},
	}}
	h := handler.NewPropertyHandler(store, &fakePropertyAlertLister{}, fetcher)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/properties/prop-1/forecast", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []risk.WeatherReading
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].TempAvgC != 25 {
		t.Fatalf("expected 1 reading with TempAvgC 25, got %v", got)
	}
}

func TestPropertyHandler_Forecast_PropertyNotFound_Returns404(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{}}
	h := handler.NewPropertyHandler(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/properties/missing/forecast", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}
