package handler

import (
	"context"
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/model"
)

const defaultRecentAlertsLimit = 50

type RecentAlertLister interface {
	ListRecent(ctx context.Context, limit int) ([]model.Alert, error)
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
	alerts, err := h.store.ListRecent(r.Context(), defaultRecentAlertsLimit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	writeJSON(w, http.StatusOK, alerts)
}
