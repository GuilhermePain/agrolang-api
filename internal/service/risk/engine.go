package risk

import (
	"strings"
	"time"
)

type CropStage string

const (
	StageGermination CropStage = "germination"
	StageFlowering   CropStage = "flowering"
	StageHarvest     CropStage = "harvest"
)

type SoilType string

const (
	SoilSandy     SoilType = "arenoso"
	SoilClay      SoilType = "argiloso"
	SoilMedium    SoilType = "medio"
	SoilShallow   SoilType = "raso"       // Neossolos litólicos / Vertissolos do semiárido
	SoilAlluvial  SoilType = "aluvial"    // Várzeas e solos aluviais
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
	CropCorn      Crop = "milho"
	CropCornEn    Crop = "corn"
	CropTomato    Crop = "tomate"
	CropTomatoEn  Crop = "tomato"
	CropCoffee    Crop = "cafe"
	CropCoffeeEn  Crop = "coffee"
	CropBean      Crop = "feijao"        // Feijão comum e feijão-caupi (Nordeste/Ceará)
	CropCassava   Crop = "mandioca"      // Mandioca / Macaxeira
	CropCashew    Crop = "caju"          // Cajueiro (símbolo do Ceará)
	CropSoy       Crop = "soja"          // Soja
	CropCotton    Crop = "algodao"       // Algodão
	CropFruit     Crop = "fruticultura"  // Melão, Melancia, Manga irrigada (Vale do Jaguaribe/CE)
	CropDefault   Crop = "default"
)

type CropThresholds struct {
	FrostTempLimitC   float64
	HeatMediumTempC   float64
	HeatCriticalTempC float64
	RainMediumMM24h   float64
	RainCriticalMM24h float64
}

var defaultThresholds = map[Crop]CropThresholds{
	CropDefault:     {FrostTempLimitC: 3.0, HeatMediumTempC: 38.0, HeatCriticalTempC: 40.0, RainMediumMM24h: 30.0, RainCriticalMM24h: 60.0},
	CropCorn:        {FrostTempLimitC: 3.0, HeatMediumTempC: 38.0, HeatCriticalTempC: 40.0, RainMediumMM24h: 30.0, RainCriticalMM24h: 60.0},
	CropCornEn:      {FrostTempLimitC: 3.0, HeatMediumTempC: 38.0, HeatCriticalTempC: 40.0, RainMediumMM24h: 30.0, RainCriticalMM24h: 60.0},
	CropTomato:      {FrostTempLimitC: 5.0, HeatMediumTempC: 35.0, HeatCriticalTempC: 38.0, RainMediumMM24h: 25.0, RainCriticalMM24h: 50.0},
	CropTomatoEn:    {FrostTempLimitC: 5.0, HeatMediumTempC: 35.0, HeatCriticalTempC: 38.0, RainMediumMM24h: 25.0, RainCriticalMM24h: 50.0},
	CropCoffee:      {FrostTempLimitC: 5.0, HeatMediumTempC: 34.0, HeatCriticalTempC: 37.0, RainMediumMM24h: 40.0, RainCriticalMM24h: 70.0},
	CropCoffeeEn:    {FrostTempLimitC: 5.0, HeatMediumTempC: 34.0, HeatCriticalTempC: 37.0, RainMediumMM24h: 40.0, RainCriticalMM24h: 70.0},
	CropBean:        {FrostTempLimitC: 4.0, HeatMediumTempC: 34.0, HeatCriticalTempC: 37.0, RainMediumMM24h: 28.0, RainCriticalMM24h: 55.0},
	CropCassava:     {FrostTempLimitC: 3.0, HeatMediumTempC: 38.0, HeatCriticalTempC: 41.0, RainMediumMM24h: 45.0, RainCriticalMM24h: 80.0},
	CropCashew:      {FrostTempLimitC: 8.0, HeatMediumTempC: 39.0, HeatCriticalTempC: 42.0, RainMediumMM24h: 50.0, RainCriticalMM24h: 90.0},
	CropSoy:         {FrostTempLimitC: 3.0, HeatMediumTempC: 36.0, HeatCriticalTempC: 39.0, RainMediumMM24h: 35.0, RainCriticalMM24h: 70.0},
	CropCotton:      {FrostTempLimitC: 5.0, HeatMediumTempC: 38.0, HeatCriticalTempC: 41.0, RainMediumMM24h: 30.0, RainCriticalMM24h: 60.0},
	CropFruit:       {FrostTempLimitC: 6.0, HeatMediumTempC: 38.0, HeatCriticalTempC: 41.0, RainMediumMM24h: 25.0, RainCriticalMM24h: 50.0},
}

// accentFolder strips the pt-BR diacritics so "Café" and "Feijão" match the
// ASCII crop constants ("cafe", "feijao").
var accentFolder = strings.NewReplacer(
	"á", "a", "à", "a", "â", "a", "ã", "a", "ä", "a",
	"é", "e", "è", "e", "ê", "e", "ë", "e",
	"í", "i", "ì", "i", "î", "i", "ï", "i",
	"ó", "o", "ò", "o", "ô", "o", "õ", "o", "ö", "o",
	"ú", "u", "ù", "u", "û", "u", "ü", "u",
	"ç", "c",
)

func normalizeCrop(crop Crop) Crop {
	return Crop(accentFolder.Replace(strings.ToLower(strings.TrimSpace(string(crop)))))
}

func thresholdsFor(crop Crop) CropThresholds {
	norm := normalizeCrop(crop)
	if t, ok := defaultThresholds[norm]; ok {
		return t
	}
	return defaultThresholds[CropDefault]
}

type Property struct {
	Crop      Crop
	CropStage CropStage
	SoilType  SoilType
}

type WeatherReading struct {
	Time            time.Time `json:"time"`
	TempAvgC        float64   `json:"temp_avg_c"`
	HumidityPct     float64   `json:"humidity_pct"`
	PrecipitationMM float64   `json:"precipitation_mm"`
}

type Result struct {
	Level       Level     `json:"level"`
	Alert       Alert     `json:"alert"`
	PeriodStart time.Time `json:"period_start,omitempty"`
	PeriodEnd   time.Time `json:"period_end,omitempty"`
}

const (
	waterStressTempThresholdC    = 35.0
	waterStressHumidityThreshold = 30.0
	heatWaveWindow               = 24 * time.Hour
	rainWindow                   = 24 * time.Hour
)

// soilWaterStressWindow adjusts drought vulnerability according to soil water holding capacity.
func soilWaterStressWindow(soil SoilType) time.Duration {
	switch strings.ToLower(string(soil)) {
	case "arenoso", "sandy":
		return 32 * time.Hour // Menor retenção hídrica: seca muito rápido
	case "raso", "pedregoso", "vertissolo", "litolico":
		return 28 * time.Hour // Solo raso do semiárido: pouca reserva útil de água
	case "argiloso", "clay":
		return 60 * time.Hour // Alta retenção hídrica: resiste por mais tempo
	case "aluvial", "silte":
		return 54 * time.Hour // Solo profundo e fértil de baixio/várzea
	default:
		return 48 * time.Hour // Solo médio ou padrão FAO-56
	}
}

// Evaluate returns the single most severe alert triggered by the readings.
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

// EvaluateAll returns every alert currently triggered by the readings with its exact incident window.
func EvaluateAll(p Property, readings []WeatherReading) []Result {
	var results []Result

	// 1. Estresse hídrico severo na floração (ajustado por textura do solo)
	stressWindow := soilWaterStressWindow(p.SoilType)
	if p.CropStage == StageFlowering {
		if ok, start, end := detectSevereWaterStress(readings, stressWindow); ok {
			results = append(results, Result{
				Level:       LevelCritical,
				Alert:       AlertSevereWaterStress,
				PeriodStart: start,
				PeriodEnd:   end,
			})
		}
	}

	thresholds := thresholdsFor(p.Crop)

	// 2. Geada
	if ok, start, end := detectFrost(readings, thresholds.FrostTempLimitC); ok {
		results = append(results, Result{
			Level:       LevelCritical,
			Alert:       AlertFrost,
			PeriodStart: start,
			PeriodEnd:   end,
		})
	}

	// 3. Onda de calor
	if ok, start, end := detectSustainedHeat(readings, thresholds.HeatCriticalTempC, heatWaveWindow); ok {
		results = append(results, Result{
			Level:       LevelCritical,
			Alert:       AlertHeatWave,
			PeriodStart: start,
			PeriodEnd:   end,
		})
	} else if ok, start, end := detectSustainedHeat(readings, thresholds.HeatMediumTempC, heatWaveWindow); ok {
		results = append(results, Result{
			Level:       LevelMedium,
			Alert:       AlertHeatWave,
			PeriodStart: start,
			PeriodEnd:   end,
		})
	}

	// 4. Chuva forte
	if ok, start, end := detectHeavyRain(readings, thresholds.RainCriticalMM24h, rainWindow); ok {
		results = append(results, Result{
			Level:       LevelCritical,
			Alert:       AlertHeavyRain,
			PeriodStart: start,
			PeriodEnd:   end,
		})
	} else if ok, start, end := detectHeavyRain(readings, thresholds.RainMediumMM24h, rainWindow); ok {
		results = append(results, Result{
			Level:       LevelMedium,
			Alert:       AlertHeavyRain,
			PeriodStart: start,
			PeriodEnd:   end,
		})
	}

	if len(results) == 0 {
		start, end := time.Time{}, time.Time{}
		if len(readings) > 0 {
			start = readings[0].Time
			end = readings[len(readings)-1].Time
		}
		results = append(results, Result{
			Level:       LevelLow,
			Alert:       AlertNone,
			PeriodStart: start,
			PeriodEnd:   end,
		})
	}

	return results
}

func detectFrost(readings []WeatherReading, limitC float64) (bool, time.Time, time.Time) {
	for i, r := range readings {
		if r.TempAvgC < limitC {
			start := r.Time
			end := r.Time.Add(6 * time.Hour)
			if i+1 < len(readings) && readings[i+1].Time.After(end) {
				end = readings[i+1].Time
			}
			return true, start, end
		}
	}
	return false, time.Time{}, time.Time{}
}

func detectHeavyRain(readings []WeatherReading, thresholdMM float64, window time.Duration) (bool, time.Time, time.Time) {
	for i := range readings {
		var sum float64
		start := readings[i].Time
		var lastTime time.Time
		for j := i; j < len(readings) && readings[j].Time.Sub(start) <= window; j++ {
			sum += readings[j].PrecipitationMM
			lastTime = readings[j].Time
			if sum >= thresholdMM {
				return true, start, lastTime
			}
		}
	}
	return false, time.Time{}, time.Time{}
}

func detectSustainedHeat(readings []WeatherReading, tempThresholdC float64, window time.Duration) (bool, time.Time, time.Time) {
	var streakStart time.Time
	inStreak := false

	for _, r := range readings {
		if r.TempAvgC > tempThresholdC {
			if !inStreak {
				streakStart = r.Time
				inStreak = true
			}
			if r.Time.Sub(streakStart) >= window {
				return true, streakStart, r.Time
			}
		} else {
			inStreak = false
		}
	}
	return false, time.Time{}, time.Time{}
}

func detectSevereWaterStress(readings []WeatherReading, window time.Duration) (bool, time.Time, time.Time) {
	if len(readings) == 0 {
		return false, time.Time{}, time.Time{}
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
			if r.Time.Sub(streakStart) >= window {
				return true, streakStart, r.Time
			}
		} else {
			inStreak = false
		}
	}
	return false, time.Time{}, time.Time{}
}
