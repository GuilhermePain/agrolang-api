import Link from "next/link";
import { unstable_rethrow } from "next/navigation";
import { orFallback } from "@/lib/fallback";
import { listProducers, listProperties, type Producer, type Property } from "@/lib/api";
import { MapPinIcon, PhoneIcon } from "@/components/Icons";

export default async function ProducersPage() {
  let producers: Producer[] = [];
  let properties: Property[] = [];
  let error: string | null = null;

  try {
    const [producersData, propertiesData] = await Promise.all([
      listProducers(),
      orFallback(listProperties(), [] as Property[]),
    ]);
    producers = producersData;
    properties = propertiesData;
  } catch (err) {
    unstable_rethrow(err);
    error = "Não foi possível carregar a lista de produtores.";
  }

  const propertyCountByProducer = properties.reduce((acc, p) => {
    acc[p.producer_id] = (acc[p.producer_id] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-1 flex-col h-full min-h-0 overflow-y-auto lg:overflow-hidden p-4 sm:p-6">
      <div className="shrink-0 mb-4 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between border-b-2 border-ink pb-3 rise-in">
        <div>
          <Link
            href="/"
            className="inline-flex items-center gap-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-accent touch"
          >
            &larr; Voltar ao mapa de risco
          </Link>
          <p className="mt-1 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
            Gestão de Cooperados &middot; RF-01
          </p>
          <h1 className="font-display text-2xl font-bold italic text-ink sm:text-3xl">Produtores Cadastrados</h1>
        </div>

        <Link
          href="/producers/new"
          className="touch w-fit border-2 border-ink bg-accent px-3 py-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-paper transition hover:bg-ink"
        >
          + Cadastrar Produtor
        </Link>
      </div>

      {error && (
        <p className="shrink-0 mb-3 border-2 border-risk-critical bg-risk-critical-bg px-4 py-2 font-mono text-xs text-risk-critical">
          {error}
        </p>
      )}

      {producers.length === 0 && !error ? (
        <div className="border-2 border-dashed border-line p-8 text-center font-mono text-sm text-ink-soft">
          Nenhum produtor rural cadastrado até o momento.
          <br />
          <Link href="/producers/new" className="mt-3 inline-block underline text-accent font-semibold">
            Cadastrar o primeiro produtor
          </Link>
        </div>
      ) : (
        <div className="flex-1 overflow-y-auto pr-1">
          <div className="grid grid-cols-1 gap-3.5 sm:grid-cols-2 lg:grid-cols-3">
          {producers.map((producer, idx) => {
            const count = propertyCountByProducer[producer.id] || 0;
            const cleanPhone = producer.whatsapp_phone.replace(/\D/g, "");
            return (
              <div
                key={producer.id}
                className="rise-in flex flex-col justify-between border-2 border-ink bg-card p-5 shadow-[var(--shadow-hard)]"
                style={{ animationDelay: `${idx * 40}ms` }}
              >
                <div>
                  <div className="flex items-start justify-between gap-2">
                    <h2 className="font-display text-lg font-bold italic text-ink truncate">
                      {producer.name}
                    </h2>
                    <span className="shrink-0 border border-line bg-paper-deep px-2 py-0.5 font-mono text-[10px] font-semibold uppercase text-ink">
                      {count} {count === 1 ? "talhão" : "talhões"}
                    </span>
                  </div>

                  <p className="mt-1 flex items-center gap-1 font-mono text-xs text-ink-soft">
                    <MapPinIcon className="h-3.5 w-3.5 text-accent shrink-0" />
                    <span>{producer.city}/{producer.state}</span>
                  </p>

                  <div className="mt-2 flex items-center gap-1.5 font-mono text-xs text-ink">
                    <PhoneIcon className="h-3.5 w-3.5 text-ink-soft shrink-0" />
                    <a
                      href={`https://wa.me/${cleanPhone}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-accent underline decoration-dashed hover:text-ink"
                    >
                      {producer.whatsapp_phone}
                    </a>
                  </div>
                </div>

                <div className="mt-5 flex flex-wrap items-center gap-2 border-t border-line pt-3 font-mono text-xs">
                  <Link
                    href={`/producers/${producer.id}`}
                    className="touch border border-ink bg-paper px-2.5 py-1 font-semibold uppercase text-ink hover:bg-ink hover:text-paper"
                  >
                    Ver Ficha
                  </Link>
                  <Link
                    href={`/producers/${producer.id}/edit`}
                    className="touch border border-line px-2.5 py-1 uppercase text-ink hover:border-ink"
                  >
                    Editar
                  </Link>
                  <Link
                    href={`/properties/new?producer_id=${producer.id}`}
                    className="touch border border-line px-2.5 py-1 uppercase text-accent hover:border-accent font-semibold"
                  >
                    + Talhão
                  </Link>
                </div>
              </div>
            );
          })}
          </div>
        </div>
      )}
    </div>
  );
}

