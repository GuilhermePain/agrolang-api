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

// fakeNotificationLister holds only the test user's logs and producers;
// any other producer behaves like another account's and yields ErrNotFound.
type fakeNotificationLister struct {
	logs           []model.NotificationLog
	ownedProducers map[string]bool
	listedFor      string
}

func (f *fakeNotificationLister) ListRecentForUser(ctx context.Context, userID string, limit int) ([]model.NotificationLog, error) {
	f.listedFor = userID
	if limit < len(f.logs) {
		return f.logs[:limit], nil
	}
	return f.logs, nil
}

func (f *fakeNotificationLister) ListByProducerForUser(ctx context.Context, producerID, userID string) ([]model.NotificationLog, error) {
	f.listedFor = userID
	if !f.ownedProducers[producerID] {
		return nil, repository.ErrNotFound
	}
	filtered := make([]model.NotificationLog, 0)
	for _, l := range f.logs {
		if l.ProducerID == producerID {
			filtered = append(filtered, l)
		}
	}
	return filtered, nil
}

func newNotificationMux(store *fakeNotificationLister) *http.ServeMux {
	mux := http.NewServeMux()
	handler.NewNotificationHandler(store).Register(mux)
	return mux
}

func TestNotificationHandler_ListRecent_ReturnsUsersLogs(t *testing.T) {
	store := &fakeNotificationLister{logs: []model.NotificationLog{
		{ID: "log-1", Phone: "5511999999999", Status: model.NotificationStatusSent},
	}}
	mux := newNotificationMux(store)

	req := authed(httptest.NewRequest(http.MethodGet, "/notifications", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.NotificationLog
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 log, got %d", len(got))
	}
	if store.listedFor != testUserID {
		t.Errorf("expected list scoped to %q, got %q", testUserID, store.listedFor)
	}
}

func TestNotificationHandler_ListByProducer_ReturnsOwnProducersLogs(t *testing.T) {
	store := &fakeNotificationLister{
		logs: []model.NotificationLog{
			{ID: "log-1", ProducerID: "prod-1"},
			{ID: "log-2", ProducerID: "prod-2"},
		},
		ownedProducers: map[string]bool{"prod-1": true, "prod-2": true},
	}
	mux := newNotificationMux(store)

	req := authed(httptest.NewRequest(http.MethodGet, "/producers/prod-1/notifications", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.NotificationLog
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(got) != 1 || got[0].ID != "log-1" {
		t.Fatalf("expected only log-1, got %+v", got)
	}
	if store.listedFor != testUserID {
		t.Errorf("expected list scoped to %q, got %q", testUserID, store.listedFor)
	}
}

func TestNotificationHandler_ListByProducer_ForeignProducer_Returns404(t *testing.T) {
	store := &fakeNotificationLister{
		logs:           []model.NotificationLog{{ID: "log-9", ProducerID: "prod-9"}},
		ownedProducers: map[string]bool{},
	}
	mux := newNotificationMux(store)

	req := authed(httptest.NewRequest(http.MethodGet, "/producers/prod-9/notifications", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for a producer the user does not own, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestNotificationHandler_WithoutUser_Returns401(t *testing.T) {
	mux := newNotificationMux(&fakeNotificationLister{ownedProducers: map[string]bool{"prod-1": true}})

	for _, path := range []string{"/notifications", "/producers/prod-1/notifications"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("GET %s: expected 401, got %d", path, rec.Code)
		}
	}
}
