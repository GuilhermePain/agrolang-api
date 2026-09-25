"use client";

import { useEffect, useState, type FormEvent } from "react";
import Link from "next/link";
import { createProperty, listProducers, ApiError, type Producer, type CropStage } from "@/lib/api";
import { validateProperty, type PropertyFormValues, type PropertyFormErrors } from "@/lib/validation";
import { FieldShell, fieldClass } from "./FieldShell";

const CROP_STAGE_OPTIONS: { value: CropStage; label: string }[] = [
  { value: "germination", label: "Germinação" },
  { value: "flowering", label: "Floração" },
  { value: "harvest", label: "Colheita" },
];

export default function PropertyForm({ initialProducerId }: { initialProducerId?: string }) {
  const [producers, setProducers] = useState<Producer[]>([]);
  const [loadingProducers, setLoadingProducers] = useState(true);
  const [producersError, setProducersError] = useState<string | null>(null);

  const [values, setValues] = useState<PropertyFormValues>({
    producer_id: initialProducerId ?? "",
    latitude: "",
    longitude: "",
    crop: "",
    soil_type: "",
    crop_stage: "",
  });
  const [errors, setErrors] = useState<PropertyFormErrors>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [created, setCreated] = useState<string | null>(null);

  useEffect(() => {
    listProducers()
      .then(setProducers)
      .catch(() => setProducersError("Não foi possível carregar os produtores."))
      .finally(() => setLoadingProducers(false));
  }, []);

  function update<K extends keyof PropertyFormValues>(key: K, value: PropertyFormValues[K]) {
    setValues((v) => ({ ...v, [key]: value }));
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const validation = validateProperty(values);
    setErrors(validation);
    if (Object.keys(validation).length > 0) return;

    setSubmitting(true);
    setSubmitError(null);
    try {
      const property = await createProperty({
        producer_id: values.producer_id,
        latitude: Number(values.latitude),
        longitude: Number(values.longitude),
        crop: values.crop,
        soil_type: values.soil_type,
        crop_stage: values.crop_stage as CropStage,
      });
      setCreated(property.id);
    } catch (err) {
      setSubmitError(
        err instanceof ApiError ? err.message : "Não foi possível cadastrar a propriedade."
      );
    } finally {
      setSubmitting(false);
    }
  }

  if (created) {
    return (
      <div className="rise-in border-2 border-risk-low bg-risk-low-bg px-5 py-6 font-mono text-sm text-risk-low">
        <p className="font-display text-lg italic text-ink">Propriedade cadastrada.</p>
        <p className="mt-2 text-ink-soft">
          O talhão já entra na próxima varredura de risco climático.
        </p>
        <div className="mt-4 flex flex-wrap gap-3">
          <Link
            href={`/properties/${created}`}
            className="border-2 border-ink bg-accent px-4 py-2 font-semibold uppercase tracking-wide text-paper hover:bg-ink"
          >
            Ver ficha do talhão
          </Link>
          <Link href="/" className="border-2 border-ink px-4 py-2 font-semibold uppercase tracking-wide text-ink">
            Voltar ao mapa
          </Link>
        </div>
      </div>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-5 rise-in">
      <FieldShell label="Produtor" htmlFor="producer_id" error={errors.producer_id}>
        <select
          id="producer_id"
          className={fieldClass}
          value={values.producer_id}
          onChange={(e) => update("producer_id", e.target.value)}
          disabled={loadingProducers}
        >
          <option value="">
            {loadingProducers ? "Carregando..." : "Selecione um produtor"}
          </option>
          {producers.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name} &middot; {p.city}/{p.state}
            </option>
          ))}
        </select>
        {producersError && (
          <p className="font-mono text-xs text-risk-critical">
            {producersError}{" "}
            <Link href="/producers/new" className="underline">
              Cadastrar produtor
            </Link>
          </p>
        )}
        {!loadingProducers && !producersError && producers.length === 0 && (
          <p className="font-mono text-xs text-ink-soft">
            Nenhum produtor cadastrado ainda.{" "}
            <Link href="/producers/new" className="underline text-accent">
              Cadastrar o primeiro
            </Link>
          </p>
        )}
      </FieldShell>

      <div className="grid grid-cols-2 gap-4">
        <FieldShell label="Latitude" htmlFor="latitude" error={errors.latitude}>
          <input
            id="latitude"
            className={`${fieldClass} tabular`}
            value={values.latitude}
            onChange={(e) => update("latitude", e.target.value)}
            placeholder="-22.9099"
            inputMode="decimal"
          />
        </FieldShell>

        <FieldShell label="Longitude" htmlFor="longitude" error={errors.longitude}>
          <input
            id="longitude"
            className={`${fieldClass} tabular`}
            value={values.longitude}
            onChange={(e) => update("longitude", e.target.value)}
            placeholder="-47.0626"
            inputMode="decimal"
          />
        </FieldShell>
      </div>

      <FieldShell label="Cultura" htmlFor="crop" error={errors.crop}>
        <input
          id="crop"
          className={fieldClass}
          value={values.crop}
          onChange={(e) => update("crop", e.target.value)}
          placeholder="Ex: Milho, Tomate, Café"
        />
      </FieldShell>

      <div className="grid grid-cols-2 gap-4">
        <FieldShell label="Tipo de solo" htmlFor="soil_type" error={errors.soil_type}>
          <input
            id="soil_type"
            className={fieldClass}
            value={values.soil_type}
            onChange={(e) => update("soil_type", e.target.value)}
            placeholder="Ex: Argiloso"
          />
        </FieldShell>

        <FieldShell label="Fase do cultivo" htmlFor="crop_stage" error={errors.crop_stage}>
          <select
            id="crop_stage"
            className={fieldClass}
            value={values.crop_stage}
            onChange={(e) => update("crop_stage", e.target.value)}
          >
            <option value="">Selecione</option>
            {CROP_STAGE_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
        </FieldShell>
      </div>

      {submitError && (
        <p className="border-2 border-risk-critical bg-risk-critical-bg px-3 py-2 font-mono text-sm text-risk-critical">
          {submitError}
        </p>
      )}

      <button
        type="submit"
        disabled={submitting}
        className="w-fit border-2 border-ink bg-accent px-5 py-2.5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink disabled:opacity-50"
      >
        {submitting ? "Cadastrando..." : "Cadastrar propriedade"}
      </button>
    </form>
  );
}
