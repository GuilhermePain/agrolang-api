package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/app"
	"github.com/GuilhermePain/agrolang-api/internal/model"
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

func TestRunCycle_PersistsOnlyMediumAndCriticalAlerts(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(2 * time.Hour)
	lister := &fakeLister{properties: []model.Property{
		{ID: "prop-1", Crop: "tomato", CropStage: model.CropStageHarvest},
	}}
	fetcher := &fakeFetcher{readings: []risk.WeatherReading{
		{Time: start, TempAvgC: 4, HumidityPct: 60},
		{Time: end, TempAvgC: 4, HumidityPct: 60},
	}}
	creator := &fakeAlertCreator{}

	err := app.RunCycle(context.Background(), lister, fetcher, creator, 2)
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
}

func TestRunCycle_DoesNotPersistLowRisk(t *testing.T) {
	lister := &fakeLister{properties: []model.Property{
		{ID: "prop-1", Crop: "tomato", CropStage: model.CropStageHarvest},
	}}
	fetcher := &fakeFetcher{readings: []risk.WeatherReading{
		{Time: time.Now(), TempAvgC: 20, HumidityPct: 60},
	}}
	creator := &fakeAlertCreator{}

	err := app.RunCycle(context.Background(), lister, fetcher, creator, 2)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if len(creator.created) != 0 {
		t.Fatalf("expected no persisted alerts, got %d", len(creator.created))
	}
}

func TestRunCycle_PropertyFetchError_DoesNotAbortCycle(t *testing.T) {
	lister := &fakeLister{properties: []model.Property{
		{ID: "prop-1", Crop: "tomato", CropStage: model.CropStageHarvest},
		{ID: "prop-2", Crop: "tomato", CropStage: model.CropStageHarvest},
	}}
	fetcher := &fakeFetcher{err: context.DeadlineExceeded}
	creator := &fakeAlertCreator{}

	err := app.RunCycle(context.Background(), lister, fetcher, creator, 2)
	if err != nil {
		t.Fatalf("RunCycle: %v", err)
	}
	if len(creator.created) != 0 {
		t.Fatalf("expected no persisted alerts, got %d", len(creator.created))
	}
}
