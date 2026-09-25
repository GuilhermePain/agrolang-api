package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

func createTestProperty(t *testing.T, producerRepo *repository.ProducerRepository, propertyRepo *repository.PropertyRepository) model.Property {
	t.Helper()
	producer := createTestProducer(t, producerRepo)
	property, err := propertyRepo.Create(context.Background(), model.Property{
		ProducerID: producer.ID,
		Latitude:   -22.9,
		Longitude:  -47.06,
		Crop:       "Milho",
		SoilType:   "Argiloso",
		CropStage:  model.CropStageFlowering,
	})
	if err != nil {
		t.Fatalf("create test property: %v", err)
	}
	return property
}

func TestAlertRepository_CreateAndListByProperty(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	alerts := repository.NewAlertRepository(pool)
	ctx := context.Background()

	property := createTestProperty(t, producers, properties)

	start := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	end := start.Add(48 * time.Hour)

	created, err := alerts.Create(ctx, model.Alert{
		PropertyID:  property.ID,
		Level:       "critical",
		AlertType:   "frost",
		PeriodStart: start,
		PeriodEnd:   end,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated ID")
	}
	if created.TriggeredAt.IsZero() {
		t.Fatal("expected TriggeredAt to be set")
	}
	if created.ResolvedAt != nil {
		t.Fatal("expected ResolvedAt to be nil for new alert")
	}

	list, err := alerts.ListByProperty(ctx, property.ID)
	if err != nil {
		t.Fatalf("ListByProperty: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(list))
	}
	if list[0].AlertType != "frost" {
		t.Errorf("expected AlertType frost, got %q", list[0].AlertType)
	}
	if list[0].Level != "critical" {
		t.Errorf("expected Level critical, got %q", list[0].Level)
	}
}

func TestAlertRepository_ListByProperty_EmptyWhenNone(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	alerts := repository.NewAlertRepository(pool)
	ctx := context.Background()

	property := createTestProperty(t, producers, properties)

	list, err := alerts.ListByProperty(ctx, property.ID)
	if err != nil {
		t.Fatalf("ListByProperty: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(list))
	}
}

func TestAlertRepository_ListRecent_OrdersByTriggeredAtDesc(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	alerts := repository.NewAlertRepository(pool)
	ctx := context.Background()

	property := createTestProperty(t, producers, properties)

	start := time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		_, err := alerts.Create(ctx, model.Alert{
			PropertyID:  property.ID,
			Level:       "medium",
			AlertType:   "heavy_rain",
			PeriodStart: start,
			PeriodEnd:   start.Add(24 * time.Hour),
		})
		if err != nil {
			t.Fatalf("create alert %d: %v", i, err)
		}
	}

	recent, err := alerts.ListRecent(ctx, 2)
	if err != nil {
		t.Fatalf("ListRecent: %v", err)
	}
	if len(recent) != 2 {
		t.Fatalf("expected 2 alerts (limit), got %d", len(recent))
	}
}
