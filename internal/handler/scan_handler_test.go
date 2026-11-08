package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/handler"
)

func newScanMux(run func(ctx context.Context) error) *http.ServeMux {
	mux := http.NewServeMux()
	handler.NewScanHandler(run).Register(mux)
	return mux
}

func TestScanHandler_Run_Success(t *testing.T) {
	calls := 0
	mux := newScanMux(func(ctx context.Context) error { calls++; return nil })

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, authed(httptest.NewRequest(http.MethodPost, "/scan/run", nil)))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if calls != 1 {
		t.Fatalf("expected the scan to run once, ran %d times", calls)
	}
}

func TestScanHandler_Run_Failure_Returns500(t *testing.T) {
	mux := newScanMux(func(ctx context.Context) error { return errors.New("boom") })

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, authed(httptest.NewRequest(http.MethodPost, "/scan/run", nil)))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rec.Code)
	}
}

func TestScanHandler_WithoutUser_Returns401(t *testing.T) {
	calls := 0
	mux := newScanMux(func(ctx context.Context) error { calls++; return nil })

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/scan/run", nil))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	if calls != 0 {
		t.Fatalf("unauthenticated request must not trigger a scan")
	}
}
