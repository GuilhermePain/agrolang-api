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
	recent []model.Alert
}

func (f *fakeRecentAlertLister) ListRecent(ctx context.Context, limit int) ([]model.Alert, error) {
	if limit < len(f.recent) {
		return f.recent[:limit], nil
	}
	return f.recent, nil
}

func TestAlertHandler_List_ReturnsRecentAlerts(t *testing.T) {
	store := &fakeRecentAlertLister{recent: []model.Alert{
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
