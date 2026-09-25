package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/app"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/repository"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
	"github.com/GuilhermePain/agrolang-api/internal/weather"
)

type fakeLister struct {
	properties []model.Property
}

func (f *fakeLister) ListAll(ctx context.Context) ([]model.Property, error) {
	return f.properties, nil
}

type fakeFetcher struct {
	readings []risk.WeatherReading
	err      error
}

func (f *fakeFetcher) FetchForecast(ctx context.Context, coords weather.Coordinates) ([]risk.WeatherReading, error) {
	return f.readings, f.err
}

type createdAlert struct {
	alert model.Alert
}

type fakeAlertCreator struct {
	created []createdAlert
	err     error
}

func (f *fakeAlertCreator) Create(ctx context.Context, a model.Alert) (model.Alert, error) {
	f.created = append(f.created, createdAlert{alert: a})
	return a, f.err
}

type fakeProducerFinder struct {
	producers map[string]model.Producer
}

func (f *fakeProducerFinder) GetByID(ctx context.Context, id string) (model.Producer, error) {
	p, ok := f.producers[id]
	if !ok {
		return model.Producer{}, repository.ErrNotFound
	}
	return p, nil
}

type sentNotification struct {
	phone   string
	message string
}

type fakeNotifier struct {
	sent []sentNotification
	err  error
}

func (f *fakeNotifier) Notify(ctx context.Context, phone, message string) error {
	f.sent = append(f.sent, sentNotification{phone: phone, message: message})
	return f.err
}

func TestRunCycle_PersistsOnlyMediumAndCriticalAlerts(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	lister := &fakeLister{properties: []model.Property{
		{ID: "prop-1", ProducerID: "producer-1", Crop: "tomato", CropStage: model.CropStageHarvest},
	}}
	fetcher := &fakeFetcher{readings: []risk.WeatherReading{
		{Time: start, TempAvgC: 4, HumidityPct: 60},
		{Time: end, TempAvgC: 4, HumidityPct: 60},
	}}
	creator := &fakeAlertCreator{}
	producers := &fakeProducerFinder{producers: map[string]model.Producer{
		"producer-1": {ID: "producer-1", WhatsAppPhone: "5511999999999"},
	}}
	notif := &fakeNotifier{}

	err := app.RunCycle(context.Background(), lister, fetcher, creator, producers, notif, 2)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if len(creator.created) != 1 {
		t.Fatalf("expected 1 persisted alert, got %d", len(creator.created))
	}
	got := creator.created[0].alert
	if got.PropertyID != "prop-1" {
		t.Errorf("expected PropertyID prop-1, got %q", got.PropertyID)
	}
	if got.AlertType != string(risk.AlertFrost) {
		t.Errorf("expected alert_type frost, got %q", got.AlertType)
	}
	if got.Level != string(risk.LevelCritical) {
		t.Errorf("expected level critical, got %q", got.Level)
	}
	if !got.PeriodStart.Equal(start) || !got.PeriodEnd.Equal(end) {
		t.Errorf("expected period %v-%v, got %v-%v", start, end, got.PeriodStart, got.PeriodEnd)
	}

	if len(notif.sent) != 1 {
		t.Fatalf("expected 1 notification sent, got %d", len(notif.sent))
	}
	if notif.sent[0].phone != "5511999999999" {
		t.Errorf("expected notification to producer phone, got %q", notif.sent[0].phone)
	}
	if notif.sent[0].message == "" {
		t.Errorf("expected non-empty notification message")
	}
}

func TestRunCycle_DoesNotPersistOrNotifyLowRisk(t *testing.T) {
	lister := &fakeLister{properties: []model.Property{
		{ID: "prop-1", ProducerID: "producer-1", Crop: "tomato", CropStage: model.CropStageHarvest},
	}}
	fetcher := &fakeFetcher{readings: []risk.WeatherReading{
		{Time: time.Now(), TempAvgC: 20, HumidityPct: 60},
	}}
	creator := &fakeAlertCreator{}
	producers := &fakeProducerFinder{producers: map[string]model.Producer{
		"producer-1": {ID: "producer-1", WhatsAppPhone: "5511999999999"},
	}}
	notif := &fakeNotifier{}

	err := app.RunCycle(context.Background(), lister, fetcher, creator, producers, notif, 2)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if len(creator.created) != 0 {
		t.Fatalf("expected no persisted alerts, got %d", len(creator.created))
	}
	if len(notif.sent) != 0 {
		t.Fatalf("expected no notifications sent, got %d", len(notif.sent))
	}
}

func TestRunCycle_PropertyFetchError_DoesNotAbortCycle(t *testing.T) {
	lister := &fakeLister{properties: []model.Property{
		{ID: "prop-1", ProducerID: "producer-1", Crop: "tomato", CropStage: model.CropStageHarvest},
		{ID: "prop-2", ProducerID: "producer-1", Crop: "tomato", CropStage: model.CropStageHarvest},
	}}
	fetcher := &fakeFetcher{err: context.DeadlineExceeded}
	creator := &fakeAlertCreator{}
	producers := &fakeProducerFinder{producers: map[string]model.Producer{
		"producer-1": {ID: "producer-1", WhatsAppPhone: "5511999999999"},
	}}
	notif := &fakeNotifier{}

	err := app.RunCycle(context.Background(), lister, fetcher, creator, producers, notif, 2)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}
	if len(creator.created) != 0 {
		t.Fatalf("expected no persisted alerts, got %d", len(creator.created))
	}
	if len(notif.sent) != 0 {
		t.Fatalf("expected no notifications sent, got %d", len(notif.sent))
	}
}

func TestRunCycle_NotifyError_DoesNotAbortCycle(t *testing.T) {
	lister := &fakeLister{properties: []model.Property{
		{ID: "prop-1", ProducerID: "producer-1", Crop: "tomato", CropStage: model.CropStageHarvest},
	}}
	fetcher := &fakeFetcher{readings: []risk.WeatherReading{
		{Time: time.Now(), TempAvgC: 2, HumidityPct: 60},
	}}
	creator := &fakeAlertCreator{}
	producers := &fakeProducerFinder{producers: map[string]model.Producer{
		"producer-1": {ID: "producer-1", WhatsAppPhone: "5511999999999"},
	}}
	notif := &fakeNotifier{err: context.DeadlineExceeded}

	err := app.RunCycle(context.Background(), lister, fetcher, creator, producers, notif, 2)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}
	if len(creator.created) != 1 {
		t.Fatalf("expected alert still persisted despite notify error, got %d", len(creator.created))
	}
}
