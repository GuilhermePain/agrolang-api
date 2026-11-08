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

// fakePropertyStore holds only the test user's data: byID/all are the user's
// properties and ownedProducers the user's producers. Anything else behaves
// like another account's data and yields ErrNotFound, as the repository does.
type fakePropertyStore struct {
	byID           map[string]model.Property
	all            []model.Property
	ownedProducers map[string]bool
	createErr      error
	createdFor     string
	listedFor      string
	updatedFor     string
	deletedFor     string
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

func (f *fakePropertyStore) ListByProducerForUser(ctx context.Context, producerID, userID string) ([]model.Property, error) {
	f.listedFor = userID
	if !f.ownedProducers[producerID] {
		return nil, repository.ErrNotFound
	}
	filtered := make([]model.Property, 0)
	for _, p := range f.all {
		if p.ProducerID == producerID {
			filtered = append(filtered, p)
		}
	}
	return filtered, nil
}

func (f *fakePropertyStore) UpdateForUser(ctx context.Context, userID string, p model.Property) (model.Property, error) {
	f.updatedFor = userID
	current, ok := f.byID[p.ID]
	if !ok {
		return model.Property{}, repository.ErrNotFound
	}
	if p.ProducerID == "" {
		p.ProducerID = current.ProducerID
	} else if !f.ownedProducers[p.ProducerID] {
		return model.Property{}, repository.ErrNotFound
	}
	f.byID[p.ID] = p
	return p, nil
}

func (f *fakePropertyStore) DeleteForUser(ctx context.Context, id, userID string) error {
	f.deletedFor = userID
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

func TestPropertyHandler_List_FilterByProducerID_StaysWithinUsersProperties(t *testing.T) {
	store := &fakePropertyStore{all: []model.Property{
		{ID: "p1", ProducerID: "prod-1"},
		{ID: "p2", ProducerID: "prod-2"},
	}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/properties?producer_id=prod-1", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.Property
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 1 || got[0].ID != "p1" {
		t.Fatalf("expected only p1, got %+v", got)
	}
	if store.listedFor != testUserID {
		t.Errorf("expected list scoped to %q, got %q", testUserID, store.listedFor)
	}
}

func TestPropertyHandler_Update_Success(t *testing.T) {
	store := &fakePropertyStore{
		byID: map[string]model.Property{
			"prop-1": {ID: "prop-1", ProducerID: "prod-1", Crop: "corn", CropStage: model.CropStageGermination},
		},
		ownedProducers: map[string]bool{"prod-1": true},
	}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	body := `{"producer_id":"prod-1","latitude":-22.9,"longitude":-47.06,"crop":"corn","soil_type":"clay","crop_stage":"flowering"}`
	rec := serve(mux, http.MethodPut, "/properties/prop-1", body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.byID["prop-1"].CropStage != model.CropStageFlowering {
		t.Errorf("expected updated crop stage flowering, got %s", store.byID["prop-1"].CropStage)
	}
	if store.updatedFor != testUserID {
		t.Errorf("expected update scoped to %q, got %q", testUserID, store.updatedFor)
	}
}

func TestPropertyHandler_Update_ForeignProperty_Returns404(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{}, ownedProducers: map[string]bool{"prod-1": true}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	body := `{"producer_id":"prod-1","latitude":0,"longitude":0,"crop":"x","soil_type":"y","crop_stage":"harvest"}`
	rec := serve(mux, http.MethodPut, "/properties/someone-elses", body)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_Update_MoveToForeignProducer_Returns404(t *testing.T) {
	store := &fakePropertyStore{
		byID:           map[string]model.Property{"prop-1": {ID: "prop-1", ProducerID: "prod-1", Crop: "corn"}},
		ownedProducers: map[string]bool{"prod-1": true},
	}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	body := `{"producer_id":"someone-elses-producer","latitude":0,"longitude":0,"crop":"x","soil_type":"y","crop_stage":"harvest"}`
	rec := serve(mux, http.MethodPut, "/properties/prop-1", body)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.byID["prop-1"].ProducerID != "prod-1" || store.byID["prop-1"].Crop != "corn" {
		t.Fatalf("property must be untouched, got %+v", store.byID["prop-1"])
	}
}

func TestPropertyHandler_Update_InvalidJSON_Returns400(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{"prop-1": {ID: "prop-1"}}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodPut, "/properties/prop-1", "not json")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_Delete_Success(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{"prop-1": {ID: "prop-1"}}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodDelete, "/properties/prop-1", "")

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, exists := store.byID["prop-1"]; exists {
		t.Fatalf("expected property to be deleted")
	}
	if store.deletedFor != testUserID {
		t.Errorf("expected delete scoped to %q, got %q", testUserID, store.deletedFor)
	}
}

func TestPropertyHandler_Delete_ForeignProperty_Returns404(t *testing.T) {
	store := &fakePropertyStore{byID: map[string]model.Property{"prop-1": {ID: "prop-1"}}}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodDelete, "/properties/someone-elses", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, exists := store.byID["prop-1"]; !exists {
		t.Fatalf("own property must be untouched")
	}
}

func TestPropertyHandler_ListByProducer_Success(t *testing.T) {
	store := &fakePropertyStore{
		all: []model.Property{
			{ID: "prop-1", ProducerID: "prod-1", Crop: "milho"},
			{ID: "prop-2", ProducerID: "prod-2", Crop: "cafe"},
		},
		ownedProducers: map[string]bool{"prod-1": true, "prod-2": true},
	}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/producers/prod-1/properties", "")

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
	if store.listedFor != testUserID {
		t.Errorf("expected list scoped to %q, got %q", testUserID, store.listedFor)
	}
}

func TestPropertyHandler_ListByProducer_ForeignProducer_Returns404(t *testing.T) {
	store := &fakePropertyStore{
		all:            []model.Property{{ID: "prop-9", ProducerID: "prod-9"}},
		ownedProducers: map[string]bool{},
	}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	rec := serve(mux, http.MethodGet, "/producers/prod-9/properties", "")

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a producer the user does not own, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestPropertyHandler_WithoutUser_Returns401(t *testing.T) {
	store := &fakePropertyStore{
		byID:           map[string]model.Property{"prop-1": {ID: "prop-1", ProducerID: "prod-1"}},
		ownedProducers: map[string]bool{"prod-1": true},
	}
	mux := newPropertyMux(store, &fakePropertyAlertLister{}, &fakePropertyForecastFetcher{})

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/properties"},
		{http.MethodPost, "/properties"},
		{http.MethodGet, "/properties/prop-1"},
		{http.MethodPut, "/properties/prop-1"},
		{http.MethodDelete, "/properties/prop-1"},
		{http.MethodGet, "/properties/prop-1/alerts"},
		{http.MethodGet, "/properties/prop-1/forecast"},
		{http.MethodGet, "/producers/prod-1/properties"},
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString("{}")))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tc.method, tc.path, rec.Code)
		}
	}
	if _, exists := store.byID["prop-1"]; !exists {
		t.Fatalf("unauthenticated requests must not delete anything")
	}
}
