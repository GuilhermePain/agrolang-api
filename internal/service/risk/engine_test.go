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

func TestEvaluate_FrostBelowCriticalLimit_Critical(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 6, 1, 5, 0, 0, 0, time.UTC), TempAvgC: 2.0, HumidityPct: 80},
	}

	property := risk.Property{Crop: risk.CropCorn, CropStage: risk.StageGermination}

	result := risk.Evaluate(property, readings)

	if result.Level != risk.LevelCritical {
		t.Fatalf("expected LevelCritical, got %v", result.Level)
	}
	if result.Alert != risk.AlertFrost {
		t.Fatalf("expected AlertFrost, got %v", result.Alert)
	}
}

func TestEvaluate_TempAboveFrostLimit_NoFrostAlert(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 6, 1, 5, 0, 0, 0, time.UTC), TempAvgC: 4.0, HumidityPct: 80},
	}

	property := risk.Property{Crop: risk.CropCorn, CropStage: risk.StageGermination}

	result := risk.Evaluate(property, readings)

	if result.Alert == risk.AlertFrost {
		t.Fatalf("did not expect AlertFrost at 4.0C, got %v", result.Alert)
	}
}

func TestEvaluate_HeatWaveMediumSustained24h_Medium(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), TempAvgC: 39, HumidityPct: 50},
		{Time: time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC), TempAvgC: 39, HumidityPct: 50},
		{Time: time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC), TempAvgC: 39, HumidityPct: 50},
	}

	property := risk.Property{Crop: risk.CropCorn, CropStage: risk.StageHarvest}

	result := risk.Evaluate(property, readings)

	if result.Level != risk.LevelMedium {
		t.Fatalf("expected LevelMedium, got %v", result.Level)
	}
	if result.Alert != risk.AlertHeatWave {
		t.Fatalf("expected AlertHeatWave, got %v", result.Alert)
	}
}

func TestEvaluate_HeatWaveCriticalSustained24h_Critical(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), TempAvgC: 41, HumidityPct: 50},
		{Time: time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC), TempAvgC: 41, HumidityPct: 50},
		{Time: time.Date(2026, 1, 11, 0, 0, 0, 0, time.UTC), TempAvgC: 41, HumidityPct: 50},
	}

	property := risk.Property{Crop: risk.CropCorn, CropStage: risk.StageHarvest}

	result := risk.Evaluate(property, readings)

	if result.Level != risk.LevelCritical {
		t.Fatalf("expected LevelCritical, got %v", result.Level)
	}
	if result.Alert != risk.AlertHeatWave {
		t.Fatalf("expected AlertHeatWave, got %v", result.Alert)
	}
}

func TestEvaluate_HeatSpikeShortOf24h_NotTriggered(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), TempAvgC: 41, HumidityPct: 50},
		{Time: time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC), TempAvgC: 41, HumidityPct: 50},
	}

	property := risk.Property{Crop: risk.CropCorn, CropStage: risk.StageHarvest}

	result := risk.Evaluate(property, readings)

	if result.Alert == risk.AlertHeatWave {
		t.Fatalf("did not expect AlertHeatWave for <24h spike, got %v", result.Alert)
	}
}

func TestEvaluate_HeavyRain24hMedium_Medium(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), TempAvgC: 22, HumidityPct: 70, PrecipitationMM: 20},
		{Time: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC), TempAvgC: 22, HumidityPct: 70, PrecipitationMM: 35},
	}

	property := risk.Property{Crop: risk.CropCorn, CropStage: risk.StageHarvest}

	result := risk.Evaluate(property, readings)

	if result.Level != risk.LevelMedium {
		t.Fatalf("expected LevelMedium, got %v", result.Level)
	}
	if result.Alert != risk.AlertHeavyRain {
		t.Fatalf("expected AlertHeavyRain, got %v", result.Alert)
	}
}

func TestEvaluate_HeavyRain24hCritical_Critical(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), TempAvgC: 22, HumidityPct: 70, PrecipitationMM: 45},
		{Time: time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC), TempAvgC: 22, HumidityPct: 70, PrecipitationMM: 40},
	}

	property := risk.Property{Crop: risk.CropCorn, CropStage: risk.StageHarvest}

	result := risk.Evaluate(property, readings)

	if result.Level != risk.LevelCritical {
		t.Fatalf("expected LevelCritical, got %v", result.Level)
	}
	if result.Alert != risk.AlertHeavyRain {
		t.Fatalf("expected AlertHeavyRain, got %v", result.Alert)
	}
}

func TestEvaluate_LightRain_NoAlert(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC), TempAvgC: 22, HumidityPct: 70, PrecipitationMM: 5},
	}

	property := risk.Property{Crop: risk.CropCorn, CropStage: risk.StageHarvest}

	result := risk.Evaluate(property, readings)

	if result.Alert == risk.AlertHeavyRain {
		t.Fatalf("did not expect AlertHeavyRain for light rain, got %v", result.Alert)
	}
}

func TestEvaluateAll_FrostAndHeavyRainSimultaneously_ReturnsBothAlerts(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 6, 1, 5, 0, 0, 0, time.UTC), TempAvgC: 1.0, HumidityPct: 80, PrecipitationMM: 45},
		{Time: time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC), TempAvgC: 1.5, HumidityPct: 78, PrecipitationMM: 30},
	}

	property := risk.Property{Crop: risk.CropCorn, CropStage: risk.StageGermination}

	results := risk.EvaluateAll(property, readings)

	hasFrost := false
	hasHeavyRain := false
	for _, r := range results {
		if r.Alert == risk.AlertFrost && r.Level == risk.LevelCritical {
			hasFrost = true
		}
		if r.Alert == risk.AlertHeavyRain && r.Level == risk.LevelCritical {
			hasHeavyRain = true
		}
	}

	if !hasFrost {
		t.Errorf("expected AlertFrost in results, got %+v", results)
	}
	if !hasHeavyRain {
		t.Errorf("expected AlertHeavyRain in results, got %+v", results)
	}
}

func TestEvaluateAll_MildConditions_ReturnsOnlyLowNone(t *testing.T) {
	readings := []risk.WeatherReading{
		{Time: time.Date(2026, 8, 30, 0, 0, 0, 0, time.UTC), TempAvgC: 28, HumidityPct: 55},
	}

	property := risk.Property{Crop: risk.CropCorn, CropStage: risk.StageFlowering}

	results := risk.EvaluateAll(property, readings)

	if len(results) != 1 {
		t.Fatalf("expected exactly 1 result for mild conditions, got %+v", results)
	}
	if results[0].Level != risk.LevelLow || results[0].Alert != risk.AlertNone {
		t.Fatalf("expected LevelLow/AlertNone, got %+v", results[0])
	}
}
