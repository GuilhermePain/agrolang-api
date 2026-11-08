package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

// fakeSessions rejects every token except the ones in users, keyed by the
// token hash RequireSession computes.
type fakeSessions struct {
	users map[string]model.User
}

func (f fakeSessions) UserBySession(_ context.Context, tokenHash string) (model.User, error) {
	if u, ok := f.users[tokenHash]; ok {
		return u, nil
	}
	return model.User{}, repository.ErrNotFound
}

// testRouter builds the real router with nil stores: protected routes must be
// rejected by RequireSession before any handler touches a store.
func testRouter(sessions handler.SessionLookup) http.Handler {
	return newRouter(routerDeps{
		sessions:      sessions,
		auth:          handler.NewAuthHandler(nil, nil),
		producers:     handler.NewProducerHandler(nil),
		properties:    handler.NewPropertyHandler(nil, nil, nil),
		alerts:        handler.NewAlertHandler(nil),
		notifications: handler.NewNotificationHandler(nil),
		webhook:       handler.NewWebhookHandler(nil),
		scan:          handler.NewScanHandler(nil),
	})
}

func do(h http.Handler, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestRouter_ProtectedRoutesRequireSession(t *testing.T) {
	router := testRouter(fakeSessions{})

	routes := []struct{ method, path string }{
		{"POST", "/auth/logout"},
		{"GET", "/auth/me"},
		{"POST", "/scan/run"},
		{"GET", "/producers"},
		{"POST", "/producers"},
		{"GET", "/producers/x"},
		{"PUT", "/producers/x"},
		{"DELETE", "/producers/x"},
		{"GET", "/producers/x/properties"},
		{"GET", "/producers/x/notifications"},
		{"GET", "/properties"},
		{"POST", "/properties"},
		{"GET", "/properties/x"},
		{"PUT", "/properties/x"},
		{"DELETE", "/properties/x"},
		{"GET", "/properties/x/alerts"},
		{"GET", "/properties/x/forecast"},
		{"GET", "/alerts"},
		{"POST", "/alerts/x/resolve"},
		{"GET", "/notifications"},
		{"GET", "/nope"},
	}
	auths := map[string]map[string]string{
		"no authorization header": nil,
		"bogus bearer token":      {"Authorization": "Bearer bogus"},
	}

	for _, rt := range routes {
		for name, headers := range auths {
			t.Run(rt.method+" "+rt.path+" / "+name, func(t *testing.T) {
				rec := do(router, rt.method, rt.path, "", headers)
				if rec.Code != http.StatusUnauthorized {
					t.Fatalf("status = %d, want 401; body=%s", rec.Code, rec.Body.String())
				}
			})
		}
	}
}

func TestRouter_PublicRoutesAreNotGated(t *testing.T) {
	router := testRouter(fakeSessions{})

	tests := []struct {
		name         string
		method, path string
		body         string
		wantStatus   int
		wantBody     string
	}{
		{"health", "GET", "/health", "", http.StatusOK, `{"status":"ok"}`},
		{"signup malformed body reaches handler", "POST", "/auth/signup", "{not json", http.StatusBadRequest, ""},
		{"login malformed body reaches handler", "POST", "/auth/login", "{not json", http.StatusBadRequest, ""},
		{"webhook empty object is ignored", "POST", "/webhooks/evolution", "{}", http.StatusOK, "ignored"},
		{"webhook event without remoteJid is ignored", "POST", "/webhooks/evolution",
			`{"event":"messages.update","data":{"status":"READ"}}`, http.StatusOK, "ignored"},
		// Only the declared method+path pairs are public. A wrong method does
		// not match the public mux's pattern, so it falls through to the
		// catch-all "/" route, i.e. the protected mux behind RequireSession.
		{"login with wrong method falls through to the session gate", "GET", "/auth/login", "", http.StatusUnauthorized, ""},
		{"signup with wrong method falls through to the session gate", "GET", "/auth/signup", "", http.StatusUnauthorized, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := do(router, tt.method, tt.path, tt.body, nil)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d; body=%s", rec.Code, tt.wantStatus, rec.Body.String())
			}
			if tt.wantBody != "" && !strings.Contains(rec.Body.String(), tt.wantBody) {
				t.Fatalf("body = %q, want it to contain %q", rec.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestRouter_ValidSessionReachesHandler(t *testing.T) {
	user := model.User{ID: "user-123", Email: "ana@example.com"}
	sessions := fakeSessions{users: map[string]model.User{auth.HashToken("good"): user}}
	router := testRouter(sessions)

	rec := do(router, "GET", "/auth/me", "", map[string]string{"Authorization": "Bearer good"})
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), user.ID) {
		t.Fatalf("body = %q, want it to contain user id %q", rec.Body.String(), user.ID)
	}

	rec = do(router, "GET", "/auth/me", "", map[string]string{"Authorization": "Bearer other"})
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unknown token: status = %d, want 401", rec.Code)
	}
}

func TestRouter_CORSPreflightNeedsNoSession(t *testing.T) {
	router := testRouter(fakeSessions{})

	rec := do(router, "OPTIONS", "/producers", "", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got == "" {
		t.Fatal("missing Access-Control-Allow-Origin on preflight")
	}
}
