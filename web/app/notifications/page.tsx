import Link from "next/link";
import { listRecentNotifications, listProducers, type NotificationLog, type Producer } from "@/lib/api";
import TriggerScanButton from "@/components/TriggerScanButton";

const statusBadges: Record<string, { label: string; class: string }> = {
  sent: { label: "Enviado", class: "bg-paper-deep text-ink border-line" },
  delivered: { label: "Entregue", class: "bg-risk-medium-bg text-risk-medium border-risk-medium" },
  read: { label: "Lido pelo Produtor", class: "bg-risk-low-bg text-risk-low border-risk-low font-bold" },
  failed: { label: "Falha no Envio", class: "bg-risk-critical-bg text-risk-critical border-risk-critical font-bold" },
};

export default async function NotificationsPage() {
  let logs: NotificationLog[] = [];
  let producers: Producer[] = [];
  let error: string | null = null;

  try {
    const [logsData, producersData] = await Promise.all([
      listRecentNotifications(),
      listProducers().catch(() => []),
    ]);
    logs = logsData;
    producers = producersData;
  } catch {
    error = "Não foi possível carregar os logs de mensagens WhatsApp.";
  }

  const producerMap = producers.reduce((acc, p) => {
    acc[p.id] = p;
    return acc;
  }, {} as Record<string, Producer>);

  return (
    <div className="mx-auto flex w-full max-w-6xl flex-1 flex-col gap-6 p-5 sm:p-8">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between border-b-2 border-ink pb-5 rise-in">
        <div>
          <Link
            href="/"
            className="inline-flex items-center gap-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-accent"
          >
            &larr; Voltar ao mapa de risco
          </Link>
          <p className="mt-2 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
            Auditoria de Notificações &middot; RF-04 &middot; WhatsApp
          </p>
          <h1 className="font-display text-3xl italic text-ink">Disparos de Alertas Climáticos</h1>
        </div>

        <TriggerScanButton />
      </div>

      {error && (
        <p className="border-2 border-risk-critical bg-risk-critical-bg px-4 py-3 font-mono text-sm text-risk-critical">
          {error}
        </p>
      )}

      {logs.length === 0 && !error ? (
        <div className="border-2 border-dashed border-line p-10 text-center font-mono text-sm text-ink-soft">
          Nenhum disparo de WhatsApp registrado ainda.
          <p className="mt-2 text-xs">
            Assim que uma condição de risco (estresse hídrico, geada, calor extremo ou chuva torrencial) for detectada nas lavouras, as mensagens enviadas aos produtores serão auditadas aqui em tempo real.
          </p>
        </div>
      ) : (
        <div className="overflow-x-auto border-2 border-ink bg-card shadow-[var(--shadow-hard)]">
          <table className="w-full min-w-[700px] border-collapse font-mono text-sm">
            <thead>
              <tr className="border-b-2 border-ink bg-paper-deep text-left text-[11px] uppercase tracking-wider text-ink-soft">
                <th className="px-4 py-3 font-semibold">Data / Hora</th>
                <th className="px-4 py-3 font-semibold">Produtor</th>
                <th className="px-4 py-3 font-semibold">Telefone</th>
                <th className="px-4 py-3 font-semibold">Conteúdo do Alerta</th>
                <th className="px-4 py-3 font-semibold">Status de Entrega</th>
              </tr>
            </thead>
            <tbody>
              {logs.map((log, i) => {
                const prod = producerMap[log.producer_id];
                const badge = statusBadges[log.status] || {
                  label: log.status,
                  class: "border-line bg-paper",
                };
                return (
                  <tr
                    key={log.id}
                    className={`border-b border-line ${i % 2 === 1 ? "bg-paper/40" : ""}`}
                  >
                    <td className="px-4 py-3 text-xs text-ink-soft whitespace-nowrap">
                      {new Date(log.sent_at).toLocaleString("pt-BR")}
                    </td>

                    <td className="px-4 py-3 font-semibold text-ink">
                      {prod ? (
                        <Link href={`/producers/${prod.id}`} className="hover:underline text-accent">
                          {prod.name}
                        </Link>
                      ) : (
                        "Produtor"
                      )}
                    </td>

                    <td className="px-4 py-3 text-xs tabular text-ink whitespace-nowrap">
                      {log.phone}
                    </td>

                    <td className="px-4 py-3 text-xs max-w-xs truncate" title={log.message}>
                      {log.message}
                    </td>

                    <td className="px-4 py-3 whitespace-nowrap">
                      <span
                        className={`inline-block border px-2 py-0.5 font-mono text-[10px] uppercase tracking-wide ${badge.class}`}
                      >
                        {badge.label}
                      </span>
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}

