import type { Alert, RiskLevel } from "./types";

export const riskColors: Record<RiskLevel, string> = {
  low: "#3f7d4f",
  medium: "#b8862a",
  critical: "#b23a2c",
};

export const riskLabels: Record<RiskLevel, string> = {
  low: "Baixo",
  medium: "Médio",
  critical: "Crítico",
};

export const cropLabels: Record<string, string> = {
  milho: "Milho",
  feijao: "Feijão / Caupi",
  mandioca: "Mandioca / Macaxeira",
  caju: "Caju / Cajueiro",
  tomate: "Tomate",
  cafe: "Café",
  soja: "Soja",
  algodao: "Algodão",
  fruticultura: "Fruticultura Tropical",
  cana: "Cana-de-Açúcar",
  hortalicas: "Hortaliças",
};

export const soilLabels: Record<string, string> = {
  medio: "Médio / Franco",
  argiloso: "Argiloso",
  arenoso: "Arenoso",
  raso: "Raso / Vertissolo",
  aluvial: "Aluvial / Várzea",
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
