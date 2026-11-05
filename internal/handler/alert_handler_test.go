package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type fakeAlertStore struct {
	recent   []model.Alert
	resolved map[string]bool
}

func (f *fakeAlertStore) ListRecent(ctx context.Context, limit int) ([]model.Alert, error) {
	if limit < len(f.recent) {
		return f.recent[:limit], nil
	}
	return f.recent, nil
}

func (f *fakeAlertStore) Resolve(ctx context.Context, id string) error {
	if f.resolved == nil {
		f.resolved = make(map[string]bool)
	}
	for _, a := range f.recent {
		if a.ID == id {
			f.resolved[id] = true
			return nil
		}
	}
	return repository.ErrNotFound
}

func TestAlertHandler_List_ReturnsRecentAlerts(t *testing.T) {
	store := &fakeAlertStore{recent: []model.Alert{
		{ID: "a1", AlertType: "frost"},
		{ID: "a2", AlertType: "heavy_rain"},
	}}
	h := handler.NewAlertHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/alerts", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.Alert
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(got))
	}
}

func TestAlertHandler_Resolve_Success(t *testing.T) {
	store := &fakeAlertStore{recent: []model.Alert{
		{ID: "a1", AlertType: "frost"},
	}}
	h := handler.NewAlertHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/alerts/a1/resolve", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !store.resolved["a1"] {
		t.Fatalf("expected alert a1 to be marked as resolved")
	}
}

func TestAlertHandler_Resolve_NotFound(t *testing.T) {
	store := &fakeAlertStore{recent: []model.Alert{}}
	h := handler.NewAlertHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/alerts/unknown/resolve", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
