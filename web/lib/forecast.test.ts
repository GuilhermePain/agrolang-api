import { describe, expect, it } from "vitest";
import { groupByDay } from "./forecast";
import type { WeatherReading } from "./api";

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

    expect(groupByDay(readings)).toEqual([
      { date: "2026-01-01", minTempC: 10, maxTempC: 30, totalPrecipitationMM: 3.5 },
    ]);
  });

  it("splits readings into separate entries per day, in first-seen order", () => {
    const readings = [
      reading("2026-01-02T00:00", 15),
      reading("2026-01-01T00:00", 10),
      reading("2026-01-02T12:00", 25),
    ];

    expect(groupByDay(readings)).toEqual([
      { date: "2026-01-02", minTempC: 15, maxTempC: 25, totalPrecipitationMM: 0 },
      { date: "2026-01-01", minTempC: 10, maxTempC: 10, totalPrecipitationMM: 0 },
    ]);
  });
});
