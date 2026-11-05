import type { WeatherReading } from "./api";

export type DailyForecast = {
  date: string;
  minTempC: number;
  maxTempC: number;
  totalPrecipitationMM: number;
  avgCloudCoverPct: number;
  maxRainProbPct: number;
  maxWindSpeedKmh: number;
  weatherCode: number;
  conditionText: string;
  simplifiedSummary: string;
  farmingAdvice: string;
};

export function getWeatherCondition(code: number, cloudPct: number, rainMm: number): {
  text: string;
  simplified: string;
} {
  if (rainMm >= 25 || code === 65 || code === 82 || code === 95 || code === 96 || code === 99) {
    return {
      text: "Chuva Forte / Tempestade",
      simplified: "Chuva intensa prevista. Evite trabalhos a céu aberto e proteja áreas baixas.",
    };
  }
  if (rainMm > 5 || code === 63 || code === 81) {
    return {
      text: "Chuva Moderada",
      simplified: "Chuva boa para reposição de umidade. Solo receberá água natural.",
    };
  }
  if (rainMm > 0.5 || code === 51 || code === 53 || code === 61 || code === 80) {
    return {
      text: "Pancadas de Chuva / Chuvisco",
      simplified: "Pancadas rápidas ou chuvisco passageiro. Não descarte a irrigação se o solo secar.",
    };
  }
  if (cloudPct >= 75 || code === 3) {
    return {
      text: "Nublado / Encoberto",
      simplified: "Céu bastante nublado. Menor perda de água por evaporação.",
    };
  }
  if (cloudPct >= 35 || code === 1 || code === 2) {
    return {
      text: "Sol entre Nuvens",
      simplified: "Tempo agradável com sol entre nuvens. Bom para manejo e vistorias.",
    };
  }
  return {
    text: "Ensolarado / Céu Limpo",
    simplified: "Sol forte e céu limpo o dia todo. Evaporação alta, regar nos horários mais frescos.",
  };
}

export function getFarmingRecommendation(
  rainMm: number,
  maxTemp: number,
  cloudPct: number,
  windSpeed: number
): string {
  if (rainMm > 15) {
    return "Não aplicar defensivos nem fertilizantes foliares (risco de lavagem pela chuva). Verificar drenagem.";
  }
  if (windSpeed > 20) {
    return "Ventos fortes (> 20 km/h). Evite pulverizações para prevenir deriva do produto químico.";
  }
  if (maxTemp >= 35 && rainMm === 0) {
    return "Calor intenso e insolação direta. Priorize irrigação noturna ou bem cedo para evitar choque térmico.";
  }
  if (cloudPct > 60 && rainMm < 2) {
    return "Tempo encoberto e ameno. Excelente janela para plantio, transplante de mudas e adubação de solo.";
  }
  return "Condições climáticas estáveis e favoráveis para as atividades normais de campo.";
}

/** Groups hourly readings into per-day statistics and farmer-friendly advice. */
export function groupByDay(readings: WeatherReading[]): DailyForecast[] {
  type Accumulator = {
    date: string;
    minTempC: number;
    maxTempC: number;
    totalPrecipitationMM: number;
    cloudSum: number;
    cloudCount: number;
    maxRainProb: number;
    maxWind: number;
    codes: number[];
  };

  const byDate = new Map<string, Accumulator>();

  for (const r of readings) {
    const date = r.time.slice(0, 10);
    const existing = byDate.get(date);
    const cloud = r.cloud_cover_pct ?? 25;
    const rainProb = r.rain_prob_pct ?? (r.precipitation_mm > 0 ? 80 : 10);
    const wind = r.wind_speed_kmh ?? 12;
    const code = r.weather_code ?? 1;

    if (!existing) {
      byDate.set(date, {
        date,
        minTempC: r.temp_avg_c,
        maxTempC: r.temp_avg_c,
        totalPrecipitationMM: r.precipitation_mm,
        cloudSum: cloud,
        cloudCount: 1,
        maxRainProb: rainProb,
        maxWind: wind,
        codes: [code],
      });
      continue;
    }

    existing.minTempC = Math.min(existing.minTempC, r.temp_avg_c);
    existing.maxTempC = Math.max(existing.maxTempC, r.temp_avg_c);
    existing.totalPrecipitationMM += r.precipitation_mm;
    existing.cloudSum += cloud;
    existing.cloudCount += 1;
    existing.maxRainProb = Math.max(existing.maxRainProb, rainProb);
    existing.maxWind = Math.max(existing.maxWind, wind);
    existing.codes.push(code);
  }

  return Array.from(byDate.values()).map((item) => {
    const avgCloud = Math.round(item.cloudSum / Math.max(1, item.cloudCount));
    // Determine representative weather code (prefer higher severity rain codes if present)
    const worstCode = item.codes.reduce((prev, curr) => (curr > prev ? curr : prev), item.codes[0] || 0);
    const condition = getWeatherCondition(worstCode, avgCloud, item.totalPrecipitationMM);
    const advice = getFarmingRecommendation(
      item.totalPrecipitationMM,
      item.maxTempC,
      avgCloud,
      item.maxWind
    );

    return {
      date: item.date,
      minTempC: item.minTempC,
      maxTempC: item.maxTempC,
      totalPrecipitationMM: item.totalPrecipitationMM,
      avgCloudCoverPct: avgCloud,
      maxRainProbPct: item.maxRainProb,
      maxWindSpeedKmh: item.maxWind,
      weatherCode: worstCode,
      conditionText: condition.text,
      simplifiedSummary: condition.simplified,
      farmingAdvice: advice,
    };
  });
}
