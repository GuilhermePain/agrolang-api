import Link from "next/link";
import { notFound } from "next/navigation";
import {
  getProperty,
  getProducer,
  getPropertyAlerts,
  getPropertyForecast,
  type Property,
  type Producer,
} from "@/lib/api";
import { groupByDay } from "@/lib/forecast";
import { cropLabels, soilLabels } from "@/lib/risk";
import AlertList from "@/components/AlertList";
import PropertyWeatherPanel from "@/components/PropertyWeatherPanel";
import DeletePropertyAction from "@/components/DeletePropertyAction";

const cropStageLabels: Record<string, string> = {
  germination: "Germinação",
  flowering: "Floração",
  harvest: "Colheita",
};

export default async function PropertyDetail({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let property: Property | null = null;
  let producer: Producer | null = null;
  let alerts: Awaited<ReturnType<typeof getPropertyAlerts>> = [];
  let daily: ReturnType<typeof groupByDay> = [];
  let error: string | null = null;

  try {
    const propData = await getProperty(id);
    property = propData;

    const [producerData, alertsResult, forecastResult] = await Promise.all([
      getProducer(propData.producer_id).catch(() => null),
      getPropertyAlerts(id),
      getPropertyForecast(id),
    ]);
    producer = producerData;
    alerts = alertsResult;
    daily = groupByDay(forecastResult).slice(0, 7);
  } catch {
    error = "Não foi possível carregar os dados desta propriedade.";
  }

  if (!property && !error) {
    notFound();
  }

  return (
    <div className="flex flex-1 flex-col h-full min-h-0 overflow-y-auto lg:overflow-hidden p-3.5 sm:p-5">
      {/* CABEÇALHO COMPACTO (SHRINK-0) */}
      <div className="shrink-0 mb-3 flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between rise-in border-b border-line pb-2.5">
        <div>
          <Link
            href="/"
            className="inline-flex items-center gap-1 font-mono text-[11px] font-semibold uppercase tracking-wide text-accent hover:underline"
          >
            &larr; Voltar ao mapa
          </Link>
          {property && (
            <h1 className="font-display text-2xl font-bold italic text-ink sm:text-3xl">
              Talhão de {cropLabels[property.crop.toLowerCase()] || property.crop}
            </h1>
          )}
        </div>

        {property && (
          <div className="flex items-center gap-2">
            <Link
              href={`/properties/${property.id}/edit`}
              className="border border-ink bg-paper px-3 py-1 font-mono text-xs font-semibold uppercase tracking-wide text-ink hover:bg-ink hover:text-paper"
            >
              Editar Talhão
            </Link>
            <DeletePropertyAction propertyId={property.id} />
          </div>
        )}
      </div>

      {error && (
        <p className="shrink-0 mb-3 border-2 border-risk-critical bg-risk-critical-bg px-4 py-2 font-mono text-xs text-risk-critical">
          {error}
        </p>
      )}

      {/* GRADE EM 2 COLUNAS NO DESKTOP (TELA ESTÁTICA INTEGRADA) / FLUIDA NO MOBILE */}
      <div className="flex flex-1 flex-col gap-4 lg:grid lg:grid-cols-12 min-h-0 lg:overflow-hidden">
        {/* COLUNA ESQUERDA: METADADOS + HISTÓRICO DE ALERTAS */}
        <div className="lg:col-span-5 flex flex-col gap-3 min-h-0 lg:overflow-y-auto pr-1">
          {/* METADADOS DO TALHÃO */}
          {property && (
            <div className="grid grid-cols-2 gap-2.5 rounded border-2 border-ink bg-card p-3 shadow-[var(--shadow-hard)]">
              <div>
                <span className="font-mono text-[10px] uppercase tracking-wide text-ink-soft">
                  Produtor Responsável
                </span>
                <p className="font-display text-sm font-semibold text-ink truncate">
                  {producer ? (
                    <Link href={`/producers/${producer.id}`} className="hover:underline text-accent">
                      {producer.name}
                    </Link>
                  ) : (
                    "Carregando..."
                  )}
                </p>
              </div>

              <div>
                <span className="font-mono text-[10px] uppercase tracking-wide text-ink-soft">
                  Fase Atual
                </span>
                <p className="font-mono text-xs font-bold text-ink">
                  {cropStageLabels[property.crop_stage] ?? property.crop_stage}
                </p>
              </div>

              <div>
                <span className="font-mono text-[10px] uppercase tracking-wide text-ink-soft">
                  Tipo de Solo
                </span>
                <p className="font-mono text-xs font-bold text-ink">
                  {soilLabels[property.soil_type.toLowerCase()] || property.soil_type}
                </p>
              </div>

              <div>
                <span className="font-mono text-[10px] uppercase tracking-wide text-ink-soft">
                  Coordenadas GPS
                </span>
                <p className="font-mono text-[11px] text-ink-soft truncate">
                  {property.latitude.toFixed(4)}, {property.longitude.toFixed(4)}
                </p>
              </div>
            </div>
          )}

          {/* ALERTAS E AÇÕES */}
          <section className="flex flex-1 flex-col min-h-0">
            <h2 className="mb-2 font-display text-base italic text-ink">Histórico de Alertas e Ações</h2>
            <div className="flex-1 overflow-y-auto">
              <AlertList initialAlerts={alerts} />
            </div>
          </section>
        </div>

        {/* COLUNA DIREITA: EXPECTATIVA DO CLIMA & NUVENS */}
        <div className="lg:col-span-7 flex flex-col min-h-0 lg:overflow-y-auto pl-1">
          {property && (
            <PropertyWeatherPanel
              daily={daily}
              crop={cropLabels[property.crop.toLowerCase()] || property.crop}
              soil={soilLabels[property.soil_type.toLowerCase()] || property.soil_type}
            />
          )}
        </div>
      </div>
    </div>
  );
}
