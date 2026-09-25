import Link from "next/link";
import ProducerForm from "@/components/ProducerForm";

export default function NewProducerPage() {
  return (
    <div className="mx-auto flex w-full max-w-xl flex-1 flex-col gap-6 p-5 sm:p-8">
      <div className="rise-in">
        <Link
          href="/"
          className="inline-flex items-center gap-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-accent"
        >
          &larr; Voltar ao mapa
        </Link>
        <p className="mt-2 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
          Novo cadastro &middot; RF-01
        </p>
        <h1 className="font-display text-2xl italic text-ink">Cadastrar produtor</h1>
      </div>

      <ProducerForm />
    </div>
  );
}
