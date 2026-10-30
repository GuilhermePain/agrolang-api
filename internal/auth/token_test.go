package auth_test

import (
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
)

func TestNewToken_ReturnsDistinctTokensWithMatchingHash(t *testing.T) {
	token1, hash1, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	token2, _, err := auth.NewToken()
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}

	if token1 == "" || token1 == token2 {
		t.Errorf("expected distinct non-empty tokens, got %q and %q", token1, token2)
	}
	if auth.HashToken(token1) != hash1 {
		t.Error("expected HashToken(token) to equal the returned hash")
	}
	if len(hash1) != 64 {
		t.Errorf("expected 64-char hex SHA-256, got %d chars", len(hash1))
	}
}

func TestHashToken_IsDeterministic(t *testing.T) {
	if auth.HashToken("abc") != auth.HashToken("abc") {
		t.Error("expected same input to give same hash")
	}
	if auth.HashToken("abc") == auth.HashToken("abd") {
		t.Error("expected different inputs to give different hashes")
	}
}
