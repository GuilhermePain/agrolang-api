package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type fakeProducerStore struct {
	created   model.Producer
	byID      map[string]model.Producer
	all       []model.Producer
	createErr error
	listErr   error
}

func (f *fakeProducerStore) Create(ctx context.Context, p model.Producer) (model.Producer, error) {
	if f.createErr != nil {
		return model.Producer{}, f.createErr
	}
	p.ID = "generated-id"
	f.created = p
	return p, nil
}

func (f *fakeProducerStore) GetByID(ctx context.Context, id string) (model.Producer, error) {
	p, ok := f.byID[id]
	if !ok {
		return model.Producer{}, repository.ErrNotFound
	}
	return p, nil
}

func (f *fakeProducerStore) ListAll(ctx context.Context) ([]model.Producer, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.all, nil
}

func (f *fakeProducerStore) Update(ctx context.Context, p model.Producer) (model.Producer, error) {
	if _, ok := f.byID[p.ID]; !ok {
		return model.Producer{}, repository.ErrNotFound
	}
	f.byID[p.ID] = p
	return p, nil
}

func (f *fakeProducerStore) Delete(ctx context.Context, id string) error {
	if _, ok := f.byID[id]; !ok {
		return repository.ErrNotFound
	}
	delete(f.byID, id)
	return nil
}

func TestProducerHandler_Create_ReturnsCreatedProducer(t *testing.T) {
	store := &fakeProducerStore{}
	h := handler.NewProducerHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	body := `{"name":"Joao","whatsapp_phone":"5511999999999","city":"Campinas","state":"SP"}`
	req := httptest.NewRequest(http.MethodPost, "/producers", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var got model.Producer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "generated-id" {
		t.Errorf("expected ID generated-id, got %q", got.ID)
	}
	if got.Name != "Joao" {
		t.Errorf("expected Name Joao, got %q", got.Name)
	}
}

func TestProducerHandler_GetByID_ReturnsProducer(t *testing.T) {
	store := &fakeProducerStore{byID: map[string]model.Producer{
		"prod-1": {ID: "prod-1", Name: "Maria"},
	}}
	h := handler.NewProducerHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/producers/prod-1", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProducerHandler_Update_Success(t *testing.T) {
	store := &fakeProducerStore{byID: map[string]model.Producer{
		"prod-1": {ID: "prod-1", Name: "Maria Antiga"},
	}}
	h := handler.NewProducerHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	body := `{"name":"Maria Atualizada","whatsapp_phone":"5511888888888","city":"Limeira","state":"SP"}`
	req := httptest.NewRequest(http.MethodPut, "/producers/prod-1", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.byID["prod-1"].Name != "Maria Atualizada" {
		t.Errorf("expected updated name, got %s", store.byID["prod-1"].Name)
	}
}

func TestProducerHandler_Delete_Success(t *testing.T) {
	store := &fakeProducerStore{byID: map[string]model.Producer{
		"prod-1": {ID: "prod-1", Name: "Maria"},
	}}
	h := handler.NewProducerHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodDelete, "/producers/prod-1", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if _, exists := store.byID["prod-1"]; exists {
		t.Fatalf("expected producer to be deleted")
	}
}
