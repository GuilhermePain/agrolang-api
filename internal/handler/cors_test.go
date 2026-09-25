package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
)

func TestWithCORS_SetsAllowOriginHeaderOnResponse(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	wrapped := handler.WithCORS(inner)

	req := httptest.NewRequest(http.MethodGet, "/properties", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("expected Access-Control-Allow-Origin '*', got %q", got)
	}
}

func TestWithCORS_HandlesPreflightOptionsRequest(t *testing.T) {
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("inner handler should not be called for OPTIONS preflight")
	})
	wrapped := handler.WithCORS(inner)

	req := httptest.NewRequest(http.MethodOptions, "/properties", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204 for OPTIONS preflight, got %d", rec.Code)
	}
}
