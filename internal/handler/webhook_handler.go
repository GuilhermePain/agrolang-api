package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

type NotificationStatusUpdater interface {
	UpdateStatusByPhone(ctx context.Context, phone string, status model.NotificationStatus, eventTime time.Time) error
}

// WebhookHandler receives delivery/read callbacks from the Evolution
// (WhatsApp) provider. It is mounted on the PUBLIC mux because the provider
// has no user session. NOTE: there is no origin verification today (no
// shared secret or signature check), so anyone who can reach the API can post
// status updates for a phone number. The listing endpoints for notification
// logs live in NotificationHandler, behind RequireSession.
type WebhookHandler struct {
	store NotificationStatusUpdater
}

func NewWebhookHandler(store NotificationStatusUpdater) *WebhookHandler {
	return &WebhookHandler{store: store}
}

// RegisterPublic mounts the provider callback; call it on the public mux.
func (h *WebhookHandler) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("POST /webhooks/evolution", h.evolutionWebhook)
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
