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
	byID       map[string]model.Property
	all        []model.Property
	createErr  error
	createdFor string
	listedFor  string
}

func (f *fakePropertyStore) CreateForUser(ctx context.Context, userID string, p model.Property) (model.Property, error) {
	if f.createErr != nil {
		return model.Property{}, f.createErr
	}
	p.ID = "generated-id"
	f.createdFor = userID
	return p, nil
}

func (f *fakePropertyStore) GetByIDForUser(ctx context.Context, id, userID string) (model.Property, error) {
	p, ok := f.byID[id]
	if !ok {
		return model.Property{}, repository.ErrNotFound
	}
	return p, nil
}

func (f *fakePropertyStore) ListByUser(ctx context.Context, userID string) ([]model.Property, error) {
	f.listedFor = userID
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

func newPropertyMux(store *fakePropertyStore, alerts *fakePropertyAlertLister, fetcher *fakePropertyForecastFetcher) *http.ServeMux {
	mux := http.NewServeMux()
	handler.NewPropertyHandler(store, alerts, fetcher).Register(mux)
	return mux
}

func serve(mux http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := authed(httptest.NewRequest(method, path, bytes.NewBufferString(body)))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestPropertyHandler_Create_ReturnsCreatedProperty(t *testing.T) {
	store := &fakePropertyStore{}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	body := `{"producer_id":"prod-1","latitude":-22.9,"longitude":-47.06,"crop":"corn","soil_type":"clay","crop_stage":"flowering"}`
	rec := serve(mux, http.MethodPost, "/properties", body)

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
	if store.createdFor != testUserID {
		t.Errorf("expected property created for %q, got %q", testUserID, store.createdFor)
	}
}

func TestPropertyHandler_Create_ForeignProducer_Returns404(t *testing.T) {
	store := &fakePropertyStore{createErr: repository.ErrNotFound}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodPost, "/properties", `{"producer_id":"not-mine","latitude":0,"longitude":0,"crop":"x","soil_type":"y","crop_stage":"harvest"}`)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_GetByID_ReturnsProperty(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{"prop-1": {ID: "prop-1", Crop: "corn"}}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties/prop-1", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_GetByID_NotFound_Returns404(t *testing.T) {
	mux := newPropertyMux(&fakePropertyStore{byID: map[string]model.Property{}}, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties/missing", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_List_ReturnsUsersProperties(t *testing.T) {
	store := &fakePropertyStore{all: []model.Property{{ID: "p1"}, {ID: "p2"}}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties", "")

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
	if store.listedFor != testUserID {
		t.Errorf("expected list scoped to %q, got %q", testUserID, store.listedFor)
	}
}

func TestPropertyHandler_ListAlerts_ReturnsAlertsForOwnedProperty(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{"prop-1": {ID: "prop-1"}}}
	alertLister := &fakePropertyAlertLister{byProperty: map[string][]model.Alert{
		"prop-1": {{ID: "a1", PropertyID: "prop-1", AlertType: "frost"}},
	}}
	mux := newPropertyMux(store, alertLister, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties/prop-1/alerts", "")

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

func TestPropertyHandler_ListAlerts_ForeignProperty_Returns404(t *testing.T) {
	alertLister := &fakePropertyAlertLister{byProperty: map[string][]model.Alert{
		"prop-9": {{ID: "a1", PropertyID: "prop-9"}},
	}}
	mux := newPropertyMux(&fakePropertyStore{byID: map[string]model.Property{}}, alertLister, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties/prop-9/alerts", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a property the user does not own, got %d", rec.Code)
	}
}

func TestPropertyHandler_Forecast_ReturnsReadingsForProperty(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{"prop-1": {ID: "prop-1", Latitude: -22.9, Longitude: -47.06}}}
	fetcher := &fakePropertyForecastFetcher{readings: []risk.WeatherReading{{Time: time.Now(), TempAvgC: 25, HumidityPct: 50}}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, fetcher)

	rec := serve(mux, http.MethodGet, "/properties/prop-1/forecast", "")

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
	mux := newPropertyMux(&fakePropertyStore{byID: map[string]model.Property{}}, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties/missing/forecast", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_WithoutUser_Returns401(t *testing.T) {
	mux := newPropertyMux(&fakePropertyStore{}, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/properties", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
