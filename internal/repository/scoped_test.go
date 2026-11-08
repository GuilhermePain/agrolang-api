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

// ownedFixture creates a producer + property (+ an unresolved alert) for a
// user, so the isolation tests below can compare two accounts.
type ownedFixture struct {
	producer model.Producer
	property model.Property
	alert    model.Alert
}

func makeOwnedFixture(t *testing.T, ctx context.Context, producers *repository.ProducerRepository, properties *repository.PropertyRepository, alerts *repository.AlertRepository, userID, name string) ownedFixture {
	t.Helper()
	producer, err := producers.CreateForUser(ctx, userID, model.Producer{Name: name, WhatsAppPhone: "55" + name, City: "X", State: "SP"})
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
	alert, err := alerts.Create(ctx, model.Alert{
		PropertyID: property.ID, Level: "critical", AlertType: "frost",
		PeriodStart: now, PeriodEnd: now.Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Create alert: %v", err)
	}
	return ownedFixture{producer: producer, property: property, alert: alert}
}

func TestProducerRepository_UpdateAndDeleteForUser_Isolation(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	ctx := context.Background()
	userA := createTestUser(t, pool, "a@exemplo.com")
	userB := createTestUser(t, pool, "b@exemplo.com")

	mine, err := producers.CreateForUser(ctx, userA, model.Producer{Name: "Meu", WhatsAppPhone: "5511", City: "Campinas", State: "SP"})
	if err != nil {
		t.Fatalf("CreateForUser: %v", err)
	}

	hijack := model.Producer{ID: mine.ID, Name: "Invadido", WhatsAppPhone: "5599", City: "Y", State: "RJ"}
	if _, err := producers.UpdateForUser(ctx, userB, hijack); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound updating another user's producer, got %v", err)
	}
	if err := producers.DeleteForUser(ctx, mine.ID, userB); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound deleting another user's producer, got %v", err)
	}
	got, err := producers.GetByIDForUser(ctx, mine.ID, userA)
	if err != nil || got.Name != "Meu" || got.WhatsAppPhone != "5511" {
		t.Fatalf("expected producer untouched by user B, got %+v, %v", got, err)
	}

	updated, err := producers.UpdateForUser(ctx, userA, model.Producer{ID: mine.ID, Name: "Novo", WhatsAppPhone: "5512", City: "Limeira", State: "SP"})
	if err != nil || updated.Name != "Novo" || updated.City != "Limeira" {
		t.Fatalf("expected owner update to succeed, got %+v, %v", updated, err)
	}
	if err := producers.DeleteForUser(ctx, mine.ID, userA); err != nil {
		t.Fatalf("expected owner delete to succeed, got %v", err)
	}
	if _, err := producers.GetByIDForUser(ctx, mine.ID, userA); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected producer gone after owner delete, got %v", err)
	}
}

func TestPropertyRepository_UpdateDeleteListByProducerForUser_Isolation(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	alerts := repository.NewAlertRepository(pool)
	ctx := context.Background()
	userA := createTestUser(t, pool, "a@exemplo.com")
	userB := createTestUser(t, pool, "b@exemplo.com")

	a := makeOwnedFixture(t, ctx, producers, properties, alerts, userA, "11")
	b := makeOwnedFixture(t, ctx, producers, properties, alerts, userB, "22")

	// List by producer: only for the owner.
	list, err := properties.ListByProducerForUser(ctx, a.producer.ID, userA)
	if err != nil || len(list) != 1 || list[0].ID != a.property.ID {
		t.Fatalf("expected own producer's property, got %+v, %v", list, err)
	}
	if _, err := properties.ListByProducerForUser(ctx, a.producer.ID, userB); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound listing another user's producer, got %v", err)
	}

	edit := a.property
	edit.Crop = "Soja"
	edit.CropStage = model.CropStageHarvest

	// Foreign user cannot update A's property.
	if _, err := properties.UpdateForUser(ctx, userB, edit); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound updating another user's property, got %v", err)
	}
	// Owner cannot move the property under someone else's producer.
	moved := edit
	moved.ProducerID = b.producer.ID
	if _, err := properties.UpdateForUser(ctx, userA, moved); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound moving a property to another user's producer, got %v", err)
	}
	// Foreign user cannot pull A's property into their own producer either.
	pulled := edit
	pulled.ProducerID = b.producer.ID
	if _, err := properties.UpdateForUser(ctx, userB, pulled); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound pulling another user's property, got %v", err)
	}
	// Foreign user cannot delete it.
	if err := properties.DeleteForUser(ctx, a.property.ID, userB); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound deleting another user's property, got %v", err)
	}
	untouched, err := properties.GetByIDForUser(ctx, a.property.ID, userA)
	if err != nil || untouched.Crop != "Milho" || untouched.ProducerID != a.producer.ID {
		t.Fatalf("expected property untouched, got %+v, %v", untouched, err)
	}

	// Owner update with an empty producer_id keeps the current producer.
	keep := edit
	keep.ProducerID = ""
	updated, err := properties.UpdateForUser(ctx, userA, keep)
	if err != nil || updated.Crop != "Soja" || updated.ProducerID != a.producer.ID || updated.CropStage != model.CropStageHarvest {
		t.Fatalf("expected owner update keeping producer, got %+v, %v", updated, err)
	}

	// Owner may move it between their own producers.
	second, err := producers.CreateForUser(ctx, userA, model.Producer{Name: "Segundo", WhatsAppPhone: "5513", City: "Z", State: "SP"})
	if err != nil {
		t.Fatalf("CreateForUser second producer: %v", err)
	}
	ownMove := edit
	ownMove.ProducerID = second.ID
	if got, err := properties.UpdateForUser(ctx, userA, ownMove); err != nil || got.ProducerID != second.ID {
		t.Fatalf("expected move between own producers to succeed, got %+v, %v", got, err)
	}

	if err := properties.DeleteForUser(ctx, a.property.ID, userA); err != nil {
		t.Fatalf("expected owner delete to succeed, got %v", err)
	}
	if _, err := properties.GetByIDForUser(ctx, a.property.ID, userA); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected property gone after owner delete, got %v", err)
	}
	if _, err := properties.GetByIDForUser(ctx, b.property.ID, userB); err != nil {
		t.Fatalf("user B's property must be untouched, got %v", err)
	}
}

func TestAlertRepository_ResolveForUser_Isolation(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	alerts := repository.NewAlertRepository(pool)
	ctx := context.Background()
	userA := createTestUser(t, pool, "a@exemplo.com")
	userB := createTestUser(t, pool, "b@exemplo.com")

	a := makeOwnedFixture(t, ctx, producers, properties, alerts, userA, "11")

	if err := alerts.ResolveForUser(ctx, a.alert.ID, userB); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound resolving another user's alert, got %v", err)
	}
	if active, err := alerts.HasActiveAlert(ctx, a.property.ID, "frost"); err != nil || !active {
		t.Fatalf("expected alert still active after foreign resolve, got %v, %v", active, err)
	}

	if err := alerts.ResolveForUser(ctx, a.alert.ID, userA); err != nil {
		t.Fatalf("expected owner resolve to succeed, got %v", err)
	}
	if active, err := alerts.HasActiveAlert(ctx, a.property.ID, "frost"); err != nil || active {
		t.Fatalf("expected alert resolved, got active=%v, %v", active, err)
	}
	if err := alerts.ResolveForUser(ctx, a.alert.ID, userA); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound resolving an already-resolved alert, got %v", err)
	}
}

func TestNotificationRepository_ScopedByUser(t *testing.T) {
	pool := testPool(t)
	producers := repository.NewProducerRepository(pool)
	properties := repository.NewPropertyRepository(pool)
	alerts := repository.NewAlertRepository(pool)
	notifications := repository.NewNotificationRepository(pool)
	ctx := context.Background()
	userA := createTestUser(t, pool, "a@exemplo.com")
	userB := createTestUser(t, pool, "b@exemplo.com")

	a := makeOwnedFixture(t, ctx, producers, properties, alerts, userA, "11")
	b := makeOwnedFixture(t, ctx, producers, properties, alerts, userB, "22")

	for _, f := range []ownedFixture{a, b} {
		alertID, propertyID := f.alert.ID, f.property.ID
		if _, err := notifications.Create(ctx, model.NotificationLog{
			AlertID: &alertID, ProducerID: f.producer.ID, PropertyID: &propertyID,
			Phone: f.producer.WhatsAppPhone, Message: "geada", Status: model.NotificationStatusSent,
		}); err != nil {
			t.Fatalf("Create notification log: %v", err)
		}
	}

	recentA, err := notifications.ListRecentForUser(ctx, userA, 50)
	if err != nil || len(recentA) != 1 || recentA[0].ProducerID != a.producer.ID {
		t.Fatalf("expected only user A's log, got %+v, %v", recentA, err)
	}

	byProducer, err := notifications.ListByProducerForUser(ctx, a.producer.ID, userA)
	if err != nil || len(byProducer) != 1 {
		t.Fatalf("expected user A's producer log, got %+v, %v", byProducer, err)
	}
	if _, err := notifications.ListByProducerForUser(ctx, a.producer.ID, userB); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound listing another user's producer logs, got %v", err)
	}

	userC := createTestUser(t, pool, "c@exemplo.com")
	empty, err := notifications.ListRecentForUser(ctx, userC, 50)
	if err != nil || empty == nil || len(empty) != 0 {
		t.Fatalf("expected empty non-nil list for a user without producers, got %#v, %v", empty, err)
	}
}
