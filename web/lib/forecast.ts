import type { WeatherReading } from "./api";

export type DailyForecast = {
  date: string;
  minTempC: number;
  maxTempC: number;
  totalPrecipitationMM: number;
};

/** Groups hourly readings into per-day min/max temp and total precipitation, in first-seen order. */
export function groupByDay(readings: WeatherReading[]): DailyForecast[] {
  const byDate = new Map<string, DailyForecast>();

  for (const r of readings) {
    const date = r.time.slice(0, 10);
    const existing = byDate.get(date);
    if (!existing) {
      byDate.set(date, {
        date,
        minTempC: r.temp_avg_c,
        maxTempC: r.temp_avg_c,
        totalPrecipitationMM: r.precipitation_mm,
      });
      continue;
    }
    existing.minTempC = Math.min(existing.minTempC, r.temp_avg_c);
    existing.maxTempC = Math.max(existing.maxTempC, r.temp_avg_c);
    existing.totalPrecipitationMM += r.precipitation_mm;
  }

  return Array.from(byDate.values());
}
