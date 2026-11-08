package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
)

type fakeNotificationStatusUpdater struct {
	updatedStatus map[string]model.NotificationStatus
	updatedTime   map[string]time.Time
}

func (f *fakeNotificationStatusUpdater) UpdateStatusByPhone(ctx context.Context, phone string, status model.NotificationStatus, eventTime time.Time) error {
	if f.updatedStatus == nil {
		f.updatedStatus = make(map[string]model.NotificationStatus)
		f.updatedTime = make(map[string]time.Time)
	}
	f.updatedStatus[phone] = status
	f.updatedTime[phone] = eventTime
	return nil
}

func newWebhookMux(store *fakeNotificationStatusUpdater) *http.ServeMux {
	mux := http.NewServeMux()
	handler.NewWebhookHandler(store).RegisterPublic(mux)
	return mux
}

// The provider callback has no user session, so these requests are
// deliberately NOT wrapped with authed().
func TestWebhookHandler_EvolutionWebhook_Delivered(t *testing.T) {
	store := &fakeNotificationStatusUpdater{}
	mux := newWebhookMux(store)

	body := `{
		"event": "messages.update",
		"data": {
			"key": { "remoteJid": "5511999999999@s.whatsapp.net" },
			"status": "DELIVERY_ACK"
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.updatedStatus["5511999999999"] != model.NotificationStatusDelivered {
		t.Errorf("expected status delivered, got %v", store.updatedStatus["5511999999999"])
	}
}

func TestWebhookHandler_EvolutionWebhook_Read(t *testing.T) {
	store := &fakeNotificationStatusUpdater{}
	mux := newWebhookMux(store)

	body := `{
		"event": "messages.update",
		"data": {
			"key": { "remoteJid": "5511888888888@s.whatsapp.net" },
			"status": "READ"
		}
	}`
	req := httptest.NewRequest(http.MethodPost, "/webhooks/evolution", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.updatedStatus["5511888888888"] != model.NotificationStatusRead {
		t.Errorf("expected status read, got %v", store.updatedStatus["5511888888888"])
	}
}

func TestWebhookHandler_DoesNotExposeNotificationListings(t *testing.T) {
	mux := newWebhookMux(&fakeNotificationStatusUpdater{})

	for _, path := range []string{"/notifications", "/producers/prod-1/notifications"} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s on the public webhook mux: expected 404, got %d", path, rec.Code)
		}
	}
}
