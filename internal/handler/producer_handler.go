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
	Create(ctx context.Context, p model.Producer) (model.Producer, error)
	GetByID(ctx context.Context, id string) (model.Producer, error)
	ListAll(ctx context.Context) ([]model.Producer, error)
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
	var p model.Producer
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	created, err := h.store.Create(r.Context(), p)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create producer")
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

func (h *ProducerHandler) list(w http.ResponseWriter, r *http.Request) {
	producers, err := h.store.ListAll(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list producers")
		return
	}

	writeJSON(w, http.StatusOK, producers)
}

func (h *ProducerHandler) getByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	p, err := h.store.GetByID(r.Context(), id)
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
