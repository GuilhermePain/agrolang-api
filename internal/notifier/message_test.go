package notifier_test

import (
	"strings"
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/notifier"
	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
)

func TestBuildMessage_SevereWaterStress_RecommendsIrrigationAtNight(t *testing.T) {
	start := time.Date(2026, 1, 10, 6, 0, 0, 0, time.UTC)
	end := time.Date(2026, 1, 12, 6, 0, 0, 0, time.UTC)

	msg := notifier.BuildMessage(risk.Result{Level: risk.LevelCritical, Alert: risk.AlertSevereWaterStress}, start, end)

	if !strings.Contains(msg, "irrigação") {
		t.Errorf("expected message to mention irrigation, got %q", msg)
	}
	if !strings.Contains(msg, "10/01") || !strings.Contains(msg, "12/01") {
		t.Errorf("expected message to contain the period dates, got %q", msg)
	}
}

func TestBuildMessage_Frost_RecommendsProtectingSeedlings(t *testing.T) {
	start := time.Date(2026, 6, 1, 4, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC)

	msg := notifier.BuildMessage(risk.Result{Level: risk.LevelCritical, Alert: risk.AlertFrost}, start, end)

	if !strings.Contains(msg, "geada") {
		t.Errorf("expected message to mention frost (geada), got %q", msg)
	}
	if !strings.Contains(msg, "cobertura") && !strings.Contains(msg, "colheita") {
		t.Errorf("expected message to recommend covering or harvesting, got %q", msg)
	}
}

func TestBuildMessage_HeatWave_RecommendsAvoidingFieldWorkAtPeakHeat(t *testing.T) {
	start := time.Now()
	end := start.Add(24 * time.Hour)

	msg := notifier.BuildMessage(risk.Result{Level: risk.LevelMedium, Alert: risk.AlertHeatWave}, start, end)

	if !strings.Contains(msg, "calor") {
		t.Errorf("expected message to mention heat (calor), got %q", msg)
	}
}

func TestBuildMessage_HeavyRain_RecommendsCheckingDrainage(t *testing.T) {
	start := time.Now()
	end := start.Add(24 * time.Hour)

	msg := notifier.BuildMessage(risk.Result{Level: risk.LevelMedium, Alert: risk.AlertHeavyRain}, start, end)

	if !strings.Contains(msg, "escoamento") && !strings.Contains(msg, "água") {
		t.Errorf("expected message to mention drainage/water, got %q", msg)
	}
}

func TestBuildMessage_NeverContainsTechnicalJargon(t *testing.T) {
	cases := []risk.Alert{risk.AlertSevereWaterStress, risk.AlertFrost, risk.AlertHeatWave, risk.AlertHeavyRain}
	jargon := []string{"evapotranspiração", "hPa"}

	for _, alertType := range cases {
		msg := notifier.BuildMessage(risk.Result{Level: risk.LevelCritical, Alert: alertType}, time.Now(), time.Now())
		for _, term := range jargon {
			if strings.Contains(msg, term) {
				t.Errorf("alert %v: message contains technical jargon %q: %q", alertType, term, msg)
			}
		}
	}
}
