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

// fakeAlertStore holds only the test user's alerts in recent; any other ID
// behaves like a missing or foreign alert.
type fakeAlertStore struct {
	recent      []model.Alert
	resolved    map[string]bool
	listedFor   string
	resolvedFor string
}

func (f *fakeAlertStore) ListRecentForUser(ctx context.Context, userID string, limit int) ([]model.Alert, error) {
	f.listedFor = userID
	if limit < len(f.recent) {
		return f.recent[:limit], nil
	}
	return f.recent, nil
}

func (f *fakeAlertStore) ResolveForUser(ctx context.Context, id, userID string) error {
	f.resolvedFor = userID
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

func newAlertMux(store *fakeAlertStore) *http.ServeMux {
	mux := http.NewServeMux()
	handler.NewAlertHandler(store).Register(mux)
	return mux
}

func TestAlertHandler_List_ReturnsRecentAlertsForUser(t *testing.T) {
	store := &fakeAlertStore{recent: []model.Alert{
		{ID: "a1", AlertType: "frost"},
		{ID: "a2", AlertType: "heavy_rain"},
	}}
	mux := newAlertMux(store)

	req := authed(httptest.NewRequest(http.MethodGet, "/alerts", nil))
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
	if store.listedFor != testUserID {
		t.Errorf("expected list scoped to %q, got %q", testUserID, store.listedFor)
	}
}

func TestAlertHandler_Resolve_Success(t *testing.T) {
	store := &fakeAlertStore{recent: []model.Alert{{ID: "a1", AlertType: "frost"}}}
	mux := newAlertMux(store)

	req := authed(httptest.NewRequest(http.MethodPost, "/alerts/a1/resolve", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !store.resolved["a1"] {
		t.Fatalf("expected alert a1 to be marked as resolved")
	}
	if store.resolvedFor != testUserID {
		t.Errorf("expected resolve scoped to %q, got %q", testUserID, store.resolvedFor)
	}
}

func TestAlertHandler_Resolve_NotFound(t *testing.T) {
	mux := newAlertMux(&fakeAlertStore{recent: []model.Alert{}})

	req := authed(httptest.NewRequest(http.MethodPost, "/alerts/unknown/resolve", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestAlertHandler_Resolve_ForeignAlert_Returns404(t *testing.T) {
	store := &fakeAlertStore{recent: []model.Alert{{ID: "a1"}}}
	mux := newAlertMux(store)

	req := authed(httptest.NewRequest(http.MethodPost, "/alerts/someone-elses/resolve", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for another user's alert, got %d", rec.Code)
	}
	if store.resolved["a1"] {
		t.Fatalf("own alert must be untouched")
	}
}

func TestAlertHandler_WithoutUser_Returns401(t *testing.T) {
	store := &fakeAlertStore{recent: []model.Alert{{ID: "a1"}}}
	mux := newAlertMux(store)

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/alerts"},
		{http.MethodPost, "/alerts/a1/resolve"},
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tc.method, tc.path, rec.Code)
		}
	}
	if store.resolved["a1"] {
		t.Fatalf("unauthenticated request must not resolve anything")
	}
}
