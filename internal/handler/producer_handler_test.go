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
	if got.WhatsAppPhone != "5511999999999" {
		t.Errorf("expected WhatsAppPhone 5511999999999, got %q", got.WhatsAppPhone)
	}

	rawBody := rec.Body.String()
	if !bytes.Contains([]byte(rawBody), []byte(`"whatsapp_phone":"5511999999999"`)) {
		t.Errorf("expected response JSON to use snake_case whatsapp_phone key, got %s", rawBody)
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
	var got model.Producer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.Name != "Maria" {
		t.Errorf("expected Name Maria, got %q", got.Name)
	}
}

func TestProducerHandler_GetByID_NotFound_Returns404(t *testing.T) {
	store := &fakeProducerStore{byID: map[string]model.Producer{}}
	h := handler.NewProducerHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/producers/missing", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProducerHandler_List_ReturnsAllProducers(t *testing.T) {
	store := &fakeProducerStore{all: []model.Producer{
		{ID: "prod-1", Name: "Maria"},
		{ID: "prod-2", Name: "Joao"},
	}}
	h := handler.NewProducerHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodGet, "/producers", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.Producer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 producers, got %d", len(got))
	}
}

func TestProducerHandler_Create_InvalidJSON_Returns400(t *testing.T) {
	store := &fakeProducerStore{}
	h := handler.NewProducerHandler(store)
	mux := http.NewServeMux()
	h.Register(mux)

	req := httptest.NewRequest(http.MethodPost, "/producers", bytes.NewBufferString("not json"))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}
