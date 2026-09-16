package risk_test

import (
	"testing"
	"time"

	"github.com/GuilhermePain/agrolang-api/internal/service/risk"
)

func TestEvaluate_SevereWaterStress_Critical(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC), TempAvgC: 36, HumidityPct: 25},
		{Time: time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC), TempAvgC: 37, HumidityPct: 20},
		{Time: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), TempAvgC: 36, HumidityPct: 22},
		{Time: time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC), TempAvgC: 38, HumidityPct: 18},
		{Time: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), TempAvgC: 36, HumidityPct: 19},
	}

	property := risk.Property{CropStage: risk.StageFlowering}

	result := risk.Evaluate(property, readings)

	if result.Level != risk.LevelCritical {
		t.Fatalf("expected LevelCritical, got %v", result.Level)
	}
	if result.Alert != risk.AlertSevereWaterStress {
		t.Fatalf("expected AlertSevereWaterStress, got %v", result.Alert)
	}
}

func TestEvaluate_MildConditions_Low(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC), TempAvgC: 28, HumidityPct: 55},
		{Time: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), TempAvgC: 27, HumidityPct: 60},
	}

	property := risk.Property{CropStage: risk.StageFlowering}

	result := risk.Evaluate(property, readings)

	if result.Level != risk.LevelLow {
		t.Fatalf("expected LevelLow, got %v", result.Level)
	}
	if result.Alert != risk.AlertNone {
		t.Fatalf("expected AlertNone, got %v", result.Alert)
	}
}

func TestEvaluate_BreachShortOf48h_NotTriggered(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC), TempAvgC: 36, HumidityPct: 25},
		{Time: time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC), TempAvgC: 37, HumidityPct: 20}, // 36h later
	}

	property := risk.Property{CropStage: risk.StageFlowering}

	result := risk.Evaluate(property, readings)

	if result.Level != risk.LevelLow {
		t.Fatalf("expected LevelLow, got %v", result.Level)
	}
}

func TestEvaluate_BreachInterrupted_ResetsStreak(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC), TempAvgC: 36, HumidityPct: 25},
		{Time: time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC), TempAvgC: 30, HumidityPct: 50}, // interrupts
		{Time: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), TempAvgC: 36, HumidityPct: 25},
	}

	property := risk.Property{CropStage: risk.StageFlowering}

	result := risk.Evaluate(property, readings)

	if result.Level != risk.LevelLow {
		t.Fatalf("expected LevelLow, got %v", result.Level)
	}
}

func TestEvaluate_NotFloweringStage_NoAlert(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC), TempAvgC: 38, HumidityPct: 15},
		{Time: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), TempAvgC: 38, HumidityPct: 15},
	}

	property := risk.Property{CropStage: risk.StageGermination}

	result := risk.Evaluate(property, readings)

	if result.Alert != risk.AlertNone {
		t.Fatalf("expected AlertNone for non-flowering stage, got %v", result.Alert)
	}
}
