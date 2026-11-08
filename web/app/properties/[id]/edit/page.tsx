import Link from "next/link";
import { notFound, unstable_rethrow } from "next/navigation";
import { ApiError, getProperty, listProducers, type Property, type Producer } from "@/lib/api";
import PropertyEditForm from "@/components/PropertyEditForm";

export default async function EditPropertyPage({ params }: PageProps<"/properties/[id]/edit">) {
  const { id } = await params;

  let property: Property | null = null;
  let producers: Producer[] = [];
  try {
    [property, producers] = await Promise.all([getProperty(id), listProducers()]);
  } catch (err) {
    unstable_rethrow(err);
    if (err instanceof ApiError && err.status === 404) notFound();
    throw err;
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

      <PropertyEditForm property={property} producers={producers} />
    </div>
  );
}

