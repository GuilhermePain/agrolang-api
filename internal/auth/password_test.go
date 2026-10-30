package auth_test

import (
	"strings"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/auth"
)

func TestHashAndCheckPassword(t *testing.T) {
	hash, err := auth.HashPassword("segredo123")
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if hash == "segredo123" {
		t.Fatal("hash must not equal the plain password")
	}
	if !auth.CheckPassword(hash, "segredo123") {
		t.Error("expected correct password to match")
	}
	if auth.CheckPassword(hash, "errada123") {
		t.Error("expected wrong password not to match")
	}
}

func TestHashPassword_TooLongReturnsError(t *testing.T) {
	_, err := auth.HashPassword(strings.Repeat("a", auth.MaxPasswordLength+1))
	if err == nil {
		t.Fatal("expected error for password longer than the bcrypt limit")
	}
}
