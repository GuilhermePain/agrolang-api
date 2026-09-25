import Link from "next/link";
import { listProperties, getPropertyAlerts } from "@/lib/api";
import PropertyMapClient from "@/components/PropertyMapClient";
import { currentRiskLevel, riskLabels } from "@/lib/risk";

export default async function Home() {
  let items: Awaited<ReturnType<typeof loadItems>> = [];
  let error: string | null = null;

  try {
    items = await loadItems();
  } catch {
    error = "Não foi possível carregar as propriedades. Verifique se a API está no ar.";
  }

  return (
    <div className="flex flex-1 flex-col">
      <header className="border-b border-black/10 px-6 py-4 dark:border-white/10">
        <h1 className="text-xl font-semibold">AgroResiliente</h1>
        <p className="text-sm text-zinc-600 dark:text-zinc-400">
          Mapa de risco das propriedades monitoradas
        </p>
      </header>

      {error && <p className="px-6 py-4 text-red-600">{error}</p>}

      <div className="flex flex-1 flex-col gap-4 p-6 lg:flex-row">
        <div className="h-[480px] flex-1 overflow-hidden rounded-lg border border-black/10 dark:border-white/10">
          <PropertyMapClient items={items} />
        </div>

        <aside className="w-full lg:w-80">
          <h2 className="mb-2 text-sm font-semibold uppercase tracking-wide text-zinc-500">
            Propriedades
          </h2>
          <ul className="flex flex-col gap-2">
            {items.map(({ property, alerts }) => {
              const level = currentRiskLevel(alerts);
              return (
                <li key={property.id}>
                  <Link
                    href={`/properties/${property.id}`}
                    className="flex items-center justify-between rounded border border-black/10 px-3 py-2 hover:bg-black/[.03] dark:border-white/10 dark:hover:bg-white/[.06]"
                  >
                    <span>{property.crop}</span>
                    <span>{riskLabels[level]}</span>
                  </Link>
                </li>
              );
            })}
            {items.length === 0 && !error && (
              <li className="text-sm text-zinc-500">Nenhuma propriedade cadastrada.</li>
            )}
          </ul>
        </aside>
      </div>
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
