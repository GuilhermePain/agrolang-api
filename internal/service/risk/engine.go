package risk

import "time"

type CropStage string

const (
	StageGermination CropStage = "germination"
	StageFlowering   CropStage = "flowering"
	StageHarvest     CropStage = "harvest"
)

type Level string

const (
	LevelLow      Level = "low"
	LevelMedium   Level = "medium"
	LevelCritical Level = "critical"
)

type Alert string

const (
	AlertNone              Alert = "none"
	AlertSevereWaterStress Alert = "severe_water_stress"
)

type Property struct {
	CropStage CropStage
}

type WeatherReading struct {
	Time        time.Time
	TempAvgC    float64
	HumidityPct float64
}

type Result struct {
	Level Level
	Alert Alert
}

const (
	waterStressTempThresholdC    = 35.0
	waterStressHumidityThreshold = 30.0
	waterStressWindow            = 48 * time.Hour
)

func Evaluate(p Property, readings []WeatherReading) Result {
	if p.CropStage == StageFlowering && severeWaterStressSustained(readings) {
		return Result{Level: LevelCritical, Alert: AlertSevereWaterStress}
	}
	return Result{Level: LevelLow, Alert: AlertNone}
}

func severeWaterStressSustained(readings []WeatherReading) bool {
	if len(readings) == 0 {
		return false
	}

	var streakStart time.Time
	inStreak := false

	for _, r := range readings {
		breach := r.TempAvgC > waterStressTempThresholdC && r.HumidityPct < waterStressHumidityThreshold
		if breach {
			if !inStreak {
				streakStart = r.Time
				inStreak = true
			}
			if r.Time.Sub(streakStart) >= waterStressWindow {
				return true
			}
		} else {
			inStreak = false
		}
	}
	return false
}
