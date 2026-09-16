package repository_test

import (
	"context"
	"testing"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

func createTestProducer(t *testing.T, repo *repository.ProducerRepository) model.Producer {
	t.Helper()
	p, err := repo.Create(context.Background(), model.Producer{
		Name: "Ana", WhatsAppPhone: "+5511988887777", City: "Campinas", State: "SP",
	})
	if err != nil {
		t.Fatalf("create test producer: %v", err)
	}
	return p
}

func TestPropertyRepository_CreateAndGet(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	ctx := context.Background()

	producer := createTestProducer(t, producers)

	property := model.Property{
		ProducerID: producer.ID,
		Latitude:   -22.9,
		Longitude:  -47.06,
		Crop:       "Milho",
		SoilType:   "Argiloso",
		CropStage:  model.CropStageFlowering,
	}

	created, err := properties.Create(ctx, property)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected generated ID")
	}

	got, err := properties.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.Crop != property.Crop {
		t.Errorf("expected Crop %q, got %q", property.Crop, got.Crop)
	}
	if got.CropStage != property.CropStage {
		t.Errorf("expected CropStage %q, got %q", property.CropStage, got.CropStage)
	}
	const epsilon = 0.0001
	if diff := got.Latitude - property.Latitude; diff > epsilon || diff < -epsilon {
		t.Errorf("expected Latitude %v, got %v", property.Latitude, got.Latitude)
	}
	if diff := got.Longitude - property.Longitude; diff > epsilon || diff < -epsilon {
		t.Errorf("expected Longitude %v, got %v", property.Longitude, got.Longitude)
	}
}

func TestPropertyRepository_ListAll_ReturnsAllProperties(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	ctx := context.Background()

	producer := createTestProducer(t, producers)

	for i := 0; i < 3; i++ {
		_, err := properties.Create(ctx, model.Property{
			ProducerID: producer.ID,
			Latitude:   -22.9,
			Longitude:  -47.06,
			Crop:       "Milho",
			SoilType:   "Argiloso",
			CropStage:  model.CropStageFlowering,
		})
		if err != nil {
			t.Fatalf("create property %d: %v", i, err)
		}
	}

	all, err := properties.ListAll(ctx)
	if err != nil {
		t.Fatalf("ListAll: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 properties, got %d", len(all))
	}
}
