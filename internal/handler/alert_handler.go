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
	ListRecentForUser(ctx context.Context, userID string, limit int) ([]model.Alert, error)
	ResolveForUser(ctx context.Context, id, userID string) error
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

func (h *AlertHandler) resolve(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	id := r.PathValue("id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing alert id")
		return
	}

	if err := h.store.ResolveForUser(r.Context(), id, user.ID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeError(w, http.StatusNotFound, "alert not found or already resolved")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to resolve alert")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "resolved", "id": id})
}
