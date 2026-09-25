package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type PropertyStore interface {
	Create(ctx context.Context, p model.Property) (model.Property, error)
	GetByID(ctx context.Context, id string) (model.Property, error)
	ListAll(ctx context.Context) ([]model.Property, error)
}

type PropertyAlertLister interface {
	ListByProperty(ctx context.Context, propertyID string) ([]model.Alert, error)
}

type PropertyHandler struct {
	store  PropertyStore
	alerts PropertyAlertLister
}

func NewPropertyHandler(store PropertyStore, alerts PropertyAlertLister) *PropertyHandler {
	return &PropertyHandler{store: store, alerts: alerts}
}

func (h *PropertyHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /properties", h.create)
	mux.HandleFunc("GET /properties", h.list)
	mux.HandleFunc("GET /properties/{id}", h.getByID)
	mux.HandleFunc("GET /properties/{id}/alerts", h.listAlerts)
}

func (h *PropertyHandler) create(w http.ResponseWriter, r *http.Request) {
	var p model.Property
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.store.Create(r.Context(), p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create property")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *PropertyHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	p, err := h.store.GetByID(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "property not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get property")
		return
	}

	writeJSON(w, http.StatusOK, p)
}

func (h *PropertyHandler) list(w http.ResponseWriter, r *http.Request) {
	properties, err := h.store.ListAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list properties")
		return
	}

	writeJSON(w, http.StatusOK, properties)
}

func (h *PropertyHandler) listAlerts(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	alerts, err := h.alerts.ListByProperty(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	writeJSON(w, http.StatusOK, alerts)
}
