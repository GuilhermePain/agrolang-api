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

func (f *fakePropertyStore) ListByProducer(ctx context.Context, producerID string) ([]model.Property, error) {
	var filtered []model.Property
	for _, p := range f.all {
		if p.ProducerID == producerID {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

func (f *fakePropertyStore) Update(ctx context.Context, p model.Property) (model.Property, error) {
	if _, ok := f.byID[p.ID]; !ok {
		return model.Property{}, repository.ErrNotFound
	}
	f.byID[p.ID] = p
	return p, nil
}

func (f *fakePropertyStore) Delete(ctx context.Context, id string) error {
	if _, ok := f.byID[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.byID, id)
	return nil
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

func TestPropertyHandler_Update_Success(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{
		"prop-1": {ID: "prop-1", ProducerID: "prod-1", Crop: "corn", CropStage: model.CropStageGermination},
	}}
	h := handler.NewPropertyHandler(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})
	mux := http.NewServeMux()
	h.Register(mux)

	body := `{"producer_id":"prod-1","latitude":-22.9,"longitude":-47.06,"crop":"corn","soil_type":"clay","crop_stage":"flowering"}`
	req := httptest.NewRequest(http.MethodPut, "/properties/prop-1", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.byID["prop-1"].CropStage != model.CropStageFlowering {
		t.Errorf("expected updated crop stage flowering, got %s", store.byID["prop-1"].CropStage)
	}
}

func TestPropertyHandler_Delete_Success(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{
		"prop-1": {ID: "prop-1"},
	}}
	h := handler.NewPropertyHandler(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodDelete, "/properties/prop-1", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, exists := store.byID["prop-1"]; exists {
		t.Fatalf("expected property to be deleted")
	}
}

func TestPropertyHandler_ListByProducer_Success(t *testing.T) {
	store := &fakePropertyStore{all: []model.Property{
		{ID: "prop-1", ProducerID: "prod-1", Crop: "milho"},
		{ID: "prop-2", ProducerID: "prod-2", Crop: "cafe"},
	}}
	h := handler.NewPropertyHandler(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/producers/prod-1/properties", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.Property
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].ID != "prop-1" {
		t.Fatalf("expected 1 property with ID prop-1, got %+v", got)
	}
}

func TestPropertyHandler_Forecast_Success(t *testing.T) {
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
}
