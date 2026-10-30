package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/app"
	"github.com/GuilhermePain/agrolang-api/internal/config"
	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/notifier"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
)

const scanConcurrency = 5

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DBDSN())
	if err != nil {
		log.Fatalf("db: connect: %v", err)
	}
	defer pool.Close()

	propertyRepo := repository.NewPropertyRepository(pool)
	alertRepo := repository.NewAlertRepository(pool)
	producerRepo := repository.NewProducerRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	signupRepo := repository.NewSignupRepository(pool)
	weatherClient := weather.NewOpenMeteoClient(weather.WithBaseURL(cfg.OpenMeteoBaseURL))
	notifierClient := notifier.NewClient(cfg.EvolutionAPIURL, cfg.EvolutionAPIKey)

	go runScanLoop(ctx, propertyRepo, weatherClient, alertRepo, producerRepo, notifierClient, cfg.ScanInterval)

	root := http.NewServeMux()
	root.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	authHandler := handler.NewAuthHandler(signupRepo, userRepo)
	authHandler.RegisterPublic(root)

	// Everything else requires a live session.
	protected := http.NewServeMux()
	authHandler.RegisterProtected(protected)
	handler.NewProducerHandler(producerRepo).Register(protected)
	handler.NewPropertyHandler(propertyRepo, alertRepo, weatherClient).Register(protected)
	handler.NewAlertHandler(alertRepo).Register(protected)
	root.Handle("/", handler.RequireSession(userRepo)(protected))

	addr := ":" + cfg.HTTPPort
	log.Printf("agrolang-api listening on %s", addr)
	if err := http.ListenAndServe(addr, handler.WithCORS(root)); err != nil {
		log.Fatal(err)
	}
}

func runScanLoop(ctx context.Context, propertyRepo *repository.PropertyRepository, weatherClient *weather.OpenMeteoClient, alertRepo *repository.AlertRepository, producerRepo *repository.ProducerRepository, notifierClient *notifier.Client, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	runOnce := func() {
		if err := app.RunCycle(ctx, propertyRepo, weatherClient, alertRepo, producerRepo, notifierClient, scanConcurrency); err != nil {
			log.Printf("app: scan cycle failed: %v", err)
		}
	}

	runOnce()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			runOnce()
		}
	}
}
