import Link from "next/link";
import { listProperties, getPropertyAlerts } from "@/lib/api";
import PropertyMapClient from "@/components/PropertyMapClient";
import { currentRiskLevel, riskLabels } from "@/lib/risk";

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

export default async function Home() {
  let items: Awaited<ReturnType<typeof loadItems>> = [];
  let error: string | null = null;

  try {
    items = await loadItems();
  } catch {
    error = "Não foi possível carregar as propriedades. Verifique se a API está no ar.";
  }

  const counts = items.reduce(
    (acc, { alerts }) => {
      acc[currentRiskLevel(alerts)] += 1;
      return acc;
    },
    { low: 0, medium: 0, critical: 0 }
  );

  return (
    <div className="flex flex-1 flex-col">
      <header className="relative overflow-hidden border-b-2 border-ink px-6 py-5 sm:px-8">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div className="rise-in">
            <p className="mb-1 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
              Estação de campo &middot; RF-05
            </p>
            <h1 className="font-display text-3xl italic tracking-tight text-ink sm:text-4xl">
              AgroResiliente
            </h1>
            <p className="mt-1 max-w-md text-sm text-ink-soft">
              Leitura do risco climático de cada talhão monitorado, em tempo quase real.
            </p>
          </div>

          <div className="flex gap-2 rise-in" style={{ animationDelay: "80ms" }}>
            <CountPill label="Baixo" count={counts.low} tone="low" />
            <CountPill label="Médio" count={counts.medium} tone="medium" />
            <CountPill label="Crítico" count={counts.critical} tone="critical" />
          </div>
        </div>
      </header>

      {error && (
        <p className="border-b-2 border-risk-critical bg-risk-critical-bg px-6 py-3 text-sm font-medium text-risk-critical sm:px-8">
          {error}
        </p>
      )}

      <div className="flex flex-1 flex-col gap-5 p-5 lg:flex-row lg:p-8">
        <div
          className="h-[420px] shrink-0 overflow-hidden border-2 border-ink shadow-[var(--shadow-hard)] lg:h-auto lg:flex-1 lg:shrink rise-in"
          style={{ animationDelay: "140ms" }}
        >
          <PropertyMapClient items={items} />
        </div>

        <aside className="flex w-full flex-col lg:w-96">
          <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
            <h2 className="font-display text-lg italic text-ink">Talhões</h2>
            <div className="flex gap-2">
              <Link
                href="/producers/new"
                className="border-2 border-ink px-3 py-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-ink transition hover:bg-ink hover:text-paper"
              >
                + Produtor
              </Link>
              <Link
                href="/properties/new"
                className="border-2 border-ink bg-accent px-3 py-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-paper transition hover:bg-ink"
              >
                + Propriedade
              </Link>
            </div>
          </div>

          <ul className="flex flex-1 flex-col gap-2.5 overflow-y-auto">
            {items.map(({ property, alerts }, i) => {
              const level = currentRiskLevel(alerts);
              return (
                <li
                  key={property.id}
                  className="rise-in"
                  style={{ animationDelay: `${180 + i * 40}ms` }}
                >
                  <Link
                    href={`/properties/${property.id}`}
                    className="group flex items-center justify-between gap-3 border-2 border-line bg-card px-4 py-3 transition hover:border-ink hover:-translate-y-0.5 hover:shadow-[var(--shadow-hard)]"
                  >
                    <div className="min-w-0">
                      <p className="truncate font-display text-base text-ink">{property.crop}</p>
                      <p className="truncate font-mono text-[11px] uppercase tracking-wide text-ink-soft">
                        {cropStageLabels[property.crop_stage] ?? property.crop_stage} &middot;{" "}
                        {property.soil_type}
                      </p>
                    </div>
                    <span
                      className={`shrink-0 border-2 px-2 py-1 font-mono text-[11px] font-semibold uppercase tracking-wide ${riskBadgeClass[level]}`}
                    >
                      {riskLabels[level]}
                    </span>
                  </Link>
                </li>
              );
            })}
            {items.length === 0 && !error && (
              <li className="border-2 border-dashed border-line px-4 py-6 text-center font-mono text-sm text-ink-soft">
                Nenhuma propriedade cadastrada.
                <br />
                <Link href="/producers/new" className="underline decoration-dashed underline-offset-4">
                  Cadastrar a primeira
                </Link>
              </li>
            )}
          </ul>
        </aside>
      </div>
    </div>
  );
}

function CountPill({ label, count, tone }: { label: string; count: number; tone: string }) {
  return (
    <div className={`border-2 px-3 py-1.5 text-center font-mono ${riskBadgeClass[tone]}`}>
      <div className="text-lg font-semibold leading-none tabular">{count}</div>
      <div className="text-[10px] uppercase tracking-wide leading-none mt-1">{label}</div>
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
