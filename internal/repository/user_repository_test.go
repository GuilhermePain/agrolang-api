package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

// createTestUser inserts a bare producer-profile user and returns its ID.
func createTestUser(t *testing.T, pool *pgxpool.Pool, email string) string {
	t.Helper()

	var id string
	err := pool.QueryRow(context.Background(),
		`INSERT INTO users (email, password_hash, profile, name, whatsapp_phone)
		 VALUES ($1, 'hash', 'producer', 'Teste', '5519999999999')
		 RETURNING id`, email).Scan(&id)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	return id
}

func TestUserRepository_GetByEmail_IsCaseInsensitive(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewUserRepository(pool)
	id := createTestUser(t, pool, "Ana@Exemplo.com")

	got, err := repo.GetByEmail(context.Background(), "ana@exemplo.com")
	if err != nil {
		t.Fatalf("GetByEmail: %v", err)
	}
	if got.ID != id {
		t.Errorf("expected ID %q, got %q", id, got.ID)
	}
	if got.Concerns == nil || got.Onboarding == nil {
		t.Error("expected Concerns and Onboarding to be non-nil")
	}
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewUserRepository(pool)

	_, err := repo.GetByEmail(context.Background(), "ninguem@exemplo.com")
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestUserRepository_SessionLifecycle(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewUserRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool, "ana@exemplo.com")

	if err := repo.CreateSession(ctx, userID, "hash-1", time.Now().Add(time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	got, err := repo.UserBySession(ctx, "hash-1")
	if err != nil {
		t.Fatalf("UserBySession: %v", err)
	}
	if got.ID != userID {
		t.Errorf("expected user %q, got %q", userID, got.ID)
	}

	if err := repo.DeleteSession(ctx, "hash-1"); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if _, err := repo.UserBySession(ctx, "hash-1"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestUserRepository_ExpiredSessionIsIgnoredAndPurged(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewUserRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool, "ana@exemplo.com")

	if err := repo.CreateSession(ctx, userID, "old", time.Now().Add(-time.Hour)); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	if _, err := repo.UserBySession(ctx, "old"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for expired session, got %v", err)
	}

	if err := repo.DeleteExpiredSessions(ctx); err != nil {
		t.Fatalf("DeleteExpiredSessions: %v", err)
	}
	var count int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM sessions").Scan(&count); err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if count != 0 {
		t.Errorf("expected expired session purged, %d remain", count)
	}
}
