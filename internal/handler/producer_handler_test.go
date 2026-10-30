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
	created     model.Producer
	createdFor  string
	lookedUpFor string
	listedFor   string
	byID        map[string]model.Producer
	all         []model.Producer
	createErr   error
	listErr     error
}

func (f *fakeProducerStore) CreateForUser(ctx context.Context, userID string, p model.Producer) (model.Producer, error) {
	if f.createErr != nil {
		return model.Producer{}, f.createErr
	}
	p.ID = "generated-id"
	f.created = p
	f.createdFor = userID
	return p, nil
}

func (f *fakeProducerStore) GetByIDForUser(ctx context.Context, id, userID string) (model.Producer, error) {
	f.lookedUpFor = userID
	p, ok := f.byID[id]
	if !ok {
		return model.Producer{}, repository.ErrNotFound
	}
	return p, nil
}

func (f *fakeProducerStore) ListByUser(ctx context.Context, userID string) ([]model.Producer, error) {
	f.listedFor = userID
	if f.listErr != nil {
		return nil, f.listErr
	}
	return f.all, nil
}

func newProducerMux(store *fakeProducerStore) *http.ServeMux {
	mux := http.NewServeMux()
	handler.NewProducerHandler(store).Register(mux)
	return mux
}

func TestProducerHandler_Create_ReturnsCreatedProducerOwnedByUser(t *testing.T) {
	store := &fakeProducerStore{}
	mux := newProducerMux(store)

	body := `{"name":"Joao","whatsapp_phone":"5511999999999","city":"Campinas","state":"SP"}`
	req := authed(httptest.NewRequest(http.MethodPost, "/producers", bytes.NewBufferString(body)))
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var got model.Producer
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got.ID != "generated-id" || got.Name != "Joao" || got.WhatsAppPhone != "5511999999999" {
		t.Errorf("unexpected producer: %+v", got)
	}
	if store.createdFor != testUserID {
		t.Errorf("expected producer created for %q, got %q", testUserID, store.createdFor)
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"whatsapp_phone":"5511999999999"`)) {
		t.Errorf("expected snake_case whatsapp_phone key, got %s", rec.Body.String())
	}
}

func TestProducerHandler_GetByID_ReturnsProducerScopedToUser(t *testing.T) {
	store := &fakeProducerStore{byID: map[string]model.Producer{"prod-1": {ID: "prod-1", Name: "Maria"}}}
	mux := newProducerMux(store)

	req := authed(httptest.NewRequest(http.MethodGet, "/producers/prod-1", nil))
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
	if store.lookedUpFor != testUserID {
		t.Errorf("expected lookup scoped to %q, got %q", testUserID, store.lookedUpFor)
	}
}

func TestProducerHandler_GetByID_NotFound_Returns404(t *testing.T) {
	mux := newProducerMux(&fakeProducerStore{byID: map[string]model.Producer{}})

	req := authed(httptest.NewRequest(http.MethodGet, "/producers/missing", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProducerHandler_List_ReturnsUsersProducers(t *testing.T) {
	store := &fakeProducerStore{all: []model.Producer{{ID: "prod-1", Name: "Maria"}, {ID: "prod-2", Name: "Joao"}}}
	mux := newProducerMux(store)

	req := authed(httptest.NewRequest(http.MethodGet, "/producers", nil))
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
	if store.listedFor != testUserID {
		t.Errorf("expected list scoped to %q, got %q", testUserID, store.listedFor)
	}
}

func TestProducerHandler_Create_InvalidJSON_Returns400(t *testing.T) {
	mux := newProducerMux(&fakeProducerStore{})

	req := authed(httptest.NewRequest(http.MethodPost, "/producers", bytes.NewBufferString("not json")))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProducerHandler_WithoutUser_Returns401(t *testing.T) {
	mux := newProducerMux(&fakeProducerStore{})

	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/producers"},
		{http.MethodGet, "/producers/prod-1"},
		{http.MethodPost, "/producers"},
	} {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(tc.method, tc.path, bytes.NewBufferString("{}")))
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: expected 401, got %d", tc.method, tc.path, rec.Code)
		}
	}
}
