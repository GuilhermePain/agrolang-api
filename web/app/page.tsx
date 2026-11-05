import Link from "next/link";
import { listProperties, getPropertyAlerts, listRecentAlerts, type Alert } from "@/lib/api";
import DashboardView from "@/components/DashboardView";
import { currentRiskLevel, riskLabels, cropLabels, soilLabels } from "@/lib/risk";

const riskBadgeClass: Record<string, string> = {
  low: "bg-risk-low-bg text-risk-low border-risk-low",
  medium: "bg-risk-medium-bg text-risk-medium border-risk-medium",
  critical: "bg-risk-critical-bg text-risk-critical border-risk-critical pulse-critical",
};

const alertTypeLabels: Record<string, string> = {
  severe_water_stress: "Estresse Hídrico",
  frost: "Geada",
  heat_wave: "Onda de Calor",
  heavy_rain: "Chuva Forte",
};

export default async function Home() {
  let items: Awaited<ReturnType<typeof loadItems>> = [];
  let recentAlerts: Alert[] = [];
  let error: string | null = null;

  try {
    const [loadedItems, loadedAlerts] = await Promise.all([
      loadItems(),
      listRecentAlerts().catch(() => []),
    ]);
    items = loadedItems;
    recentAlerts = loadedAlerts;
  } catch {
    error = "Não foi possível carregar as propriedades. Verifique se a API está em execução.";
  }

  const counts = items.reduce(
    (acc, { alerts }) => {
      acc[currentRiskLevel(alerts)] += 1;
      return acc;
    },
    { low: 0, medium: 0, critical: 0 }
  );

  const activeAlerts = recentAlerts.filter((a) => !a.resolved_at);

  return (
    <div className="flex flex-1 flex-col h-full min-h-0 overflow-y-auto lg:overflow-hidden">
      {/* CABEÇALHO COMPACTO (ESTÁTICO NO TOPO) */}
      <header className="shrink-0 relative border-b-2 border-ink px-4 py-2.5 sm:px-6">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <div className="flex items-center gap-2">
              <span className="font-mono text-[10px] uppercase tracking-[0.2em] text-ink-soft">
                Estação de Campo RF-05
              </span>
              <span className="h-1.5 w-1.5 rounded-full bg-accent" />
              <span className="font-mono text-[10px] text-accent font-semibold uppercase">
                Monitoramento Ativo
              </span>
            </div>
            <h1 className="font-display text-2xl font-bold italic tracking-tight text-ink sm:text-3xl">
              AgroResiliente
            </h1>
          </div>

          <div className="flex gap-2">
            <CountPill label="Baixo" count={counts.low} tone="low" />
            <CountPill label="Médio" count={counts.medium} tone="medium" />
            <CountPill label="Crítico" count={counts.critical} tone="critical" />
          </div>
        </div>
      </header>

      {error && (
        <p className="shrink-0 border-b-2 border-risk-critical bg-risk-critical-bg px-4 py-2 text-xs font-medium text-risk-critical sm:px-6">
          {error}
        </p>
      )}

      {/* FEED DE ALERTAS ATIVOS COMPACTO */}
      {activeAlerts.length > 0 && (
        <div className="shrink-0 border-b-2 border-ink bg-risk-critical-bg/30 px-4 py-1.5 sm:px-6">
          <div className="flex items-center justify-between gap-3">
            <div className="flex items-center gap-2">
              <span className="flex h-2 w-2 rounded-full bg-risk-critical animate-ping" />
              <span className="font-mono text-[11px] font-bold uppercase tracking-wider text-risk-critical">
                {activeAlerts.length} alerta(s) ativo(s)
              </span>
            </div>
            <div className="flex gap-2 overflow-x-auto py-0.5">
              {activeAlerts.slice(0, 4).map((alert) => (
                <Link
                  key={alert.id}
                  href={`/properties/${alert.property_id}`}
                  className="flex shrink-0 items-center gap-1.5 rounded border border-ink bg-paper px-2 py-0.5 font-mono text-[10px] text-ink transition hover:bg-ink hover:text-paper"
                >
                  <span className="font-bold">{alertTypeLabels[alert.alert_type] || alert.alert_type}</span>
                  <span className="text-ink-soft">({alert.level})</span>
                  <span>&rarr;</span>
                </Link>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* ÁREA PRINCIPAL DO DASHBOARD */}
      <DashboardView items={items} />
    </div>
  );
}

function CountPill({ label, count, tone }: { label: string; count: number; tone: string }) {
  return (
    <div className={`border-2 px-2.5 py-1 text-center font-mono ${riskBadgeClass[tone]}`}>
      <div className="text-base font-bold leading-none tabular">{count}</div>
      <div className="text-[9px] uppercase tracking-wide leading-none mt-0.5">{label}</div>
    </div>
  );
}

async function loadItems() {
  const properties = await listProperties();
  return Promise.all(
    properties.map(async (property) => ({
      property,
      alerts: await getPropertyAlerts(property.id),
    }))
  );
}
