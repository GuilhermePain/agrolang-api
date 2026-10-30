package handler

import (
	"context"
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

const defaultRecentAlertsLimit = 50

type RecentAlertLister interface {
	ListRecentForUser(ctx context.Context, userID string, limit int) ([]model.Alert, error)
}

type AlertHandler struct {
	store RecentAlertLister
}

func NewAlertHandler(store RecentAlertLister) *AlertHandler {
	return &AlertHandler{store: store}
}

func (h *AlertHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /alerts", h.list)
}

func (h *AlertHandler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	alerts, err := h.store.ListRecentForUser(r.Context(), user.ID, defaultRecentAlertsLimit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	writeJSON(w, http.StatusOK, alerts)
}
