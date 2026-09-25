import Link from "next/link";
import { getPropertyAlerts, getPropertyForecast } from "@/lib/api";
import { groupByDay } from "@/lib/forecast";
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

export default async function PropertyDetail({ params }: PageProps<"/properties/[id]">) {
  const { id } = await params;

  let alerts: Awaited<ReturnType<typeof getPropertyAlerts>> = [];
  let daily: ReturnType<typeof groupByDay> = [];
  let error: string | null = null;

  try {
    const [alertsResult, forecastResult] = await Promise.all([
      getPropertyAlerts(id),
      getPropertyForecast(id),
    ]);
    alerts = alertsResult;
    daily = groupByDay(forecastResult).slice(0, 7);
  } catch {
    error = "Não foi possível carregar os dados desta propriedade.";
  }

  return (
    <div className="flex flex-1 flex-col gap-8 p-5 sm:p-8">
      <div className="rise-in">
        <Link
          href="/"
          className="inline-flex items-center gap-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-accent"
        >
          &larr; Voltar ao mapa
        </Link>
        <p className="mt-2 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
          Ficha do talhão
        </p>
      </div>

      {error && (
        <p className="border-2 border-risk-critical bg-risk-critical-bg px-4 py-3 font-mono text-sm text-risk-critical">
          {error}
        </p>
      )}

      <section className="rise-in" style={{ animationDelay: "80ms" }}>
        <h2 className="mb-3 font-display text-xl italic text-ink">
          Previsão para os próximos dias
        </h2>
        {daily.length === 0 ? (
          <p className="border-2 border-dashed border-line px-4 py-6 font-mono text-sm text-ink-soft">
            Sem dados de previsão disponíveis.
          </p>
        ) : (
          <div className="overflow-x-auto border-2 border-ink bg-card shadow-[var(--shadow-hard)]">
            <table className="w-full min-w-[520px] border-collapse font-mono text-sm">
              <thead>
                <tr className="border-b-2 border-ink bg-paper-deep text-left uppercase tracking-wide text-[11px] text-ink-soft">
                  <th className="px-4 py-2.5 font-semibold">Data</th>
                  <th className="px-4 py-2.5 font-semibold">Mín (°C)</th>
                  <th className="px-4 py-2.5 font-semibold">Máx (°C)</th>
                  <th className="px-4 py-2.5 font-semibold">Chuva (mm)</th>
                </tr>
              </thead>
              <tbody>
                {daily.map((day, i) => (
                  <tr
                    key={day.date}
                    className={`border-b border-line ${i % 2 === 1 ? "bg-paper" : ""}`}
                  >
                    <td className="px-4 py-2 tabular">{day.date}</td>
                    <td className="px-4 py-2 tabular">{day.minTempC.toFixed(1)}</td>
                    <td className="px-4 py-2 tabular">{day.maxTempC.toFixed(1)}</td>
                    <td className="px-4 py-2 tabular">{day.totalPrecipitationMM.toFixed(1)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section className="rise-in" style={{ animationDelay: "160ms" }}>
        <h2 className="mb-3 font-display text-xl italic text-ink">Histórico de alertas</h2>
        {alerts.length === 0 ? (
          <p className="border-2 border-dashed border-line px-4 py-6 font-mono text-sm text-ink-soft">
            Nenhum alerta registrado.
          </p>
        ) : (
          <ul className="flex flex-col gap-2.5">
            {alerts.map((alert) => (
              <li
                key={alert.id}
                className="flex flex-col gap-1 border-2 border-line bg-card px-4 py-3 sm:flex-row sm:items-center sm:justify-between"
              >
                <div>
                  <span className="font-display text-base text-ink">
                    {alertTypeLabels[alert.alert_type] ?? alert.alert_type}
                  </span>
                  <div className="font-mono text-xs text-ink-soft">
                    {new Date(alert.period_start).toLocaleDateString("pt-BR")} a{" "}
                    {new Date(alert.period_end).toLocaleDateString("pt-BR")}
                  </div>
                </div>
                <span
                  className={`w-fit shrink-0 border-2 px-2 py-1 font-mono text-[11px] font-semibold uppercase tracking-wide ${riskBadgeClass[alert.level]}`}
                >
                  {riskLabels[alert.level]}
                </span>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
