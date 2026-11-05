"use client";

import { useState } from "react";
import type { Alert } from "@/lib/api";
import { resolveAlert } from "@/lib/api";
import { riskLabels } from "@/lib/risk";

const alertTypeLabels: Record<string, string> = {
  severe_water_stress: "Estresse hídrico severo",
  frost: "Geada",
  heat_wave: "Onda de calor",
  heavy_rain: "Chuva forte",
  none: "Sem risco",
};

const riskBadgeClass: Record<string, string> = {
  low: "bg-risk-low-bg text-risk-low border-risk-low",
  medium: "bg-risk-medium-bg text-risk-medium border-risk-medium",
  critical: "bg-risk-critical-bg text-risk-critical border-risk-critical",
};

export default function AlertList({ initialAlerts }: { initialAlerts: Alert[] }) {
  const [alerts, setAlerts] = useState<Alert[]>(initialAlerts);
  const [resolvingId, setResolvingId] = useState<string | null>(null);

  async function handleResolve(id: string) {
    setResolvingId(id);
    try {
      await resolveAlert(id);
      setAlerts((prev) =>
        prev.map((a) => (a.id === id ? { ...a, resolved_at: new Date().toISOString() } : a))
      );
    } catch {
      alert("Não foi possível resolver o alerta. Tente novamente.");
    } finally {
      setResolvingId(null);
    }
  }

  if (alerts.length === 0) {
    return (
      <p className="border-2 border-dashed border-line px-4 py-6 font-mono text-sm text-ink-soft">
        Nenhum alerta registrado para esta propriedade.
      </p>
    );
  }

  return (
    <ul className="flex flex-col gap-3">
      {alerts.map((alert) => {
        const isResolved = Boolean(alert.resolved_at);
        return (
          <li
            key={alert.id}
            className={`flex flex-col gap-2 border-2 p-4 transition sm:flex-row sm:items-center sm:justify-between ${
              isResolved
                ? "border-line bg-paper-deep/30 opacity-75"
                : "border-line bg-card shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]"
            }`}
          >
            <div>
              <div className="flex items-center gap-2">
                <span className="font-display text-base font-semibold text-ink">
                  {alertTypeLabels[alert.alert_type] ?? alert.alert_type}
                </span>
                {isResolved && (
                  <span className="rounded bg-paper-deep px-2 py-0.5 font-mono text-[10px] uppercase font-bold text-ink-soft border border-line">
                    Resolvido
                  </span>
                )}
              </div>
              <div className="mt-1 font-mono text-xs text-ink-soft">
                Período crítico: {new Date(alert.period_start).toLocaleDateString("pt-BR")} a{" "}
                {new Date(alert.period_end).toLocaleDateString("pt-BR")} &middot; Disparado em:{" "}
                {new Date(alert.triggered_at).toLocaleString("pt-BR")}
              </div>
            </div>

            <div className="flex items-center gap-3 self-end sm:self-auto">
              <span
                className={`w-fit shrink-0 border-2 px-2.5 py-1 font-mono text-[11px] font-semibold uppercase tracking-wide ${
                  riskBadgeClass[alert.level]
                }`}
              >
                {riskLabels[alert.level]}
              </span>

              {!isResolved && (
                <button
                  type="button"
                  onClick={() => handleResolve(alert.id)}
                  disabled={resolvingId === alert.id}
                  className="border-2 border-ink bg-paper px-3 py-1 font-mono text-xs font-semibold uppercase tracking-wide text-ink transition hover:bg-ink hover:text-paper disabled:opacity-50"
                >
                  {resolvingId === alert.id ? "Resolvendo..." : "Resolver"}
                </button>
              )}
            </div>
          </li>
        );
      })}
    </ul>
  );
}

