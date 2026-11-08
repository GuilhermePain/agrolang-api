package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

const defaultRecentNotificationsLimit = 50

type NotificationLister interface {
	ListRecentForUser(ctx context.Context, userID string, limit int) ([]model.NotificationLog, error)
	ListByProducerForUser(ctx context.Context, producerID, userID string) ([]model.NotificationLog, error)
}

// NotificationHandler exposes the WhatsApp notification logs, scoped to the
// logged-in user's producers. Mount it behind RequireSession.
type NotificationHandler struct {
	store NotificationLister
}

func NewNotificationHandler(store NotificationLister) *NotificationHandler {
	return &NotificationHandler{store: store}
}

func (h *NotificationHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /notifications", h.listRecent)
	mux.HandleFunc("GET /producers/{id}/notifications", h.listByProducer)
}

func (h *NotificationHandler) listRecent(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	logs, err := h.store.ListRecentForUser(r.Context(), user.ID, defaultRecentNotificationsLimit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list notification logs")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}

func (h *NotificationHandler) listByProducer(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	logs, err := h.store.ListByProducerForUser(r.Context(), r.PathValue("id"), user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "producer not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list producer notifications")
		return
	}
	writeJSON(w, http.StatusOK, logs)
}
