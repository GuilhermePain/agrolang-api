package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

type fakeSessionLookup struct {
	byHash map[string]model.User
}

func (f *fakeSessionLookup) UserBySession(ctx context.Context, tokenHash string) (model.User, error) {
	u, ok := f.byHash[tokenHash]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return u, nil
}

func serveThroughSession(lookup handler.SessionLookup, req *http.Request) (*httptest.ResponseRecorder, *model.User) {
	var seen *model.User
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u, ok := handler.UserFromContext(r.Context()); ok {
			seen = &u
		}
		w.WriteHeader(http.StatusNoContent)
	})
	rec := httptest.NewRecorder()
	handler.RequireSession(lookup)(next).ServeHTTP(rec, req)
	return rec, seen
}

func TestRequireSession_ValidTokenPutsUserInContext(t *testing.T) {
	lookup := &fakeSessionLookup{byHash: map[string]model.User{
		auth.HashToken("good-token"): {ID: "user-1"},
	}}
	req := httptest.NewRequest(http.MethodGet, "/anything", nil)
	req.Header.Set("Authorization", "Bearer good-token")

	rec, seen := serveThroughSession(lookup, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected next handler to run, got %d", rec.Code)
	}
	if seen == nil || seen.ID != "user-1" {
		t.Fatalf("expected user-1 in context, got %+v", seen)
	}
}

func TestRequireSession_RejectsMissingOrUnknownToken(t *testing.T) {
	lookup := &fakeSessionLookup{byHash: map[string]model.User{}}

	cases := map[string]string{
		"no header":      "",
		"wrong scheme":   "Basic abc",
		"empty bearer":   "Bearer ",
		"unknown bearer": "Bearer nope",
	}
	for name, header := range cases {
		t.Run(name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/anything", nil)
			if header != "" {
				req.Header.Set("Authorization", header)
			}
			rec, _ := serveThroughSession(lookup, req)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", rec.Code)
			}
		})
	}
}
