package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

const (
	sessionTTL       = 30 * 24 * time.Hour
	maxAuthBodyBytes = 1 << 20
)

type Signuper interface {
	Signup(ctx context.Context, in repository.SignupInput) (repository.SignupResult, error)
}

type AuthStore interface {
	GetByEmail(ctx context.Context, email string) (model.User, error)
	CreateSession(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	DeleteSession(ctx context.Context, tokenHash string) error
	DeleteExpiredSessions(ctx context.Context) error
}

type AuthHandler struct {
	signuper Signuper
	store    AuthStore
	// dummyHash lets login spend the same bcrypt time for unknown emails,
	// so response timing doesn't reveal which emails are registered.
	dummyHash string
}

func NewAuthHandler(signuper Signuper, store AuthStore) *AuthHandler {
	dummy, err := auth.HashPassword("not-a-real-password")
	if err != nil {
		// Startup-time wiring: without the dummy hash, logins for unknown
		// emails would return fast and leak which emails are registered, so
		// the process cannot authenticate safely.
		panic("handler: cannot compute dummy password hash for login timing: " + err.Error())
	}
	return &AuthHandler{signuper: signuper, store: store, dummyHash: dummy}
}

func (h *AuthHandler) RegisterPublic(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/signup", h.signup)
	mux.HandleFunc("POST /auth/login", h.login)
}

func (h *AuthHandler) RegisterProtected(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/logout", h.logout)
	mux.HandleFunc("GET /auth/me", h.me)
}

type authResponse struct {
	Token     string     `json:"token"`
	ExpiresAt time.Time  `json:"expires_at"`
	User      model.User `json:"user"`
}

func (h *AuthHandler) signup(w http.ResponseWriter, r *http.Request) {
	var req signupRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if fields := req.validate(); len(fields) > 0 {
		writeFieldErrors(w, http.StatusBadRequest, "invalid signup data", fields)
		return
	}

	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create account")
		return
	}
	token, tokenHash, err := auth.NewToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create account")
		return
	}
	expiresAt := time.Now().Add(sessionTTL)

	user := model.User{
		Email:         strings.ToLower(strings.TrimSpace(req.Email)),
		PasswordHash:  passwordHash,
		Profile:       req.Profile,
		Name:          strings.TrimSpace(req.Name),
		WhatsAppPhone: req.WhatsAppPhone,
		Concerns:      req.Concerns,
		Onboarding:    req.Onboarding,
	}
	if req.Organization != nil && strings.TrimSpace(*req.Organization) != "" {
		org := strings.TrimSpace(*req.Organization)
		user.Organization = &org
	}

	input := repository.SignupInput{
		User:             user,
		SessionTokenHash: tokenHash,
		SessionExpiresAt: expiresAt,
	}
	if req.Profile == model.ProfileProducer {
		input.Producer = &model.Producer{
			City:  strings.TrimSpace(req.Producer.City),
			State: strings.ToUpper(strings.TrimSpace(req.Producer.State)),
		}
		for _, c := range req.Property.Crops {
			input.Properties = append(input.Properties, model.Property{
				Latitude:  *req.Property.Latitude,
				Longitude: *req.Property.Longitude,
				Crop:      strings.TrimSpace(c.Crop),
				SoilType:  req.Property.SoilType,
				CropStage: c.CropStage,
			})
		}
	}

	result, err := h.signuper.Signup(r.Context(), input)
	if errors.Is(err, repository.ErrEmailTaken) {
		writeFieldErrors(w, http.StatusConflict, "email already registered",
			map[string]string{"email": "Este e-mail já tem cadastro."})
		return
	}
	if err != nil {
		log.Printf("handler: signup: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to create account")
		return
	}

	writeJSON(w, http.StatusCreated, authResponse{Token: token, ExpiresAt: expiresAt, User: result.User})
}

func (h *AuthHandler) login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxAuthBodyBytes)).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.store.GetByEmail(r.Context(), strings.ToLower(strings.TrimSpace(req.Email)))
	found := err == nil
	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		log.Printf("handler: login lookup: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to log in")
		return
	}

	hash := h.dummyHash
	if found {
		hash = user.PasswordHash
	}
	passwordOK := auth.CheckPassword(hash, req.Password)
	if !found || !passwordOK {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	if err := h.store.DeleteExpiredSessions(r.Context()); err != nil {
		log.Printf("handler: purge expired sessions: %v", err)
	}

	token, tokenHash, err := auth.NewToken()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to log in")
		return
	}
	expiresAt := time.Now().Add(sessionTTL)
	if err := h.store.CreateSession(r.Context(), user.ID, tokenHash, expiresAt); err != nil {
		log.Printf("handler: create session: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to log in")
		return
	}

	writeJSON(w, http.StatusOK, authResponse{Token: token, ExpiresAt: expiresAt, User: user})
}

func (h *AuthHandler) logout(w http.ResponseWriter, r *http.Request) {
	hash, ok := tokenHashFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := h.store.DeleteSession(r.Context(), hash); err != nil {
		log.Printf("handler: delete session: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to log out")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) me(w http.ResponseWriter, r *http.Request) {
	user, ok := requireUser(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, user)
}
