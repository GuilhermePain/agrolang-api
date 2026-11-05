package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

type NotificationStore interface {
	ListRecent(ctx context.Context, limit int) ([]model.NotificationLog, error)
	ListByProducer(ctx context.Context, producerID string) ([]model.NotificationLog, error)
	UpdateStatusByPhone(ctx context.Context, phone string, status model.NotificationStatus, eventTime time.Time) error
}

type WebhookHandler struct {
	store NotificationStore
}

func NewWebhookHandler(store NotificationStore) *WebhookHandler {
	return &WebhookHandler{store: store}
}

func (h *WebhookHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /webhooks/evolution", h.evolutionWebhook)
	mux.HandleFunc("GET /notifications", h.listRecent)
	mux.HandleFunc("GET /producers/{id}/notifications", h.listByProducer)
}

type evolutionWebhookPayload struct {
	Event string `json:"event"`
	Data  struct {
		Key struct {
			RemoteJID string `json:"remoteJid"`
		} `json:"key"`
		Status   string `json:"status"`
		Datetime string `json:"datetime"`
	} `json:"data"`
}

func (h *WebhookHandler) evolutionWebhook(w http.ResponseWriter, r *http.Request) {
	var payload evolutionWebhookPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, "invalid webhook json payload")
		return
	}

	rawPhone := payload.Data.Key.RemoteJID
	if rawPhone == "" {
		// Eventos sem JID específico
		writeJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	// Normaliza JID: 5511999999999@s.whatsapp.net -> 5511999999999
	phone := strings.Split(rawPhone, "@")[0]

	statusStr := strings.ToUpper(payload.Data.Status)
	var targetStatus model.NotificationStatus

	switch statusStr {
	case "DELIVERY_ACK", "DELIVERED":
		targetStatus = model.NotificationStatusDelivered
	case "READ":
		targetStatus = model.NotificationStatusRead
	case "ERROR", "FAILED":
		targetStatus = model.NotificationStatusFailed
	default:
		writeJSON(w, http.StatusOK, map[string]string{"status": "unhandled_status"})
		return
	}

	eventTime := time.Now()
	if payload.Data.Datetime != "" {
		if t, err := time.Parse(time.RFC3339, payload.Data.Datetime); err == nil {
			eventTime = t
		}
	}

	if err := h.store.UpdateStatusByPhone(r.Context(), phone, targetStatus, eventTime); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update notification status")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "updated",
		"phone":  phone,
	})
}

func (h *WebhookHandler) listRecent(w http.ResponseWriter, r *http.Request) {
	logs, err := h.store.ListRecent(r.Context(), 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list notification logs")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *WebhookHandler) listByProducer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	logs, err := h.store.ListByProducer(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list producer notifications")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

