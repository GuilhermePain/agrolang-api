package repository_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://agrolang:agrolang@localhost:5432/agrolang?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		t.Fatalf("connect to test db: %v", err)
	}
	t.Cleanup(pool.Close)

	if _, err := pool.Exec(context.Background(), "TRUNCATE alerts, properties, producers CASCADE"); err != nil {
		t.Fatalf("truncate test db: %v", err)
	}

	return pool
}

func TestProducerRepository_CreateAndGet(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewProducerRepository(pool)
	ctx := context.Background()

	producer := model.Producer{
		Name:          "Seu Zé",
		WhatsAppPhone: "+5511999998888",
		City:          "Piracicaba",
		State:         "SP",
	}

	created, err := repo.Create(ctx, producer)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated ID")
	}

	got, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.Name != producer.Name {
		t.Errorf("expected Name %q, got %q", producer.Name, got.Name)
	}
	if got.WhatsAppPhone != producer.WhatsAppPhone {
		t.Errorf("expected WhatsAppPhone %q, got %q", producer.WhatsAppPhone, got.WhatsAppPhone)
	}
}

func TestProducerRepository_GetByID_NotFound(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewProducerRepository(pool)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for missing producer, got nil")
	}
}
