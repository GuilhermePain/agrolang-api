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

type ProducerFinder interface {
	GetByID(ctx context.Context, id string) (model.Producer, error)
}

type Notifier interface {
	Notify(ctx context.Context, phone, message string) error
}

// RunCycle scans every property, evaluates all triggered risks, persists
// an Alert for each Medium/Critical result, and notifies the property's
// producer over WhatsApp (RF-04.1). A per-property fetch error, a failed
// persist, or a failed notification is logged and skipped so one bad
// property never aborts the rest of the cycle.
func RunCycle(ctx context.Context, lister worker.PropertyLister, fetcher worker.ForecastFetcher, creator AlertCreator, producers ProducerFinder, notif Notifier, concurrency int) error {
	results, err := worker.Scan(ctx, lister, fetcher, concurrency)
	if err != nil {
		return err
	}

	for _, r := range results {
		if r.Err != nil {
			log.Printf("app: scan property %s: %v", r.Property.ID, r.Err)
			continue
		}

		periodStart, periodEnd := readingsRange(r.Readings)

		for _, res := range r.Results {
			if res.Level != risk.LevelMedium && res.Level != risk.LevelCritical {
				continue
			}

			alert := model.Alert{
				PropertyID:  r.Property.ID,
				Level:       string(res.Level),
				AlertType:   string(res.Alert),
				PeriodStart: periodStart,
				PeriodEnd:   periodEnd,
			}
			if _, err := creator.Create(ctx, alert); err != nil {
				log.Printf("app: persist alert for property %s: %v", r.Property.ID, err)
			}

			notifyProducer(ctx, producers, notif, r.Property, res, periodStart, periodEnd)
		}
	}

	return nil
}

func notifyProducer(ctx context.Context, producers ProducerFinder, notif Notifier, property model.Property, res risk.Result, periodStart, periodEnd time.Time) {
	producer, err := producers.GetByID(ctx, property.ProducerID)
	if err != nil {
		log.Printf("app: lookup producer for property %s: %v", property.ID, err)
		return
	}

	message := notifier.BuildMessage(res, periodStart, periodEnd)
	if err := notif.Notify(ctx, producer.WhatsAppPhone, message); err != nil {
		log.Printf("app: notify producer %s: %v", producer.ID, err)
	}
}

// readingsRange returns the earliest and latest reading timestamps.
// The engine doesn't expose which sub-window triggered a given alert,
// so the full fetched range is used as the alert's period.
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
