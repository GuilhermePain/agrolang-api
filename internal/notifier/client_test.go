package notifier_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/notifier"
)

func TestClient_Notify_SendsPhoneAndMessageWithAPIKeyHeader(t *testing.T) {
	var gotBody map[string]string
	var gotAPIKey string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAPIKey = r.Header.Get("apikey")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := notifier.NewClient(server.URL, "test-api-key")

	err := client.Notify(context.Background(), "5511999999999", "mensagem de teste")
	if err != nil {
		t.Fatalf("Notify: %v", err)
	}

	if gotAPIKey != "test-api-key" {
		t.Errorf("expected apikey header 'test-api-key', got %q", gotAPIKey)
	}
	if gotBody["phone"] != "5511999999999" {
		t.Errorf("expected phone 5511999999999, got %q", gotBody["phone"])
	}
	if gotBody["message"] != "mensagem de teste" {
		t.Errorf("expected message 'mensagem de teste', got %q", gotBody["message"])
	}
}

func TestClient_Notify_NonOKStatus_ReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := notifier.NewClient(server.URL, "test-api-key")

	err := client.Notify(context.Background(), "5511999999999", "mensagem de teste")
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}
