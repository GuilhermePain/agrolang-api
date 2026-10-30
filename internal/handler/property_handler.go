package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
)

type PropertyStore interface {
	CreateForUser(ctx context.Context, userID string, p model.Property) (model.Property, error)
	GetByIDForUser(ctx context.Context, id, userID string) (model.Property, error)
	ListByUser(ctx context.Context, userID string) ([]model.Property, error)
}

type PropertyAlertLister interface {
	ListByProperty(ctx context.Context, propertyID string) ([]model.Alert, error)
}

type PropertyForecastFetcher interface {
	FetchForecast(ctx context.Context, coords weather.Coordinates) ([]risk.WeatherReading, error)
}

type PropertyHandler struct {
	store    PropertyStore
	alerts   PropertyAlertLister
	forecast PropertyForecastFetcher
}

func NewPropertyHandler(store PropertyStore, alerts PropertyAlertLister, forecast PropertyForecastFetcher) *PropertyHandler {
	return &PropertyHandler{store: store, alerts: alerts, forecast: forecast}
}

func (h *PropertyHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /properties", h.create)
	mux.HandleFunc("GET /properties", h.list)
	mux.HandleFunc("GET /properties/{id}", h.getByID)
	mux.HandleFunc("GET /properties/{id}/alerts", h.listAlerts)
	mux.HandleFunc("GET /properties/{id}/forecast", h.forecastFor)
}

func (h *PropertyHandler) create(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	var p model.Property
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.store.CreateForUser(r.Context(), user.ID, p)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "producer not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create property")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

// ownedProperty loads the property only if it belongs to the user, writing
// the 404/500 response itself when it can't.
func (h *PropertyHandler) ownedProperty(w http.ResponseWriter, r *http.Request) (model.Property, bool) {
	user, ok := requireUser(w, r)
	if !ok {
		return model.Property{}, false
	}

	p, err := h.store.GetByIDForUser(r.Context(), r.PathValue("id"), user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "property not found")
		return model.Property{}, false
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get property")
		return model.Property{}, false
	}
	return p, true
}

func (h *PropertyHandler) getByID(w http.ResponseWriter, r *http.Request) {
	p, ok := h.ownedProperty(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, p)
}

func (h *PropertyHandler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	properties, err := h.store.ListByUser(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list properties")
		return
	}

	writeJSON(w, http.StatusOK, properties)
}

func (h *PropertyHandler) listAlerts(w http.ResponseWriter, r *http.Request) {
	p, ok := h.ownedProperty(w, r)
	if !ok {
		return
	}

	alerts, err := h.alerts.ListByProperty(r.Context(), p.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list alerts")
		return
	}

	writeJSON(w, http.StatusOK, alerts)
}

func (h *PropertyHandler) forecastFor(w http.ResponseWriter, r *http.Request) {
	p, ok := h.ownedProperty(w, r)
	if !ok {
		return
	}

	readings, err := h.forecast.FetchForecast(r.Context(), weather.Coordinates{Lat: p.Latitude, Lon: p.Longitude})
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to fetch forecast")
		return
	}

	writeJSON(w, http.StatusOK, readings)
}
