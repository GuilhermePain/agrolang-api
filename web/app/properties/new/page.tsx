import Link from "next/link";
import PropertyForm from "@/components/PropertyForm";

export default async function NewPropertyPage({
  searchParams,
}: {
  searchParams: Promise<{ producer_id?: string }>;
}) {
  const { producer_id } = await searchParams;

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
        <h1 className="font-display text-2xl italic text-ink">Cadastrar propriedade</h1>
      </div>

      <PropertyForm initialProducerId={producer_id} />
    </div>
  );
}
