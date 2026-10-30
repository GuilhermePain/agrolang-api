package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type ctxKey int

const (
	userCtxKey ctxKey = iota
	tokenHashCtxKey
)

// ContextWithUser is exported so tests can authenticate a request without a
// session store.
func ContextWithUser(ctx context.Context, u model.User) context.Context {
	return context.WithValue(ctx, userCtxKey, u)
}

func UserFromContext(ctx context.Context) (model.User, bool) {
	u, ok := ctx.Value(userCtxKey).(model.User)
	return u, ok
}

func contextWithTokenHash(ctx context.Context, hash string) context.Context {
	return context.WithValue(ctx, tokenHashCtxKey, hash)
}

func tokenHashFromContext(ctx context.Context) (string, bool) {
	h, ok := ctx.Value(tokenHashCtxKey).(string)
	return h, ok
}

// requireUser reads the authenticated user set by RequireSession and writes
// a 401 when it is missing, so handlers never run unauthenticated.
func requireUser(w http.ResponseWriter, r *http.Request) (model.User, bool) {
	u, ok := UserFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return model.User{}, false
	}
	return u, true
}

type SessionLookup interface {
	UserBySession(ctx context.Context, tokenHash string) (model.User, error)
}

// RequireSession authenticates "Authorization: Bearer <token>" against live
// sessions and stores the user (and the token hash, for logout) in the
// request context.
func RequireSession(lookup SessionLookup) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := bearerToken(r)
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			hash := auth.HashToken(token)
			user, err := lookup.UserBySession(r.Context(), hash)
			if errors.Is(err, repository.ErrNotFound) {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to check session")
				return
			}

			ctx := contextWithTokenHash(ContextWithUser(r.Context(), user), hash)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	return token, token != ""
}
