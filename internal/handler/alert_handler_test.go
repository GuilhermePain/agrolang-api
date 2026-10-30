package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
)

type fakeRecentAlertLister struct {
	recent    []model.Alert
	listedFor string
}

func (f *fakeRecentAlertLister) ListRecentForUser(ctx context.Context, userID string, limit int) ([]model.Alert, error) {
	f.listedFor = userID
	if limit < len(f.recent) {
		return f.recent[:limit], nil
	}
	return f.recent, nil
}

func TestAlertHandler_List_ReturnsRecentAlertsForUser(t *testing.T) {
	store := &fakeRecentAlertLister{recent: []model.Alert{
		{ID: "a1", AlertType: "frost"},
		{ID: "a2", AlertType: "heavy_rain"},
	}}
	mux := http.NewServeMux()
	handler.NewAlertHandler(store).Register(mux)

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

func TestAlertHandler_WithoutUser_Returns401(t *testing.T) {
	mux := http.NewServeMux()
	handler.NewAlertHandler(&fakeRecentAlertLister{}).Register(mux)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/alerts", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}
