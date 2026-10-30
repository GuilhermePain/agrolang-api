import { currentRiskLevel } from "./risk";
import type { Alert, AuthUser } from "./types";

const LEVEL_RANK = { low: 0, medium: 1, critical: 2 } as const;

export const CONCERN_LABELS: Record<string, string> = {
  frost: "Geada",
  severe_water_stress: "Seca",
  heat_wave: "Calor forte",
  heavy_rain: "Chuva forte",
};

/** True when an unresolved alert is of a type the account owner asked to prioritize. */
export function matchesConcern(alerts: Alert[], concerns: string[]): boolean {
  if (concerns.length === 0) return false;
  return alerts.some((a) => a.resolved_at === null && concerns.includes(a.alert_type));
}

/** Highest risk first; within a level, areas with a chosen risk type come first. */
export function sortByPriority<T extends { alerts: Alert[] }>(items: T[], concerns: string[]): T[] {
  const score = (item: T) =>
    LEVEL_RANK[currentRiskLevel(item.alerts)] * 2 + (matchesConcern(item.alerts, concerns) ? 1 : 0);
  return [...items].sort((a, b) => score(b) - score(a));
}

export function currentHourBR(now: Date = new Date()): number {
  const hour = new Intl.DateTimeFormat("en-US", {
    hour: "numeric",
    hourCycle: "h23",
    timeZone: "America/Sao_Paulo",
  }).format(now);
  return Number(hour);
}

export function greeting(name: string, hour: number): string {
  const salutation = hour < 12 ? "Bom dia" : hour < 18 ? "Boa tarde" : "Boa noite";
  return `${salutation}, ${name.trim().split(/\s+/)[0]}`;
}

const AREA_BUCKETS: Record<string, { min: number; label: string }> = {
  "2-5": { min: 2, label: "de 2 a 5" },
  "6+": { min: 6, label: "mais de 5" },
};

/** Reminder for producers who announced more areas than they have registered. */
export function missingAreasNotice(bucket: string | undefined, registered: number): string | null {
  const info = bucket ? AREA_BUCKETS[bucket] : undefined;
  if (!info || registered >= info.min) return null;
  const registeredText = registered === 1 ? "1 cadastrada" : `${registered} cadastradas`;
  return `Você disse ter ${info.label} áreas e há ${registeredText}. Cadastre as outras para acompanhar todas.`;
}

export function joinList(items: string[]): string {
  if (items.length <= 1) return items.join("");
  return `${items.slice(0, -1).join(", ")} e ${items[items.length - 1]}`;
}

export function welcomeMessage(user: Pick<AuthUser, "profile" | "concerns">, city?: string): string {
  const labels = user.concerns.map((c) => CONCERN_LABELS[c]).filter(Boolean).map((l) => l.toLowerCase());

  if (user.profile === "technician") {
    return labels.length > 0
      ? `Vamos destacar ${joinList(labels)} no seu painel.`
      : "Vamos acompanhar todos os riscos do clima no seu painel.";
  }

  const what = labels.length > 0 ? `de ${joinList(labels)}` : "de qualquer risco do clima";
  return city ? `Vamos te avisar ${what} em ${city}.` : `Vamos te avisar ${what} na sua terra.`;
}
