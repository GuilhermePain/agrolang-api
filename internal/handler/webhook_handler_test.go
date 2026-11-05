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
)

type fakeNotificationStore struct {
	logs           []model.NotificationLog
	updatedStatus  map[string]model.NotificationStatus
	updatedTime    map[string]time.Time
}

func (f *fakeNotificationStore) ListRecent(ctx context.Context, limit int) ([]model.NotificationLog, error) {
	if limit < len(f.logs) {
		return f.logs[:limit], nil
	}
	return f.logs, nil
}

func (f *fakeNotificationStore) ListByProducer(ctx context.Context, producerID string) ([]model.NotificationLog, error) {
	var filtered []model.NotificationLog
	for _, l := range f.logs {
		if l.ProducerID == producerID {
			filtered = append(filtered, l)
		}
	}
	return filtered, nil
}

func (f *fakeNotificationStore) UpdateStatusByPhone(ctx context.Context, phone string, status model.NotificationStatus, eventTime time.Time) error {
	if f.updatedStatus == nil {
		f.updatedStatus = make(map[string]model.NotificationStatus)
		f.updatedTime = make(map[string]time.Time)
	}
	f.updatedStatus[phone] = status
	f.updatedTime[phone] = eventTime
	return nil
}

func TestWebhookHandler_EvolutionWebhook_Delivered(t *testing.T) {
	store := &fakeNotificationStore{}
	h := handler.NewWebhookHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

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
	store := &fakeNotificationStore{}
	h := handler.NewWebhookHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

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

func TestWebhookHandler_ListRecent_ReturnsLogs(t *testing.T) {
	store := &fakeNotificationStore{logs: []model.NotificationLog{
		{ID: "log-1", Phone: "5511999999999", Status: model.NotificationStatusSent},
	}}
	h := handler.NewWebhookHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
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
}

