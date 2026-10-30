package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type fakeSignuper struct {
	got    repository.SignupInput
	result repository.SignupResult
	err    error
}

func (f *fakeSignuper) Signup(ctx context.Context, in repository.SignupInput) (repository.SignupResult, error) {
	f.got = in
	if f.err != nil {
		return repository.SignupResult{}, f.err
	}
	user := in.User
	user.ID = "user-1"
	return repository.SignupResult{User: user, Producer: in.Producer, Property: in.Property}, nil
}

type fakeAuthStore struct {
	users          map[string]model.User
	createdSession string
	deletedSession string
	purged         bool
}

func (f *fakeAuthStore) GetUserByEmail(ctx context.Context, email string) (model.User, error) {
	u, ok := f.users[email]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return u, nil
}

func (f *fakeAuthStore) CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	f.createdSession = tokenHash
	return nil
}

func (f *fakeAuthStore) DeleteSession(ctx context.Context, tokenHash string) error {
	f.deletedSession = tokenHash
	return nil
}

func (f *fakeAuthStore) DeleteExpiredSessions(ctx context.Context) error {
	f.purged = true
	return nil
}

const validProducerSignup = `{
	"profile":"producer","name":"José da Silva","whatsapp_phone":"5519999999999",
	"email":"Jose@Exemplo.com","password":"segredo123","organization":null,
	"concerns":["frost","severe_water_stress"],
	"onboarding":{"work_style":"family","areas_bucket":"1"},
	"producer":{"city":"Campinas","state":"sp"},
	"property":{"latitude":-22.9099,"longitude":-47.0626,"crop":"Milho","soil_type":"argiloso","crop_stage":"flowering"}
}`

const validTechnicianSignup = `{
	"profile":"technician","name":"Ana","whatsapp_phone":"5519988887777",
	"email":"ana@coop.com","password":"segredo123","organization":"Coopera Campo",
	"concerns":[],"onboarding":{"work_style":"technician","producers_bucket":"11-40"}
}`

func newAuthMux(signuper *fakeSignuper, store *fakeAuthStore) *http.ServeMux {
	h := handler.NewAuthHandler(signuper, store)
	mux := http.NewServeMux()
	h.RegisterPublic(mux)
	h.RegisterProtected(mux)
	return mux
}

func postJSON(mux http.Handler, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	return rec
}

func TestAuthHandler_Signup_Producer_Returns201WithSession(t *testing.T) {
	signuper := &fakeSignuper{}
	mux := newAuthMux(signuper, &fakeAuthStore{})

	rec := postJSON(mux, "/auth/signup", validProducerSignup)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Token     string     `json:"token"`
		ExpiresAt time.Time  `json:"expires_at"`
		User      model.User `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Token == "" || resp.User.ID != "user-1" {
		t.Errorf("expected token and user, got %+v", resp)
	}
	if !resp.ExpiresAt.After(time.Now().Add(29 * 24 * time.Hour)) {
		t.Errorf("expected ~30 day expiry, got %v", resp.ExpiresAt)
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Errorf("response must not expose password data: %s", rec.Body.String())
	}

	in := signuper.got
	if in.User.Email != "jose@exemplo.com" {
		t.Errorf("expected normalized email, got %q", in.User.Email)
	}
	if in.User.PasswordHash == "" || in.User.PasswordHash == "segredo123" || !auth.CheckPassword(in.User.PasswordHash, "segredo123") {
		t.Error("expected a bcrypt hash of the password")
	}
	if in.SessionTokenHash != auth.HashToken(resp.Token) {
		t.Error("expected stored session hash to match the returned token")
	}
	if in.Producer == nil || in.Producer.State != "SP" {
		t.Errorf("expected producer with upper-cased state, got %+v", in.Producer)
	}
	if in.Property == nil || in.Property.CropStage != model.CropStageFlowering {
		t.Errorf("expected property with flowering stage, got %+v", in.Property)
	}
}

func TestAuthHandler_Signup_Technician_HasNoProducerOrProperty(t *testing.T) {
	signuper := &fakeSignuper{}
	mux := newAuthMux(signuper, &fakeAuthStore{})

	rec := postJSON(mux, "/auth/signup", validTechnicianSignup)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	if signuper.got.Producer != nil || signuper.got.Property != nil {
		t.Errorf("expected no producer/property, got %+v / %+v", signuper.got.Producer, signuper.got.Property)
	}
	if signuper.got.User.Organization == nil || *signuper.got.User.Organization != "Coopera Campo" {
		t.Errorf("expected organization saved, got %v", signuper.got.User.Organization)
	}
}

func TestAuthHandler_Signup_InvalidBody_Returns400WithFieldMessages(t *testing.T) {
	mux := newAuthMux(&fakeSignuper{}, &fakeAuthStore{})

	body := `{"profile":"producer","name":"J","whatsapp_phone":"123","email":"nope","password":"curta",
		"concerns":["tornado"],"onboarding":{"work_style":"alien"}}`
	rec := postJSON(mux, "/auth/signup", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp struct {
		Fields map[string]string `json:"fields"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	for _, key := range []string{"name", "whatsapp_phone", "email", "password", "concerns", "onboarding", "producer", "property"} {
		if resp.Fields[key] == "" {
			t.Errorf("expected a message for field %q, got fields %v", key, resp.Fields)
		}
	}
}

func TestAuthHandler_Signup_TechnicianWithProducer_Returns400(t *testing.T) {
	mux := newAuthMux(&fakeSignuper{}, &fakeAuthStore{})

	body := strings.Replace(validTechnicianSignup, `"concerns":[]`, `"concerns":[],"producer":{"city":"X","state":"SP"}`, 1)
	rec := postJSON(mux, "/auth/signup", body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"producer"`) {
		t.Errorf("expected a producer field error, got %s", rec.Body.String())
	}
}

func TestAuthHandler_Signup_EmailTaken_Returns409(t *testing.T) {
	mux := newAuthMux(&fakeSignuper{err: repository.ErrEmailTaken}, &fakeAuthStore{})

	rec := postJSON(mux, "/auth/signup", validProducerSignup)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"email"`) {
		t.Errorf("expected an email field error, got %s", rec.Body.String())
	}
}

func TestAuthHandler_Login(t *testing.T) {
	hash, err := auth.HashPassword("segredo123")
	if err != nil {
		t.Fatalf("hash: %v", err)
	}
	store := &fakeAuthStore{users: map[string]model.User{
		"jose@exemplo.com": {ID: "user-1", Email: "jose@exemplo.com", PasswordHash: hash},
	}}
	mux := newAuthMux(&fakeSignuper{}, store)

	t.Run("correct credentials return a session", func(t *testing.T) {
		rec := postJSON(mux, "/auth/login", `{"email":" Jose@Exemplo.com ","password":"segredo123"}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
		}
		var resp struct {
			Token string     `json:"token"`
			User  model.User `json:"user"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if resp.Token == "" || resp.User.ID != "user-1" {
			t.Errorf("expected token and user, got %+v", resp)
		}
		if store.createdSession != auth.HashToken(resp.Token) {
			t.Error("expected the session hash to be stored")
		}
		if !store.purged {
			t.Error("expected expired sessions to be purged on login")
		}
	})

	t.Run("wrong password and unknown email give the same 401", func(t *testing.T) {
		wrong := postJSON(mux, "/auth/login", `{"email":"jose@exemplo.com","password":"errada123"}`)
		unknown := postJSON(mux, "/auth/login", `{"email":"ninguem@exemplo.com","password":"segredo123"}`)
		if wrong.Code != http.StatusUnauthorized || unknown.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 for both, got %d and %d", wrong.Code, unknown.Code)
		}
		if wrong.Body.String() != unknown.Body.String() {
			t.Errorf("expected identical bodies, got %q and %q", wrong.Body.String(), unknown.Body.String())
		}
	})
}

func TestAuthHandler_Me_ReturnsUserFromContext(t *testing.T) {
	mux := newAuthMux(&fakeSignuper{}, &fakeAuthStore{})

	req := authed(httptest.NewRequest(http.MethodGet, "/auth/me", nil))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got model.User
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got.ID != testUserID {
		t.Errorf("expected %q, got %q", testUserID, got.ID)
	}
}

func TestAuthHandler_Me_WithoutUser_Returns401(t *testing.T) {
	mux := newAuthMux(&fakeSignuper{}, &fakeAuthStore{})

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/auth/me", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestAuthHandler_Logout_DeletesCurrentSession(t *testing.T) {
	store := &fakeAuthStore{}
	h := handler.NewAuthHandler(&fakeSignuper{}, store)
	mux := http.NewServeMux()
	h.RegisterProtected(mux)

	lookup := &fakeSessionLookup{byHash: map[string]model.User{
		auth.HashToken("tok"): {ID: testUserID},
	}}
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer tok")
	rec := httptest.NewRecorder()
	handler.RequireSession(lookup)(mux).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}
	if store.deletedSession != auth.HashToken("tok") {
		t.Errorf("expected session %q deleted, got %q", auth.HashToken("tok"), store.deletedSession)
	}
}
