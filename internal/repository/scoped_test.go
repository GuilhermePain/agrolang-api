package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

func TestProducerRepository_ScopedByUser(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewProducerRepository(pool)
	ctx := context.Background()
	userA := createTestUser(t, pool, "a@exemplo.com")
	userB := createTestUser(t, pool, "b@exemplo.com")

	mine, err := repo.CreateForUser(ctx, userA, model.Producer{Name: "Meu", WhatsAppPhone: "5511", City: "Campinas", State: "SP"})
	if err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}
	if _, err := repo.CreateForUser(ctx, userB, model.Producer{Name: "Dele", WhatsAppPhone: "5522", City: "Recife", State: "PE"}); err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}

	list, err := repo.ListByUser(ctx, userA)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if len(list) != 1 || list[0].ID != mine.ID {
		t.Fatalf("expected only user A's producer, got %+v", list)
	}

	if _, err := repo.GetByIDForUser(ctx, mine.ID, userB); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for another user's producer, got %v", err)
	}
	if got, err := repo.GetByIDForUser(ctx, mine.ID, userA); err != nil || got.ID != mine.ID {
		t.Fatalf("expected own producer, got %+v, %v", got, err)
	}
}

func TestProducerRepository_ListByUser_EmptyIsNotNil(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewProducerRepository(pool)
	userID := createTestUser(t, pool, "a@exemplo.com")

	list, err := repo.ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListByUser: %v", err)
	}
	if list == nil || len(list) != 0 {
		t.Fatalf("expected empty non-nil slice, got %#v", list)
	}
}

func TestProducerRepository_GetByID_IncludesOwnerConcerns(t *testing.T) {
	pool := testPool(t)
	repo := repository.NewProducerRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, pool, "a@exemplo.com")
	if _, err := pool.Exec(ctx, `UPDATE users SET concerns = '{frost,heavy_rain}' WHERE id = $1`, userID); err != nil {
		t.Fatalf("set concerns: %v", err)
	}

	owned, err := repo.CreateForUser(ctx, userID, model.Producer{Name: "Meu", WhatsAppPhone: "5511", City: "Campinas", State: "SP"})
	if err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}
	legacy, err := repo.Create(ctx, model.Producer{Name: "Antigo", WhatsAppPhone: "5533", City: "Recife", State: "PE"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(ctx, owned.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if len(got.Concerns) != 2 || got.Concerns[0] != "frost" {
		t.Errorf("expected owner concerns [frost heavy_rain], got %v", got.Concerns)
	}

	legacyGot, err := repo.GetByID(ctx, legacy.ID)
	if err != nil {
		t.Fatalf("GetByID legacy: %v", err)
	}
	if len(legacyGot.Concerns) != 0 {
		t.Errorf("expected no concerns for a producer without owner, got %v", legacyGot.Concerns)
	}
}

func TestPropertyRepository_ScopedByUser(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	ctx := context.Background()
	userA := createTestUser(t, pool, "a@exemplo.com")
	userB := createTestUser(t, pool, "b@exemplo.com")

	producerA, err := producers.CreateForUser(ctx, userA, model.Producer{Name: "A", WhatsAppPhone: "5511", City: "Campinas", State: "SP"})
	if err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}

	newProperty := model.Property{
		ProducerID: producerA.ID, Latitude: -22.9, Longitude: -47.06,
		Crop: "Milho", SoilType: "argiloso", CropStage: model.CropStageFlowering,
	}

	if _, err := properties.CreateForUser(ctx, userB, newProperty); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound creating under another user's producer, got %v", err)
	}

	created, err := properties.CreateForUser(ctx, userA, newProperty)
	if err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}
	if created.Latitude != -22.9 || created.Longitude != -47.06 {
		t.Errorf("expected coordinates round-trip, got %v,%v", created.Latitude, created.Longitude)
	}

	list, err := properties.ListByUser(ctx, userA)
	if err != nil || len(list) != 1 {
		t.Fatalf("expected 1 property for user A, got %v, %v", list, err)
	}
	other, err := properties.ListByUser(ctx, userB)
	if err != nil || other == nil || len(other) != 0 {
		t.Fatalf("expected empty non-nil list for user B, got %#v, %v", other, err)
	}

	if _, err := properties.GetByIDForUser(ctx, created.ID, userB); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for another user's property, got %v", err)
	}
	if got, err := properties.GetByIDForUser(ctx, created.ID, userA); err != nil || got.ID != created.ID {
		t.Fatalf("expected own property, got %+v, %v", got, err)
	}
}

func TestAlertRepository_ListRecentForUser_OnlyOwnAlerts(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	alerts := repository.NewAlertRepository(pool)
	ctx := context.Background()
	userA := createTestUser(t, pool, "a@exemplo.com")
	userB := createTestUser(t, pool, "b@exemplo.com")

	makeAlert := func(userID, email string) {
		t.Helper()
		producer, err := producers.CreateForUser(ctx, userID, model.Producer{Name: email, WhatsAppPhone: "5511", City: "X", State: "SP"})
		if err != nil {
			t.Fatalf("CreateForUser producer: %v", err)
		}
		property, err := properties.CreateForUser(ctx, userID, model.Property{
			ProducerID: producer.ID, Latitude: -22.9, Longitude: -47.06,
			Crop: "Milho", SoilType: "argiloso", CropStage: model.CropStageFlowering,
		})
		if err != nil {
			t.Fatalf("CreateForUser property: %v", err)
		}
		now := time.Now()
		if _, err := alerts.Create(ctx, model.Alert{
			PropertyID: property.ID, Level: "critical", AlertType: "frost",
			PeriodStart: now, PeriodEnd: now.Add(time.Hour),
		}); err != nil {
			t.Fatalf("Create alert: %v", err)
		}
	}
	makeAlert(userA, "a")
	makeAlert(userB, "b")

	got, err := alerts.ListRecentForUser(ctx, userA, 50)
	if err != nil {
		t.Fatalf("ListRecentForUser: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 alert for user A, got %d", len(got))
	}
}
