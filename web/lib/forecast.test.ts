import { describe, expect, it } from "vitest";
import { getFarmingRecommendation, getWeatherCondition, groupByDay } from "./forecast";
import type { WeatherReading } from "./types";

function reading(time: string, tempAvgC: number, precipitationMM = 0): WeatherReading {
  return { time, temp_avg_c: tempAvgC, humidity_pct: 50, precipitation_mm: precipitationMM };
}

describe("groupByDay", () => {
  it("returns an empty list for no readings", () => {
    expect(groupByDay([])).toEqual([]);
  });

  it("groups a single day's readings into one entry with min/max temp and total precipitation", () => {
    const readings = [
      reading("2026-01-01T00:00", 10, 1),
      reading("2026-01-01T12:00", 30, 2),
      reading("2026-01-01T23:00", 20, 0.5),
    ];

    expect(groupByDay(readings)).toMatchObject([
      { date: "2026-01-01", minTempC: 10, maxTempC: 30, totalPrecipitationMM: 3.5 },
    ]);
  });

  it("splits readings into separate entries per day, in first-seen order", () => {
    const readings = [
      reading("2026-01-02T00:00", 15),
      reading("2026-01-01T00:00", 10),
      reading("2026-01-02T12:00", 25),
    ];

    expect(groupByDay(readings)).toMatchObject([
      { date: "2026-01-02", minTempC: 15, maxTempC: 25, totalPrecipitationMM: 0 },
      { date: "2026-01-01", minTempC: 10, maxTempC: 10, totalPrecipitationMM: 0 },
    ]);
  });

  it("derives cloud, rain chance, wind and the worst weather code per day", () => {
    const readings: WeatherReading[] = [
      { ...reading("2026-01-01T00:00", 20), cloud_cover_pct: 80, rain_prob_pct: 30, wind_speed_kmh: 5, weather_code: 3 },
      { ...reading("2026-01-01T12:00", 22), cloud_cover_pct: 90, rain_prob_pct: 60, wind_speed_kmh: 9, weather_code: 61 },
    ];

    expect(groupByDay(readings)).toMatchObject([
      { avgCloudCoverPct: 85, maxRainProbPct: 60, maxWindSpeedKmh: 9, weatherCode: 61 },
    ]);
  });

  it("falls back to defaults when the API omits the optional fields", () => {
    const [day] = groupByDay([reading("2026-01-01T00:00", 20, 1)]);
    expect(day).toMatchObject({ avgCloudCoverPct: 25, maxRainProbPct: 80, maxWindSpeedKmh: 12, weatherCode: 1 });
  });
});

describe("getWeatherCondition", () => {
  it("ranks heavy rain above clouds and sun", () => {
    expect(getWeatherCondition(1, 10, 30).text).toBe("Chuva Forte / Tempestade");
    expect(getWeatherCondition(3, 80, 0).text).toBe("Nublado / Encoberto");
    expect(getWeatherCondition(0, 10, 0).text).toBe("Ensolarado / Céu Limpo");
  });
});

describe("getFarmingRecommendation", () => {
  it("warns against spraying in rain or wind before anything else", () => {
    expect(getFarmingRecommendation(20, 36, 0, 30)).toMatch(/defensivos/);
    expect(getFarmingRecommendation(0, 36, 0, 30)).toMatch(/Ventos fortes/);
    expect(getFarmingRecommendation(0, 36, 0, 5)).toMatch(/Calor intenso/);
  });
});
