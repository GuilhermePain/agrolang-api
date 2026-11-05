package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

const defaultRecentAlertsLimit = 50

type AlertStore interface {
	ListRecent(ctx context.Context, limit int) ([]model.Alert, error)
	Resolve(ctx context.Context, id string) error
}

type AlertHandler struct {
	store AlertStore
}

func NewAlertHandler(store AlertStore) *AlertHandler {
	return &AlertHandler{store: store}
}

func (h *AlertHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /alerts", h.list)
	mux.HandleFunc("POST /alerts/{id}/resolve", h.resolve)
}

func (h *AlertHandler) list(w http.ResponseWriter, r *http.Request) {
	alerts, err := h.store.ListRecent(r.Context(), defaultRecentAlertsLimit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	writeJSON(w, http.StatusOK, alerts)
}

func (h *AlertHandler) resolve(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing alert id")
		return
	}

	if err := h.store.Resolve(r.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "alert not found or already resolved")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to resolve alert")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "resolved", "id": id})
}
