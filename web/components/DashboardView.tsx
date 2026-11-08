"use client";

import { useState } from "react";
import Link from "next/link";
import PropertyMapClient from "./PropertyMapClient";
import type { Property, Alert } from "@/lib/types";
import { currentRiskLevel, riskLabels, cropLabels, soilLabels } from "@/lib/risk";
import { matchesConcern } from "@/lib/personalize";
import {
  SproutIcon,
  EyeIcon,
  CheckCircleIcon,
  AlertTriangleIcon,
} from "./Icons";

type PropertyWithAlerts = {
  property: Property;
  alerts: Alert[];
};

const riskBadgeClass: Record<string, string> = {
  low: "bg-risk-low-bg text-risk-low border-risk-low",
  medium: "bg-risk-medium-bg text-risk-medium border-risk-medium",
  critical: "bg-risk-critical-bg text-risk-critical border-risk-critical pulse-critical",
};

const cropStageLabels: Record<string, string> = {
  germination: "Germinação",
  flowering: "Floração",
  harvest: "Colheita",
};

type DashboardViewProps = {
  /** Already in priority order (see sortByPriority). */
  items: PropertyWithAlerts[];
  /** Risk types the account owner picked in the onboarding quiz. */
  concerns: string[];
  isTechnician: boolean;
  loadFailed?: boolean;
};

const ctaPrimary =
  "flex min-h-11 items-center border border-ink bg-accent px-2.5 font-mono text-[11px] font-semibold uppercase tracking-wide text-paper transition hover:bg-ink";
const ctaSecondary =
  "flex min-h-11 items-center border border-ink bg-card px-2.5 font-mono text-[11px] font-semibold uppercase tracking-wide text-ink transition hover:bg-ink hover:text-paper";

export default function DashboardView({ items, concerns, isTechnician, loadFailed }: DashboardViewProps) {
  const [simplified, setSimplified] = useState(false);
  // technicians start from producers; producers start from their own areas
  const producerCta = isTechnician ? ctaPrimary : ctaSecondary;
  const propertyCta = isTechnician ? ctaSecondary : ctaPrimary;
  const empty =
    items.length === 0 && !loadFailed ? (
      <div className="border-2 border-dashed border-line px-4 py-6 text-center font-mono text-sm text-ink-soft">
        {isTechnician ? "Você ainda não cadastrou produtores." : "Nenhuma propriedade cadastrada."}
        <div className="mt-3 flex justify-center">
          <Link href={isTechnician ? "/producers/new" : "/properties/new"} className={ctaPrimary}>
            {isTechnician ? "+ Produtor" : "+ Talhão"}
          </Link>
        </div>
      </div>
    ) : null;

  const counts = items.reduce(
    (acc, { alerts }) => {
      acc[currentRiskLevel(alerts)] += 1;
      return acc;
    },
    { low: 0, medium: 0, critical: 0 }
  );

  return (
    <div className="flex flex-1 flex-col h-full min-h-0 overflow-y-auto lg:overflow-hidden">
      {/* SELETOR DE MODO NA BARRA SUPERIOR (SHRINK-0) */}
      <div className="shrink-0 flex flex-wrap items-center justify-between gap-3 border-b-2 border-ink bg-paper px-4 py-1.5 sm:px-6">
        <div className="flex items-center gap-2 font-mono text-xs text-ink">
          <EyeIcon className="h-4 w-4 text-accent" />
          <span className="font-bold uppercase tracking-wider text-[11px]">Visualização:</span>
        </div>

        <div className="flex rounded border border-ink p-0.5 bg-card">
          <button
            type="button"
            onClick={() => setSimplified(false)}
            aria-pressed={!simplified}
            className={`min-h-11 px-3 py-0.5 font-mono text-[11px] font-bold transition ${
              !simplified
                ? "bg-ink text-paper"
                : "text-ink hover:bg-line/40"
            }`}
          >
            Mapa & Talhões (Completo)
          </button>
          <button
            type="button"
            onClick={() => setSimplified(true)}
            aria-pressed={simplified}
            className={`min-h-11 px-3 py-0.5 font-mono text-[11px] font-bold transition ${
              simplified
                ? "bg-accent text-paper"
                : "text-ink hover:bg-line/40"
            }`}
          >
            Modo Fácil (Produtor)
          </button>
        </div>
      </div>

      {simplified ? (
        /* ================= MODO SIMPLIFICADO / PAINEL DO AGRICULTOR ================= */
        <div className="flex-1 overflow-y-auto p-4 sm:p-6 min-h-0 flex flex-col gap-4">
          {/* SEMÁFORO GERAL DO CAMPO (COMPACTO) */}
          <div className="rounded border-2 border-ink bg-card p-4 shadow-[var(--shadow-hard)]">
            <h2 className="font-display text-xl font-bold italic text-ink">
              Situação Geral das Suas Plantações
            </h2>
            <p className="text-xs text-ink-soft">
              Resumo fácil para você saber como está cada talhão hoje sem complicação:
            </p>

            <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-3">
              <div className="flex items-center gap-3 rounded border-2 border-risk-low bg-risk-low-bg/40 p-3">
                <CheckCircleIcon className="h-6 w-6 text-risk-low shrink-0" />
                <div>
                  <span className="font-display text-xl font-bold text-risk-low">
                    {counts.low}
                  </span>
                  <p className="font-mono text-[11px] font-bold uppercase text-risk-low">
                    Tranquilos (Sem Risco)
                  </p>
                  <p className="text-[10px] text-ink-soft">Clima e solo em boa condição</p>
                </div>
              </div>

              <div className="flex items-center gap-3 rounded border-2 border-risk-medium bg-risk-medium-bg/40 p-3">
                <AlertTriangleIcon className="h-6 w-6 text-risk-medium shrink-0" />
                <div>
                  <span className="font-display text-xl font-bold text-risk-medium">
                    {counts.medium}
                  </span>
                  <p className="font-mono text-[11px] font-bold uppercase text-risk-medium">
                    Em Atenção
                  </p>
                  <p className="text-[10px] text-ink-soft">Calor forte ou chuva se aproximando</p>
                </div>
              </div>

              <div className="flex items-center gap-3 rounded border-2 border-risk-critical bg-risk-critical-bg/40 p-3">
                <AlertTriangleIcon className="h-6 w-6 text-risk-critical shrink-0 animate-pulse" />
                <div>
                  <span className="font-display text-xl font-bold text-risk-critical">
                    {counts.critical}
                  </span>
                  <p className="font-mono text-[11px] font-bold uppercase text-risk-critical">
                    Ação Necessária
                  </p>
                  <p className="text-[10px] text-ink-soft">Risco severo de seca ou perda</p>
                </div>
              </div>
            </div>
          </div>

          {/* LISTAGEM DE TALHÕES EM CARDS FACILITADOS */}
          <div>
            <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
              <h3 className="font-display text-lg italic text-ink">
                Seus Talhões e O Que Fazer Hoje ({items.length})
              </h3>
              <div className="flex gap-2">
                <Link href="/producers/new" className={producerCta}>
                  + Produtor
                </Link>
                <Link href="/properties/new" className={propertyCta}>
                  + Plantação
                </Link>
              </div>
            </div>

            {empty}
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {items.map(({ property, alerts }) => {
                const level = currentRiskLevel(alerts);
                const prioritized = matchesConcern(alerts, concerns);
                const cropName = cropLabels[property.crop.toLowerCase()] || property.crop;
                const soilName = soilLabels[property.soil_type.toLowerCase()] || property.soil_type;

                let advice = "Tudo em ordem. Mantenha os tratos culturais habituais.";
                if (level === "critical") {
                  advice = "Solo com pouca água e calor forte! Regue à noite ou nas primeiras horas da manhã.";
                } else if (level === "medium") {
                  advice = "Mudança climática prevista (chuva forte ou calor elevado). Fique atento aos drenos.";
                }

                return (
                  <div
                    key={property.id}
                    className={`flex flex-col justify-between rounded-lg border-2 p-3.5 transition shadow-[var(--shadow-hard)] ${
                      level === "critical"
                        ? "border-risk-critical bg-risk-critical-bg/15"
                        : level === "medium"
                        ? "border-risk-medium bg-risk-medium-bg/15"
                        : "border-ink bg-card"
                    }`}
                  >
                    <div>
                      <div className="flex items-start justify-between gap-2">
                        <div className="flex items-center gap-2">
                          <SproutIcon className="h-5 w-5 text-accent shrink-0" />
                          <div>
                            <h4 className="font-display text-base font-bold italic text-ink">
                              {cropName}
                            </h4>
                            <p className="font-mono text-[10px] text-ink-soft">
                              Solo: {soilName}
                            </p>
                          </div>
                        </div>

                        <div className="flex shrink-0 flex-col items-end gap-1">
                          {prioritized && <PriorityTag />}
                          <span
                            className={`rounded border-2 px-2 py-0.5 font-mono text-[10px] font-bold uppercase ${riskBadgeClass[level]}`}
                          >
                            {riskLabels[level]}
                          </span>
                        </div>
                      </div>

                      <div className="mt-2.5 rounded border border-line bg-paper p-2">
                        <p className="font-mono text-[9px] font-bold uppercase tracking-wider text-ink-soft">
                          Orientação para hoje:
                        </p>
                        <p className="text-xs font-medium text-ink mt-0.5">
                          {advice}
                        </p>
                      </div>
                    </div>

                    <div className="mt-3 flex items-center justify-between border-t border-line pt-2 font-mono text-xs">
                      <span className="text-[10px] text-ink-soft">
                        Fase: {cropStageLabels[property.crop_stage] ?? property.crop_stage}
                      </span>
                      <Link
                        href={`/properties/${property.id}`}
                        className="flex min-h-11 items-center font-bold text-accent hover:underline text-[11px]"
                      >
                        Ver Clima e Nuvens &rarr;
                      </Link>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      ) : (
        /* ================= MODO COMPLETO (ESTÁTICO NO PC: MAPA + TALHÕES OCUPAM 100% DA ALTURA) ================= */
        <div className="flex flex-1 flex-col gap-3 p-3 sm:p-4 lg:flex-row lg:overflow-hidden lg:h-full min-h-0">
          {/* MAPA INTERATIVO (NO PC OCUPA EXATAMENTE O ESPAÇO DISPONÍVEL) */}
          <div
            className="h-[360px] sm:h-[420px] shrink-0 overflow-hidden border-2 border-ink shadow-[var(--shadow-hard)] lg:h-full lg:flex-1 lg:shrink-0 min-h-0 rise-in"
            style={{ animationDelay: "140ms" }}
          >
            <PropertyMapClient items={items} />
          </div>

          {/* BARRA LATERAL DE TALHÕES (NO PC TEM ALTURA FIXA E SCROLL INTERNO) */}
          <aside className="flex w-full flex-col lg:w-96 lg:h-full min-h-0">
            <div className="shrink-0 mb-2 flex items-center justify-between gap-2">
              <h2 className="font-display text-base italic text-ink">Talhões ({items.length})</h2>
              <div className="flex gap-1.5">
                <Link href="/producers/new" className={producerCta}>
                  + Produtor
                </Link>
                <Link href="/properties/new" className={propertyCta}>
                  + Talhão
                </Link>
              </div>
            </div>

            {empty}
            <ul className="flex flex-1 flex-col gap-2 overflow-y-auto pr-1 min-h-0">
              {items.map(({ property, alerts }, i) => {
                const level = currentRiskLevel(alerts);
                const prioritized = matchesConcern(alerts, concerns);
                return (
                  <li
                    key={property.id}
                    className="rise-in shrink-0"
                    style={{ animationDelay: `${180 + i * 30}ms` }}
                  >
                    <Link
                      href={`/properties/${property.id}`}
                      className="group flex min-h-11 items-center justify-between gap-3 border border-line bg-card px-3 py-2.5 transition hover:border-ink hover:shadow-[var(--shadow-hard)]"
                    >
                      <div className="min-w-0">
                        <p className="truncate font-display text-sm font-semibold text-ink">
                          {cropLabels[property.crop.toLowerCase()] || property.crop}
                        </p>
                        <p className="truncate font-mono text-[10px] uppercase tracking-wide text-ink-soft">
                          {cropStageLabels[property.crop_stage] ?? property.crop_stage} &middot;{" "}
                          Solo {soilLabels[property.soil_type.toLowerCase()] || property.soil_type}
                        </p>
                      </div>
                      <div className="flex shrink-0 flex-col items-end gap-1">
                        {prioritized && <PriorityTag />}
                        <span
                          className={`border px-2 py-0.5 font-mono text-[10px] font-bold uppercase tracking-wide ${riskBadgeClass[level]}`}
                        >
                          {riskLabels[level]}
                        </span>
                      </div>
                    </Link>
                  </li>
                );
              })}
            </ul>
          </aside>
        </div>
      )}
    </div>
  );
}

function PriorityTag() {
  return (
    <span className="font-mono text-[10px] font-semibold uppercase tracking-wide text-accent">
      Prioridade
    </span>
  );
}
