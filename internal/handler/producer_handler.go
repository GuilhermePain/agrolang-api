package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type ProducerStore interface {
	CreateForUser(ctx context.Context, userID string, p model.Producer) (model.Producer, error)
	GetByIDForUser(ctx context.Context, id, userID string) (model.Producer, error)
	ListByUser(ctx context.Context, userID string) ([]model.Producer, error)
}

type ProducerHandler struct {
	store ProducerStore
}

func NewProducerHandler(store ProducerStore) *ProducerHandler {
	return &ProducerHandler{store: store}
}

func (h *ProducerHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /producers", h.create)
	mux.HandleFunc("GET /producers", h.list)
	mux.HandleFunc("GET /producers/{id}", h.getByID)
}

func (h *ProducerHandler) create(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	var p model.Producer
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.store.CreateForUser(r.Context(), user.ID, p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create producer")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *ProducerHandler) list(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	producers, err := h.store.ListByUser(r.Context(), user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list producers")
		return
	}

	writeJSON(w, http.StatusOK, producers)
}

func (h *ProducerHandler) getByID(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}

	p, err := h.store.GetByIDForUser(r.Context(), r.PathValue("id"), user.ID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "producer not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get producer")
		return
	}

	writeJSON(w, http.StatusOK, p)
}
