package app

import (
	"context"
	"log"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/notifier"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
	"github.com/GuilhermePain/agrolang-api/internal/worker"
)

type AlertCreator interface {
	Create(ctx context.Context, a model.Alert) (model.Alert, error)
}

// AlertChecker guards against duplicate persists and notifications: if an
// unresolved alert of the same type already exists for the property, the
// current cycle tick must skip both the DB insert and the WhatsApp message.
type AlertChecker interface {
	HasActiveAlert(ctx context.Context, propertyID, alertType string) (bool, error)
}

// AlertAutoResolver allows closing alerts when the risk event is no longer detected.
type AlertAutoResolver interface {
	ResolveInactiveForProperty(ctx context.Context, propertyID string, activeAlertTypes []string) error
}

type ProducerFinder interface {
	GetByID(ctx context.Context, id string) (model.Producer, error)
}

type Notifier interface {
	Notify(ctx context.Context, phone, message string) error
}

type NotificationLogger interface {
	Create(ctx context.Context, n model.NotificationLog) (model.NotificationLog, error)
}

// RunCycle scans every property, evaluates all triggered risks, persists
// an Alert for each Medium/Critical result, and notifies the property's
// producer over WhatsApp (RF-04.1).
func RunCycle(ctx context.Context, lister worker.PropertyLister, fetcher worker.ForecastFetcher, creator AlertCreator, checker AlertChecker, producers ProducerFinder, notif Notifier, concurrency int) error {
	results, err := worker.Scan(ctx, lister, fetcher, concurrency)
	if err != nil {
		return err
	}

	for _, r := range results {
		if r.Err != nil {
			log.Printf("app: scan property %s: %v", r.Property.ID, r.Err)
			continue
		}

		defaultStart, defaultEnd := readingsRange(r.Readings)

		// Coletar tipos de alertas atualmente ativos para auto-resolução
		var activeTypes []string
		for _, res := range r.Results {
			if res.Level == risk.LevelMedium || res.Level == risk.LevelCritical {
				activeTypes = append(activeTypes, string(res.Alert))
			}
		}

		if resolver, ok := checker.(AlertAutoResolver); ok {
			if err := resolver.ResolveInactiveForProperty(ctx, r.Property.ID, activeTypes); err != nil {
				log.Printf("app: auto-resolve inactive alerts for property %s: %v", r.Property.ID, err)
			}
		}

		for _, res := range r.Results {
			if res.Level != risk.LevelMedium && res.Level != risk.LevelCritical {
				continue
			}

			active, err := checker.HasActiveAlert(ctx, r.Property.ID, string(res.Alert))
			if err != nil {
				log.Printf("app: check active alert for property %s: %v", r.Property.ID, err)
				continue
			}
			if active {
				continue
			}

			// Usa a janela exata calculada pelo motor de risco se disponível
			periodStart := res.PeriodStart
			periodEnd := res.PeriodEnd
			if periodStart.IsZero() || periodEnd.IsZero() {
				periodStart = defaultStart
				periodEnd = defaultEnd
			}

			alert := model.Alert{
				PropertyID:  r.Property.ID,
				Level:       string(res.Level),
				AlertType:   string(res.Alert),
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
			}
			persistedAlert, err := creator.Create(ctx, alert)
			if err != nil {
				log.Printf("app: persist alert for property %s: %v", r.Property.ID, err)
			}

			notifyProducer(ctx, producers, notif, r.Property, persistedAlert.ID, res, periodStart, periodEnd)
		}
	}

	return nil
}

func notifyProducer(ctx context.Context, producers ProducerFinder, notif Notifier, property model.Property, alertID string, res risk.Result, periodStart, periodEnd time.Time) {
	producer, err := producers.GetByID(ctx, property.ProducerID)
	if err != nil {
		log.Printf("app: lookup producer for property %s: %v", property.ID, err)
		return
	}

	message := notifier.BuildMessage(res, periodStart, periodEnd)
	notifyErr := notif.Notify(ctx, producer.WhatsAppPhone, message)
	if notifyErr != nil {
		log.Printf("app: notify producer %s: %v", producer.ID, notifyErr)
	}

	// Registrar log de auditoria se notif implementar NotificationLogger
	if logger, ok := notif.(NotificationLogger); ok {
		var aID *string
		if alertID != "" {
			aID = &alertID
		}
		propID := property.ID
		status := model.NotificationStatusSent
		var errMsg *string
		if notifyErr != nil {
			status = model.NotificationStatusFailed
			msg := notifyErr.Error()
			errMsg = &msg
		}

		logEntry := model.NotificationLog{
			AlertID:      aID,
			ProducerID:   producer.ID,
			PropertyID:   &propID,
			Phone:        producer.WhatsAppPhone,
			Message:      message,
			Status:       status,
			ErrorMessage: errMsg,
		}
		if _, err := logger.Create(ctx, logEntry); err != nil {
			log.Printf("app: persist notification log for producer %s: %v", producer.ID, err)
		}
	}
}

func readingsRange(readings []risk.WeatherReading) (time.Time, time.Time) {
	if len(readings) == 0 {
		return time.Time{}, time.Time{}
	}

	start, end := readings[0].Time, readings[0].Time
	for _, r := range readings[1:] {
		if r.Time.Before(start) {
			start = r.Time
		}
		if r.Time.After(end) {
			end = r.Time
		}
	}
	return start, end
}
