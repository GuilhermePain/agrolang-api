package repository_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
)

func producerSignup(email string) repository.SignupInput {
	return repository.SignupInput{
		User: model.User{
			Email: email, PasswordHash: "hash", Profile: model.ProfileProducer,
			Name: "José", WhatsAppPhone: "5519999999999",
			Concerns:   []string{"frost", "severe_water_stress"},
			Onboarding: map[string]string{"work_style": "family", "areas_bucket": "1"},
		},
		Producer: &model.Producer{City: "Campinas", State: "SP"},
		Property: &model.Property{
			Latitude: -22.9099, Longitude: -47.0626,
			Crop: "Milho", SoilType: "argiloso", CropStage: model.CropStageFlowering,
		},
		SessionTokenHash: "token-hash",
		SessionExpiresAt: time.Now().Add(time.Hour),
	}
}

func TestSignupRepository_Producer_CreatesUserProducerPropertyAndSession(t *testing.T) {
	pool := testPool(t)
	signups := repository.NewSignupRepository(pool)
	users := repository.NewUserRepository(pool)
	ctx := context.Background()

	result, err := signups.Signup(ctx, producerSignup("Jose@Exemplo.com"))
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}

	if result.User.ID == "" || result.User.Email != "jose@exemplo.com" {
		t.Errorf("expected stored lowercase email, got %+v", result.User)
	}
	if len(result.User.Concerns) != 2 || result.User.Onboarding["areas_bucket"] != "1" {
		t.Errorf("expected concerns and onboarding saved, got %+v", result.User)
	}
	if result.Producer == nil || result.Producer.Name != "José" || result.Producer.WhatsAppPhone != "5519999999999" {
		t.Errorf("expected producer copied from the user, got %+v", result.Producer)
	}
	if result.Property == nil || result.Property.ProducerID != result.Producer.ID {
		t.Errorf("expected property linked to producer, got %+v", result.Property)
	}

	owned, err := repository.NewProducerRepository(pool).ListByUser(ctx, result.User.ID)
	if err != nil || len(owned) != 1 {
		t.Fatalf("expected the producer owned by the new user, got %v, %v", owned, err)
	}
	if _, err := users.UserBySession(ctx, "token-hash"); err != nil {
		t.Errorf("expected live session, got %v", err)
	}
}

func TestSignupRepository_Technician_CreatesOnlyUserAndSession(t *testing.T) {
	pool := testPool(t)
	signups := repository.NewSignupRepository(pool)
	ctx := context.Background()
	org := "Coopera Campo"

	result, err := signups.Signup(ctx, repository.SignupInput{
		User: model.User{
			Email: "ana@coop.com", PasswordHash: "hash", Profile: model.ProfileTechnician,
			Name: "Ana", WhatsAppPhone: "5519988887777", Organization: &org,
		},
		SessionTokenHash: "tech-token",
		SessionExpiresAt: time.Now().Add(time.Hour),
	})
	if err != nil {
		t.Fatalf("Signup: %v", err)
	}
	if result.Producer != nil || result.Property != nil {
		t.Errorf("expected no producer or property for a technician, got %+v / %+v", result.Producer, result.Property)
	}
	if result.User.Organization == nil || *result.User.Organization != org {
		t.Errorf("expected organization saved, got %v", result.User.Organization)
	}

	var producers int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM producers").Scan(&producers); err != nil {
		t.Fatalf("count producers: %v", err)
	}
	if producers != 0 {
		t.Errorf("expected 0 producers, got %d", producers)
	}
}

func TestSignupRepository_DuplicateEmail_ReturnsErrEmailTaken(t *testing.T) {
	pool := testPool(t)
	signups := repository.NewSignupRepository(pool)
	ctx := context.Background()

	if _, err := signups.Signup(ctx, producerSignup("jose@exemplo.com")); err != nil {
		t.Fatalf("first Signup: %v", err)
	}

	second := producerSignup("JOSE@exemplo.com")
	second.SessionTokenHash = "other-token"
	if _, err := signups.Signup(ctx, second); !errors.Is(err, repository.ErrEmailTaken) {
		t.Fatalf("expected ErrEmailTaken, got %v", err)
	}
}

func TestSignupRepository_FailureRollsBackEverything(t *testing.T) {
	pool := testPool(t)
	signups := repository.NewSignupRepository(pool)
	ctx := context.Background()

	input := producerSignup("rollback@exemplo.com")
	input.Property.CropStage = "not-a-stage" // violates the crop_stage enum

	if _, err := signups.Signup(ctx, input); err == nil {
		t.Fatal("expected an error for an invalid crop stage")
	}

	var users, producers int
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&users); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM producers").Scan(&producers); err != nil {
		t.Fatalf("count producers: %v", err)
	}
	if users != 0 || producers != 0 {
		t.Errorf("expected rollback, found %d users and %d producers", users, producers)
	}
}
