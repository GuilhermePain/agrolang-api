"use client";

import { useState } from "react";
import type { DailyForecast } from "@/lib/forecast";
import {
  SunIcon,
  SunCloudIcon,
  CloudIcon,
  RainIcon,
  WindIcon,
  DropletIcon,
  EyeIcon,
  CheckCircleIcon,
} from "./Icons";

interface PropertyWeatherPanelProps {
  daily: DailyForecast[];
  crop: string;
  soil: string;
}

function WeatherConditionIcon({ code, cloudPct, rainMm, className = "w-5 h-5" }: { code: number; cloudPct: number; rainMm: number; className?: string }) {
  if (rainMm > 0.5 || code >= 50) {
    return <RainIcon className={`${className} text-blue-600`} />;
  }
  if (cloudPct >= 70 || code === 3) {
    return <CloudIcon className={`${className} text-stone-500`} />;
  }
  if (cloudPct >= 30 || code === 1 || code === 2) {
    return <SunCloudIcon className={`${className} text-amber-500`} />;
  }
  return <SunIcon className={`${className} text-amber-500`} />;
}

export default function PropertyWeatherPanel({ daily, crop, soil }: PropertyWeatherPanelProps) {
  const [simplifiedMode, setSimplifiedMode] = useState(true);

  if (!daily || daily.length === 0) {
    return (
      <p className="border-2 border-dashed border-line px-4 py-6 font-mono text-xs text-ink-soft">
        Sem dados de expectativa do clima disponíveis no momento.
      </p>
    );
  }

  const today = daily[0];

  return (
    <div className="flex flex-col gap-3 lg:min-h-0">
      {/* SELETOR DE MODO: SIMPLIFICADO VS TÉCNICO (SHRINK-0) */}
      <div className="shrink-0 flex flex-wrap items-center justify-between gap-2 rounded border border-ink bg-card px-3 py-1.5 shadow-[var(--shadow-hard)]">
        <div className="flex items-center gap-2">
          <EyeIcon className="h-4 w-4 text-accent" />
          <div>
            <h3 className="font-display text-sm italic font-bold text-ink">
              Expectativa do Clima & Nuvens
            </h3>
            <p className="font-mono text-[10px] text-ink-soft">
              {simplifiedMode ? "Modo Fácil para o Produtor" : "Modo Técnico Completo"}
            </p>
          </div>
        </div>

        <div className="flex rounded border border-ink p-0.5 bg-paper">
          <button
            type="button"
            onClick={() => setSimplifiedMode(true)}
            className={`touch px-2.5 py-0.5 font-mono text-[11px] font-bold transition ${
              simplifiedMode
                ? "bg-accent text-paper"
                : "text-ink hover:bg-line/40"
            }`}
          >
            Modo Fácil
          </button>
          <button
            type="button"
            onClick={() => setSimplifiedMode(false)}
            className={`touch px-2.5 py-0.5 font-mono text-[11px] font-bold transition ${
              !simplifiedMode
                ? "bg-accent text-paper"
                : "text-ink hover:bg-line/40"
            }`}
          >
            Modo Técnico
          </button>
        </div>
      </div>

      {simplifiedMode ? (
        /* ================= MODO SIMPLIFICADO / VISÃO DO AGRICULTOR ================= */
        <div className="flex flex-col gap-3 lg:min-h-0">
          {/* CARTÃO PRINCIPAL: COMO ESTÁ O CLIMA HOJE */}
          <div className="rounded border-2 border-ink bg-paper p-3.5 shadow-[var(--shadow-hard)]">
            <div className="flex flex-col gap-2.5 sm:flex-row sm:items-center sm:justify-between border-b border-line pb-3">
              <div className="flex items-center gap-3">
                <div className="flex h-11 w-11 items-center justify-center rounded-full border border-ink bg-card shrink-0">
                  <WeatherConditionIcon
                    code={today.weatherCode}
                    cloudPct={today.avgCloudCoverPct}
                    rainMm={today.totalPrecipitationMM}
                    className="h-6 w-6"
                  />
                </div>
                <div>
                  <span className="font-mono text-[10px] font-bold uppercase tracking-wider text-accent">
                    Previsão para Hoje ({today.date})
                  </span>
                  <h4 className="font-display text-xl font-bold italic text-ink">
                    {today.conditionText}
                  </h4>
                  <p className="text-xs text-ink-soft">
                    {today.simplifiedSummary}
                  </p>
                </div>
              </div>

              <div className="flex gap-2 font-mono text-xs">
                <div className="flex flex-col items-center justify-center rounded border border-ink bg-card px-2.5 py-1 text-center">
                  <span className="text-[9px] uppercase text-ink-soft">Temp.</span>
                  <span className="font-bold text-ink">
                    {today.minTempC.toFixed(0)}° - {today.maxTempC.toFixed(0)}°C
                  </span>
                </div>
                <div className="flex flex-col items-center justify-center rounded border border-ink bg-card px-2.5 py-1 text-center">
                  <span className="text-[9px] uppercase text-ink-soft">Chuva</span>
                  <span className="font-bold text-ink">
                    {today.totalPrecipitationMM > 0 ? `${today.totalPrecipitationMM.toFixed(1)} mm` : "Sem chuva"}
                  </span>
                </div>
              </div>
            </div>

            {/* BARRA DAS NUVENS E VENTOS DE HOJE */}
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-3 pt-2.5 font-mono text-xs">
              <div className="flex items-center gap-2 rounded border border-line bg-card p-2">
                <CloudIcon className="h-4 w-4 text-ink-soft shrink-0" />
                <div className="flex-1 min-w-0">
                  <div className="flex justify-between text-[10px] font-medium text-ink">
                    <span>Nuvens no céu:</span>
                    <span className="font-bold">{today.avgCloudCoverPct}%</span>
                  </div>
                  <div className="mt-0.5 h-1.5 w-full rounded-full bg-paper-deep overflow-hidden border border-line">
                    <div
                      className="h-full bg-accent transition-all"
                      style={{ width: `${today.avgCloudCoverPct}%` }}
                    />
                  </div>
                </div>
              </div>

              <div className="flex items-center gap-2 rounded border border-line bg-card p-2">
                <WindIcon className="h-4 w-4 text-ink-soft shrink-0" />
                <div className="min-w-0">
                  <span className="text-[9px] uppercase text-ink-soft">Vento Máximo</span>
                  <p className="font-bold text-ink text-[11px] truncate">
                    {today.maxWindSpeedKmh.toFixed(0)} km/h{" "}
                    <span className="font-normal text-ink-soft">
                      {today.maxWindSpeedKmh > 20 ? "(Forte)" : "(Calmo)"}
                    </span>
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-2 rounded border border-line bg-card p-2">
                <DropletIcon className="h-4 w-4 text-ink-soft shrink-0" />
                <div className="min-w-0">
                  <span className="text-[9px] uppercase text-ink-soft">Chance de Chuva</span>
                  <p className="font-bold text-ink text-[11px]">
                    {today.maxRainProbPct}%{" "}
                    <span className="font-normal text-ink-soft">
                      {today.maxRainProbPct > 50 ? "(Alta)" : "(Baixa)"}
                    </span>
                  </p>
                </div>
              </div>
            </div>

            {/* RECOMENDAÇÃO PRÁTICA DO DIA */}
            <div className="mt-2.5 flex items-start gap-2.5 rounded border border-ink bg-card p-2.5">
              <CheckCircleIcon className="mt-0.5 h-4 w-4 shrink-0 text-accent" />
              <div>
                <p className="font-mono text-[10px] font-bold uppercase tracking-wider text-ink">
                  Dica de Manejo para o talhão de {crop}:
                </p>
                <p className="text-xs text-ink-soft mt-0.5">
                  {today.farmingAdvice}
                </p>
              </div>
            </div>
          </div>

          {/* EXPECTATIVA PARA OS PRÓXIMOS DIAS (CARDS FACILITADOS) */}
          <div>
            <h4 className="font-display text-base italic text-ink mb-2">
              Expectativa para os Próximos Dias
            </h4>
            <div className="grid grid-cols-2 lg:grid-cols-3 gap-2">
              {daily.slice(1, 7).map((day) => (
                <div
                  key={day.date}
                  className="flex flex-col justify-between rounded border border-line bg-card p-2.5 transition hover:border-ink hover:shadow-[var(--shadow-hard)]"
                >
                  <div className="flex items-start justify-between gap-1">
                    <div className="min-w-0">
                      <span className="font-mono text-[10px] font-bold uppercase text-accent">
                        {day.date}
                      </span>
                      <p className="font-display text-sm font-semibold italic text-ink truncate">
                        {day.conditionText}
                      </p>
                    </div>
                    <WeatherConditionIcon
                      code={day.weatherCode}
                      cloudPct={day.avgCloudCoverPct}
                      rainMm={day.totalPrecipitationMM}
                      className="h-4 w-4 shrink-0"
                    />
                  </div>

                  <div className="mt-2 flex items-center justify-between border-t border-line pt-1.5 font-mono text-[10px] text-ink-soft">
                    <span>
                      {day.minTempC.toFixed(0)}° - {day.maxTempC.toFixed(0)}°C
                    </span>
                    <span>
                      {day.totalPrecipitationMM > 0 ? `${day.totalPrecipitationMM.toFixed(1)}mm` : "0mm"}
                    </span>
                    <span className="flex items-center gap-0.5">
                      <CloudIcon className="h-3 w-3" />
                      {day.avgCloudCoverPct}%
                    </span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      ) : (
        /* ================= MODO TÉCNICO DETALHADO ================= */
        <div className="overflow-x-auto border border-ink bg-card shadow-[var(--shadow-hard)]">
          <table className="w-full min-w-[580px] border-collapse font-mono text-xs">
            <thead>
              <tr className="border-b border-ink bg-paper-deep text-left uppercase tracking-wide text-[10px] text-ink-soft">
                <th className="px-3 py-2 font-semibold">Data</th>
                <th className="px-3 py-2 font-semibold">Condição</th>
                <th className="px-3 py-2 font-semibold">Mín/Máx</th>
                <th className="px-3 py-2 font-semibold">Chuva</th>
                <th className="px-3 py-2 font-semibold">Nuvens</th>
                <th className="px-3 py-2 font-semibold">Vento</th>
                <th className="px-3 py-2 font-semibold">Recomendação</th>
              </tr>
            </thead>
            <tbody>
              {daily.map((day, i) => (
                <tr
                  key={day.date}
                  className={`border-b border-line ${i % 2 === 1 ? "bg-paper" : ""}`}
                >
                  <td className="px-3 py-1.5 tabular font-semibold">{day.date}</td>
                  <td className="px-3 py-1.5">
                    <div className="flex items-center gap-1">
                      <WeatherConditionIcon
                        code={day.weatherCode}
                        cloudPct={day.avgCloudCoverPct}
                        rainMm={day.totalPrecipitationMM}
                        className="h-3.5 w-3.5 shrink-0"
                      />
                      <span className="truncate max-w-[120px]">{day.conditionText}</span>
                    </div>
                  </td>
                  <td className="px-3 py-1.5 tabular">{day.minTempC.toFixed(0)}° / {day.maxTempC.toFixed(0)}°C</td>
                  <td className="px-3 py-1.5 tabular">
                    {day.totalPrecipitationMM > 0 ? (
                      <span className="font-bold text-blue-600">{day.totalPrecipitationMM.toFixed(1)}mm</span>
                    ) : (
                      "0mm"
                    )}
                  </td>
                  <td className="px-3 py-1.5 tabular">{day.avgCloudCoverPct}%</td>
                  <td className="px-3 py-1.5 tabular">{day.maxWindSpeedKmh.toFixed(0)} km/h</td>
                  <td className="px-3 py-1.5 text-[10px] text-ink-soft max-w-xs truncate" title={day.farmingAdvice}>
                    {day.farmingAdvice}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
