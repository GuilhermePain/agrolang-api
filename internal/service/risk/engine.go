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
	AlertFrost             Alert = "frost"
	AlertHeatWave          Alert = "heat_wave"
	AlertHeavyRain         Alert = "heavy_rain"
)

type Crop string

const (
	CropCorn    Crop = "corn"
	CropTomato  Crop = "tomato"
	CropCoffee  Crop = "coffee"
	CropDefault Crop = "default"
)

// CropThresholds holds the per-crop limits that decide when a weather
// condition becomes a risk alert. Zero-valued fields fall back to
// defaultThresholds[CropDefault] via thresholdsFor.
type CropThresholds struct {
	FrostTempLimitC     float64
	HeatMediumTempC     float64
	HeatCriticalTempC   float64
	RainMediumMM24h     float64
	RainCriticalMM24h   float64
}

var defaultThresholds = map[Crop]CropThresholds{
	CropDefault: {FrostTempLimitC: 3.0, HeatMediumTempC: 38.0, HeatCriticalTempC: 40.0, RainMediumMM24h: 30.0, RainCriticalMM24h: 60.0},
	CropCorn:    {FrostTempLimitC: 3.0, HeatMediumTempC: 38.0, HeatCriticalTempC: 40.0, RainMediumMM24h: 30.0, RainCriticalMM24h: 60.0},
	CropTomato:  {FrostTempLimitC: 5.0, HeatMediumTempC: 35.0, HeatCriticalTempC: 38.0, RainMediumMM24h: 25.0, RainCriticalMM24h: 50.0},
	CropCoffee:  {FrostTempLimitC: 5.0, HeatMediumTempC: 34.0, HeatCriticalTempC: 37.0, RainMediumMM24h: 40.0, RainCriticalMM24h: 70.0},
}

func thresholdsFor(crop Crop) CropThresholds {
	if t, ok := defaultThresholds[crop]; ok {
		return t
	}
	return defaultThresholds[CropDefault]
}

type Property struct {
	Crop      Crop
	CropStage CropStage
}

type WeatherReading struct {
	Time            time.Time `json:"time"`
	TempAvgC        float64   `json:"temp_avg_c"`
	HumidityPct     float64   `json:"humidity_pct"`
	PrecipitationMM float64   `json:"precipitation_mm"`
}

type Result struct {
	Level Level `json:"level"`
	Alert Alert `json:"alert"`
}

const (
	waterStressTempThresholdC    = 35.0
	waterStressHumidityThreshold = 30.0
	waterStressWindow            = 48 * time.Hour
	heatWaveWindow               = 24 * time.Hour
	rainWindow                   = 24 * time.Hour
)

// Evaluate returns the single most severe alert triggered by the
// readings (Critical beats Medium beats Low). Prefer EvaluateAll when
// the caller must act on every concurrent risk, not just the worst one.
func Evaluate(p Property, readings []WeatherReading) Result {
	results := EvaluateAll(p, readings)

	most := results[0]
	for _, r := range results[1:] {
		if severityRank(r.Level) > severityRank(most.Level) {
			most = r
		}
	}
	return most
}

func severityRank(l Level) int {
	switch l {
	case LevelCritical:
		return 2
	case LevelMedium:
		return 1
	default:
		return 0
	}
}

// EvaluateAll returns every alert currently triggered by the readings,
// unlike Evaluate which reports only the single most severe one. A
// property can face independent risks at once (e.g. frost and heavy
// rain), and callers that persist/notify (Etapa 3/5) need all of them.
func EvaluateAll(p Property, readings []WeatherReading) []Result {
	var results []Result

	if p.CropStage == StageFlowering && severeWaterStressSustained(readings) {
		results = append(results, Result{Level: LevelCritical, Alert: AlertSevereWaterStress})
	}

	thresholds := thresholdsFor(p.Crop)

	if frostDetected(readings, thresholds.FrostTempLimitC) {
		results = append(results, Result{Level: LevelCritical, Alert: AlertFrost})
	}

	switch {
	case sustainedAbove(readings, thresholds.HeatCriticalTempC, heatWaveWindow):
		results = append(results, Result{Level: LevelCritical, Alert: AlertHeatWave})
	case sustainedAbove(readings, thresholds.HeatMediumTempC, heatWaveWindow):
		results = append(results, Result{Level: LevelMedium, Alert: AlertHeatWave})
	}

	maxRain := maxPrecipitationInWindow(readings, rainWindow)
	switch {
	case maxRain >= thresholds.RainCriticalMM24h:
		results = append(results, Result{Level: LevelCritical, Alert: AlertHeavyRain})
	case maxRain >= thresholds.RainMediumMM24h:
		results = append(results, Result{Level: LevelMedium, Alert: AlertHeavyRain})
	}

	if len(results) == 0 {
		results = append(results, Result{Level: LevelLow, Alert: AlertNone})
	}

	return results
}

func frostDetected(readings []WeatherReading, limitC float64) bool {
	for _, r := range readings {
		if r.TempAvgC < limitC {
			return true
		}
	}
	return false
}

// maxPrecipitationInWindow returns the largest rolling sum of precipitation
// found within any span of `window` duration across the readings.
func maxPrecipitationInWindow(readings []WeatherReading, window time.Duration) float64 {
	var maxSum float64
	for i := range readings {
		var sum float64
		start := readings[i].Time
		for j := i; j < len(readings) && readings[j].Time.Sub(start) < window; j++ {
			sum += readings[j].PrecipitationMM
		}
		if sum > maxSum {
			maxSum = sum
		}
	}
	return maxSum
}

func sustainedAbove(readings []WeatherReading, tempThresholdC float64, window time.Duration) bool {
	var streakStart time.Time
	inStreak := false

	for _, r := range readings {
		if r.TempAvgC > tempThresholdC {
			if !inStreak {
				streakStart = r.Time
				inStreak = true
			}
			if r.Time.Sub(streakStart) >= window {
				return true
			}
		} else {
			inStreak = false
		}
	}
	return false
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
