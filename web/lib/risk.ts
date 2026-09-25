import type { Alert, RiskLevel } from "./api";

export const riskColors: Record<RiskLevel, string> = {
  low: "#22c55e",
  medium: "#eab308",
  critical: "#ef4444",
};

export const riskLabels: Record<RiskLevel, string> = {
  low: "Baixo",
  medium: "Médio",
  critical: "Crítico",
};

/** Current risk is the level of the most recently triggered, unresolved alert; defaults to low. */
export function currentRiskLevel(alerts: Alert[]): RiskLevel {
  const active = alerts.filter((a) => a.resolved_at === null);
  if (active.length === 0) return "low";

  const mostRecent = active.reduce((latest, a) =>
    new Date(a.triggered_at) > new Date(latest.triggered_at) ? a : latest
  );
  return mostRecent.level;
}
