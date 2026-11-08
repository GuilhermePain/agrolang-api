package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/GuilhermePain/agrolang-api/internal/app"
	"github.com/GuilhermePain/agrolang-api/internal/config"
	"github.com/GuilhermePain/agrolang-api/internal/handler"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/notifier"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
)

const scanConcurrency = 5

type loggingNotifier struct {
	client *notifier.Client
	repo   *repository.NotificationRepository
}

func (l *loggingNotifier) Notify(ctx context.Context, phone, message string) error {
	return l.client.Notify(ctx, phone, message)
}

func (l *loggingNotifier) Create(ctx context.Context, n model.NotificationLog) (model.NotificationLog, error) {
	return l.repo.Create(ctx, n)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

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
	notificationRepo := repository.NewNotificationRepository(pool)

	weatherClient := weather.NewOpenMeteoClient(weather.WithBaseURL(cfg.OpenMeteoBaseURL))
	notifierClient := notifier.NewClient(cfg.EvolutionAPIURL, cfg.EvolutionAPIKey)
	activeNotifier := &loggingNotifier{client: notifierClient, repo: notificationRepo}

	go runScanLoop(ctx, propertyRepo, weatherClient, alertRepo, producerRepo, activeNotifier, cfg.ScanInterval)

	// Public routes only: health, signup/login and the WhatsApp provider
	// callback. Everything else is mounted on `protected` below.
	root := http.NewServeMux()
	root.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	authHandler := handler.NewAuthHandler(signupRepo, userRepo)
	authHandler.RegisterPublic(root)
	// Evolution (WhatsApp) delivery callbacks carry no session: public, and
	// without origin verification for now (see WebhookHandler).
	handler.NewWebhookHandler(notificationRepo).RegisterPublic(root)

	// Everything else requires a live session and is scoped to its user.
	protected := http.NewServeMux()
	authHandler.RegisterProtected(protected)

	// On-demand scan (RF-02, RF-03): any logged-in user may trigger it, but
	// the scan itself stays global, like the background loop.
	handler.NewScanHandler(func(ctx context.Context) error {
		return app.RunCycle(ctx, propertyRepo, weatherClient, alertRepo, alertRepo, producerRepo, activeNotifier, scanConcurrency)
	}).Register(protected)

	handler.NewProducerHandler(producerRepo).Register(protected)
	handler.NewPropertyHandler(propertyRepo, alertRepo, weatherClient).Register(protected)
	handler.NewAlertHandler(alertRepo).Register(protected)
	handler.NewNotificationHandler(notificationRepo).Register(protected)
	root.Handle("/", handler.RequireSession(userRepo)(protected))

	addr := ":" + cfg.HTTPPort
	server := &http.Server{
		Addr:    addr,
		Handler: handler.WithCORS(root),
	}

	go func() {
		log.Printf("agrolang-api listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down agrolang-api gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
	}
	log.Println("agrolang-api stopped cleanly")
}

func runScanLoop(ctx context.Context, propertyRepo *repository.PropertyRepository, weatherClient *weather.OpenMeteoClient, alertRepo *repository.AlertRepository, producerRepo *repository.ProducerRepository, activeNotifier *loggingNotifier, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	runOnce := func() {
		if err := app.RunCycle(ctx, propertyRepo, weatherClient, alertRepo, alertRepo, producerRepo, activeNotifier, scanConcurrency); err != nil {
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
