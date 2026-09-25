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
    <div className="flex flex-1 flex-col gap-6 p-6">
      <Link href="/" className="text-sm text-blue-600 underline">
        &larr; Voltar ao mapa
      </Link>

      {error && <p className="text-red-600">{error}</p>}

      <section>
        <h2 className="mb-3 text-lg font-semibold">Previsão para os próximos dias</h2>
        {daily.length === 0 ? (
          <p className="text-sm text-zinc-500">Sem dados de previsão disponíveis.</p>
        ) : (
          <table className="w-full max-w-2xl border-collapse text-sm">
            <thead>
              <tr className="border-b border-black/10 text-left dark:border-white/10">
                <th className="py-2">Data</th>
                <th className="py-2">Mín (°C)</th>
                <th className="py-2">Máx (°C)</th>
                <th className="py-2">Chuva (mm)</th>
              </tr>
            </thead>
            <tbody>
              {daily.map((day) => (
                <tr key={day.date} className="border-b border-black/5 dark:border-white/5">
                  <td className="py-2">{day.date}</td>
                  <td className="py-2">{day.minTempC.toFixed(1)}</td>
                  <td className="py-2">{day.maxTempC.toFixed(1)}</td>
                  <td className="py-2">{day.totalPrecipitationMM.toFixed(1)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      <section>
        <h2 className="mb-3 text-lg font-semibold">Histórico de alertas</h2>
        {alerts.length === 0 ? (
          <p className="text-sm text-zinc-500">Nenhum alerta registrado.</p>
        ) : (
          <ul className="flex flex-col gap-2">
            {alerts.map((alert) => (
              <li
                key={alert.id}
                className="rounded border border-black/10 px-3 py-2 text-sm dark:border-white/10"
              >
                <div className="flex items-center justify-between">
                  <span className="font-medium">
                    {alertTypeLabels[alert.alert_type] ?? alert.alert_type}
                  </span>
                  <span>{riskLabels[alert.level]}</span>
                </div>
                <div className="text-zinc-500">
                  {new Date(alert.period_start).toLocaleDateString("pt-BR")} a{" "}
                  {new Date(alert.period_end).toLocaleDateString("pt-BR")}
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
