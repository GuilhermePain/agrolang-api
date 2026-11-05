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
	Create(ctx context.Context, p model.Property) (model.Property, error)
	GetByID(ctx context.Context, id string) (model.Property, error)
	ListAll(ctx context.Context) ([]model.Property, error)
	ListByProducer(ctx context.Context, producerID string) ([]model.Property, error)
	Update(ctx context.Context, p model.Property) (model.Property, error)
	Delete(ctx context.Context, id string) error
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
	mux.HandleFunc("PUT /properties/{id}", h.update)
	mux.HandleFunc("DELETE /properties/{id}", h.delete)
	mux.HandleFunc("GET /properties/{id}/alerts", h.listAlerts)
	mux.HandleFunc("GET /properties/{id}/forecast", h.forecastFor)
	mux.HandleFunc("GET /producers/{id}/properties", h.listByProducer)
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
	producerID := r.URL.Query().Get("producer_id")
	var properties []model.Property
	var err error

	if producerID != "" {
		properties, err = h.store.ListByProducer(r.Context(), producerID)
	} else {
		properties, err = h.store.ListAll(r.Context())
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list properties")
		return
	}

	writeJSON(w, http.StatusOK, properties)
}

func (h *PropertyHandler) listByProducer(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	properties, err := h.store.ListByProducer(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list properties for producer")
		return
	}
	writeJSON(w, http.StatusOK, properties)
}

func (h *PropertyHandler) update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var p model.Property
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	p.ID = id

	updated, err := h.store.Update(r.Context(), p)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "property not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update property")
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *PropertyHandler) delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.store.Delete(r.Context(), id)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "property not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete property")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "id": id})
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

func (h *PropertyHandler) forecastFor(w http.ResponseWriter, r *http.Request) {
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

	readings, err := h.forecast.FetchForecast(r.Context(), weather.Coordinates{Lat: p.Latitude, Lon: p.Longitude})
	if err != nil {
		writeError(w, http.StatusBadGateway, "failed to fetch forecast")
		return
	}

	writeJSON(w, http.StatusOK, readings)
}
