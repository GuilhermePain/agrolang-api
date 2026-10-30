package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/app"
	"github.com/GuilhermePain/agrolang-api/internal/model"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
)

func TestShouldNotify(t *testing.T) {
	cases := []struct {
		name     string
		level    risk.Level
		alert    risk.Alert
		concerns []string
		want     bool
	}{
		{"critical always notifies even when not chosen", risk.LevelCritical, risk.AlertFrost, []string{"heavy_rain"}, true},
		{"medium notifies when the type was chosen", risk.LevelMedium, risk.AlertFrost, []string{"frost", "heavy_rain"}, true},
		{"medium is skipped when the type was not chosen", risk.LevelMedium, risk.AlertHeatWave, []string{"frost"}, false},
		{"medium notifies when no concerns were chosen", risk.LevelMedium, risk.AlertHeatWave, nil, true},
		{"medium notifies for an empty concerns list", risk.LevelMedium, risk.AlertHeavyRain, []string{}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := app.ShouldNotify(tc.level, tc.alert, tc.concerns); got != tc.want {
				t.Errorf("ShouldNotify(%s, %s, %v) = %v, want %v", tc.level, tc.alert, tc.concerns, got, tc.want)
			}
		})
	}
}

func TestRunCycle_CriticalAlertNotifiesEvenWhenOwnerDidNotChooseIt(t *testing.T) {
	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	lister := &fakeLister{properties: []model.Property{
		{ID: "prop-1", ProducerID: "producer-1", Crop: "tomato", CropStage: model.CropStageHarvest},
	}}
	fetcher := &fakeFetcher{readings: []risk.WeatherReading{
		{Time: start, TempAvgC: 4, HumidityPct: 60},
		{Time: start.Add(2 * time.Hour), TempAvgC: 4, HumidityPct: 60},
	}}
	creator := &fakeAlertCreator{}
	producers := &fakeProducerFinder{producers: map[string]model.Producer{
		"producer-1": {ID: "producer-1", WhatsAppPhone: "5511999999999", Concerns: []string{"heavy_rain"}},
	}}
	notif := &fakeNotifier{}

	if err := app.RunCycle(context.Background(), lister, fetcher, creator, producers, notif, 2); err != nil {
		t.Fatalf("RunCycle: %v", err)
	}

	if len(creator.created) != 1 {
		t.Fatalf("expected the alert to be persisted, got %d", len(creator.created))
	}
	if len(notif.sent) != 1 {
		t.Fatalf("expected a critical frost alert to notify despite concerns, got %d notifications", len(notif.sent))
	}
}
