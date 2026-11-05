import Link from "next/link";
import { notFound } from "next/navigation";
import { getProducer, type Producer } from "@/lib/api";
import ProducerEditForm from "@/components/ProducerEditForm";

export default async function EditProducerPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let producer: Producer | null = null;
  try {
    producer = await getProducer(id);
  } catch {
    notFound();
  }

  if (!producer) {
    notFound();
  }

  return (
    <div className="mx-auto flex w-full max-w-xl flex-1 flex-col gap-6 p-5 sm:p-8">
      <div className="rise-in">
        <Link
          href={`/producers/${producer.id}`}
          className="inline-flex items-center gap-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-accent"
        >
          &larr; Voltar à ficha do produtor
        </Link>
        <p className="mt-2 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
          Atualização Cadastral &middot; RF-01
        </p>
        <h1 className="font-display text-2xl italic text-ink">Editar Produtor</h1>
      </div>

      <ProducerEditForm producer={producer} />
    </div>
  );
}

