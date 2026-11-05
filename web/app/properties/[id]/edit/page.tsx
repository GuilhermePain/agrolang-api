import Link from "next/link";
import { notFound } from "next/navigation";
import { getProperty, type Property } from "@/lib/api";
import PropertyEditForm from "@/components/PropertyEditForm";

export default async function EditPropertyPage({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;

  let property: Property | null = null;
  try {
    property = await getProperty(id);
  } catch {
    notFound();
  }

  if (!property) {
    notFound();
  }

  return (
    <div className="mx-auto flex w-full max-w-xl flex-1 flex-col gap-6 p-5 sm:p-8">
      <div className="rise-in">
        <Link
          href={`/properties/${property.id}`}
          className="inline-flex items-center gap-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-accent"
        >
          &larr; Voltar à ficha do talhão
        </Link>
        <p className="mt-2 font-mono text-[11px] uppercase tracking-[0.25em] text-ink-soft">
          Manejo e Atualização &middot; RF-01
        </p>
        <h1 className="font-display text-2xl italic text-ink">Editar Propriedade</h1>
      </div>

      <PropertyEditForm property={property} />
    </div>
  );
}

