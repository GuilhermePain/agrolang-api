import Link from "next/link";
import { notFound } from "next/navigation";
import {
  getProducer,
  listProducerProperties,
  listProducerNotifications,
  type Producer,
  type Property,
  type NotificationLog,
} from "@/lib/api";
import DeleteProducerAction from "@/components/DeleteProducerAction";
import { MapPinIcon, ChatIcon, DoubleCheckIcon } from "@/components/Icons";
import { cropLabels, soilLabels } from "@/lib/risk";

const cropStageLabels: Record<string, string> = {
  germination: "Germinação",
  flowering: "Floração",
  harvest: "Colheita",
};

const statusBadges: Record<string, { label: string; class: string }> = {
  sent: { label: "Enviado", class: "bg-paper-deep text-ink border-line" },
  delivered: { label: "Entregue", class: "bg-risk-medium-bg text-risk-medium border-risk-medium" },
  read: { label: "Lido pelo Produtor", class: "bg-risk-low-bg text-risk-low border-risk-low font-bold" },
  failed: { label: "Falha", class: "bg-risk-critical-bg text-risk-critical border-risk-critical font-bold" },
};

export default async function ProducerDetailPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let producer: Producer | null = null;
  let properties: Property[] = [];
  let notifications: NotificationLog[] = [];
  let error: string | null = null;

  try {
    const [producerData, propertiesData, notificationsData] = await Promise.all([
      getProducer(id),
      listProducerProperties(id).catch(() => []),
      listProducerNotifications(id).catch(() => []),
    ]);
    producer = producerData;
    properties = propertiesData;
    notifications = notificationsData;
  } catch {
    error = "Não foi possível carregar os dados deste produtor.";
  }

  if (!producer && !error) {
    notFound();
  }

  const cleanPhone = producer ? producer.whatsapp_phone.replace(/\D/g, "") : "";

  return (
    <div className="mx-auto flex w-full max-w-4xl flex-1 flex-col gap-8 p-5 sm:p-8">
      <div className="rise-in">
        <Link
          href="/producers"
          className="inline-flex items-center gap-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-accent"
        >
          &larr; Voltar à lista de produtores
        </Link>
        <p className="mt-2 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
          Ficha do Cooperado &middot; RF-01
        </p>
      </div>

      {error && (
        <p className="border-2 border-risk-critical bg-risk-critical-bg px-4 py-3 font-mono text-sm text-risk-critical">
          {error}
        </p>
      )}

      {producer && (
        <>
          {/* CARTÃO DO PRODUTOR */}
          <div className="flex flex-col justify-between gap-4 border-2 border-ink bg-card p-6 shadow-[var(--shadow-hard)] sm:flex-row sm:items-center">
            <div>
              <h1 className="font-display text-3xl font-bold italic text-ink">
                {producer.name}
              </h1>
              <p className="mt-1 flex items-center gap-1.5 font-mono text-sm text-ink-soft">
                <MapPinIcon className="h-4 w-4 text-accent shrink-0" />
                <span>{producer.city}, {producer.state} &middot; Cadastrado em:{" "}
                {new Date(producer.created_at).toLocaleDateString("pt-BR")}</span>
              </p>
              <div className="mt-3 flex items-center gap-2">
                <span className="font-mono text-xs text-ink-soft">WhatsApp para alertas:</span>
                <a
                  href={`https://wa.me/${cleanPhone}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 rounded border border-ink bg-paper px-2.5 py-1 font-mono text-xs font-semibold text-ink hover:bg-ink hover:text-paper"
                >
                  <ChatIcon className="h-3.5 w-3.5 text-accent" />
                  <span>{producer.whatsapp_phone} (Abrir conversa)</span>
                </a>
              </div>
            </div>

            <div className="flex flex-wrap items-center gap-3">
              <Link
                href={`/producers/${producer.id}/edit`}
                className="border-2 border-ink bg-paper px-4 py-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-ink hover:bg-ink hover:text-paper"
              >
                Editar Dados
              </Link>
              <DeleteProducerAction producerId={producer.id} />
            </div>
          </div>

          {/* LISTA DE TALHÕES / PROPRIEDADES */}
          <section className="flex flex-col gap-4">
            <div className="flex items-center justify-between">
              <h2 className="font-display text-2xl italic text-ink">
                Talhões Monitorados ({properties.length})
              </h2>
              <Link
                href={`/properties/new?producer_id=${producer.id}`}
                className="border-2 border-ink bg-accent px-3 py-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-paper hover:bg-ink"
              >
                + Adicionar Talhão
              </Link>
            </div>

            {properties.length === 0 ? (
              <div className="border-2 border-dashed border-line p-6 text-center font-mono text-sm text-ink-soft">
                Este produtor ainda não possui nenhuma propriedade cadastrada.
                <br />
                <Link
                  href={`/properties/new?producer_id=${producer.id}`}
                  className="mt-2 inline-block font-semibold text-accent underline"
                >
                  Cadastrar primeiro talhão
                </Link>
              </div>
            ) : (
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                {properties.map((prop) => (
                  <div
                    key={prop.id}
                    className="flex flex-col justify-between border-2 border-line bg-card p-4 transition hover:border-ink hover:shadow-[var(--shadow-hard)]"
                  >
                    <div>
                      <div className="flex items-center justify-between">
                        <span className="font-display text-lg font-bold italic text-ink">
                          {cropLabels[prop.crop.toLowerCase()] || prop.crop}
                        </span>
                        <span className="border border-line bg-paper-deep px-2 py-0.5 font-mono text-[10px] uppercase text-ink">
                          {cropStageLabels[prop.crop_stage] ?? prop.crop_stage}
                        </span>
                      </div>
                      <p className="mt-1 font-mono text-xs text-ink-soft">
                        Solo: {soilLabels[prop.soil_type.toLowerCase()] || prop.soil_type} &middot; Coordenadas: {prop.latitude.toFixed(4)},{" "}
                        {prop.longitude.toFixed(4)}
                      </p>
                    </div>

                    <div className="mt-4 flex items-center justify-between border-t border-line pt-3 font-mono text-xs">
                      <Link
                        href={`/properties/${prop.id}`}
                        className="font-semibold text-accent hover:underline"
                      >
                        Ver Monitoramento &rarr;
                      </Link>
                      <Link
                        href={`/properties/${prop.id}/edit`}
                        className="text-ink-soft hover:text-ink"
                      >
                        Editar
                      </Link>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </section>

          {/* HISTÓRICO DE MENSAGENS WHATSAPP ENVIADAS */}
          <section className="flex flex-col gap-4">
            <h2 className="font-display text-2xl italic text-ink">
              Histórico de Mensagens WhatsApp ({notifications.length})
            </h2>

            {notifications.length === 0 ? (
              <p className="border-2 border-dashed border-line p-6 text-center font-mono text-sm text-ink-soft">
                Nenhum alerta foi disparado para o WhatsApp deste produtor até o momento.
              </p>
            ) : (
              <div className="flex flex-col gap-3">
                {notifications.map((notif) => {
                  const badge = statusBadges[notif.status] || {
                    label: notif.status,
                    class: "border-line bg-paper",
                  };
                  return (
                    <div
                      key={notif.id}
                      className="border-2 border-line bg-card p-4 flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between"
                    >
                      <div>
                        <div className="flex items-center gap-2">
                          <span className="font-mono text-xs font-semibold text-ink">
                            Enviado em: {new Date(notif.sent_at).toLocaleString("pt-BR")}
                          </span>
                          <span
                            className={`border px-2 py-0.5 font-mono text-[10px] uppercase tracking-wide ${badge.class}`}
                          >
                            {badge.label}
                          </span>
                        </div>
                        <p className="mt-2 whitespace-pre-line font-mono text-xs text-ink/90 bg-paper-deep/30 p-2.5 rounded border border-line">
                          {notif.message}
                        </p>
                      </div>

                      {notif.read_at && (
                        <span className="inline-flex items-center gap-1 font-mono text-[11px] text-risk-low shrink-0 font-semibold">
                          <DoubleCheckIcon className="h-3.5 w-3.5 text-risk-low" />
                          <span>Lido em {new Date(notif.read_at).toLocaleTimeString("pt-BR")}</span>
                        </span>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </section>
        </>
      )}
    </div>
  );
}
