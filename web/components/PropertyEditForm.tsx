"use client";

import { useEffect, useState, type FormEvent } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { updateProperty, listProducers, ApiError, type Property, type Producer, type CropStage } from "@/lib/api";
import { validateProperty, type PropertyFormValues, type PropertyFormErrors } from "@/lib/validation";
import { FieldShell, fieldClass } from "./FieldShell";
import LocationPickerClient from "./LocationPickerClient";
import { GpsIcon } from "./Icons";

const CROP_STAGE_OPTIONS: { value: CropStage; label: string }[] = [
  { value: "germination", label: "Germinação" },
  { value: "flowering", label: "Floração" },
  { value: "harvest", label: "Colheita" },
];

const SOIL_TYPE_OPTIONS = [
  { value: "medio", label: "Médio / Franco (Equilibrado, drenagem regular)" },
  { value: "argiloso", label: "Argiloso (Alta retenção de água)" },
  { value: "arenoso", label: "Arenoso (Baixa retenção, seca rápida)" },
  { value: "raso", label: "Raso / Vertissolo (Semiárido / Pedregoso, satura e desseca rápido)" },
  { value: "aluvial", label: "Aluvial / Várzea (Fértil, úmido em margens fluviais)" },
];

const CROP_OPTIONS = [
  { value: "milho", label: "Milho" },
  { value: "feijao", label: "Feijão / Caupi" },
  { value: "mandioca", label: "Mandioca / Macaxeira" },
  { value: "caju", label: "Caju / Cajueiro" },
  { value: "tomate", label: "Tomate" },
  { value: "cafe", label: "Café" },
  { value: "fruticultura", label: "Fruticultura Tropical (Melão, Melancia, Manga)" },
  { value: "soja", label: "Soja" },
  { value: "algodao", label: "Algodão" },
  { value: "cana", label: "Cana-de-Açúcar" },
  { value: "hortalicas", label: "Hortaliças / Olericultura" },
];

export default function PropertyEditForm({ property }: { property: Property }) {
  const router = useRouter();
  const [producers, setProducers] = useState<Producer[]>([]);
  const [loadingProducers, setLoadingProducers] = useState(true);

  const [values, setValues] = useState<PropertyFormValues>({
    producer_id: property.producer_id,
    latitude: property.latitude.toString(),
    longitude: property.longitude.toString(),
    crop: property.crop.toLowerCase(),
    soil_type: property.soil_type.toLowerCase(),
    crop_stage: property.crop_stage,
  });
  const [errors, setErrors] = useState<PropertyFormErrors>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  // Estados de GPS e busca
  const [gettingLocation, setGettingLocation] = useState(false);
  const [locationStatus, setLocationStatus] = useState<{
    type: "success" | "error" | "info";
    text: string;
  } | null>(null);

  useEffect(() => {
    listProducers()
      .then(setProducers)
      .finally(() => setLoadingProducers(false));
  }, []);

  function update<K extends keyof PropertyFormValues>(key: K, value: PropertyFormValues[K]) {
    setValues((v) => ({ ...v, [key]: value }));
  }

  function handleMapPick(lat: number, lon: number) {
    update("latitude", lat.toString());
    update("longitude", lon.toString());
    setLocationStatus({
      type: "info",
      text: `Ponto selecionado no mapa: ${lat.toFixed(5)}, ${lon.toFixed(5)}`,
    });
    setErrors((prev) => ({ ...prev, latitude: undefined, longitude: undefined }));
  }

  function handleGetCurrentLocation() {
    if (!navigator.geolocation) {
      setLocationStatus({
        type: "error",
        text: "Geolocalização não suportada pelo navegador.",
      });
      return;
    }

    setGettingLocation(true);
    setLocationStatus({ type: "info", text: "Acessando GPS..." });

    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const lat = Number(pos.coords.latitude.toFixed(6));
        const lon = Number(pos.coords.longitude.toFixed(6));
        update("latitude", lat.toString());
        update("longitude", lon.toString());
        setGettingLocation(false);
        setLocationStatus({
          type: "success",
          text: `Localização GPS capturada (precisão: ~${Math.round(pos.coords.accuracy)}m)`,
        });
        setErrors((prev) => ({ ...prev, latitude: undefined, longitude: undefined }));
      },
      () => {
        setGettingLocation(false);
        setLocationStatus({ type: "error", text: "Falha ao obter sinal de GPS." });
      },
      { enableHighAccuracy: true, timeout: 15000 }
    );
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const validation = validateProperty(values);
    setErrors(validation);
    if (Object.keys(validation).length > 0) return;

    setSubmitting(true);
    setSubmitError(null);
    try {
      await updateProperty(property.id, {
        producer_id: values.producer_id,
        latitude: Number(values.latitude),
        longitude: Number(values.longitude),
        crop: values.crop,
        soil_type: values.soil_type,
        crop_stage: values.crop_stage as CropStage,
      });
      router.push(`/properties/${property.id}`);
      router.refresh();
    } catch (err) {
      setSubmitError(
        err instanceof ApiError ? err.message : "Não foi possível atualizar o talhão."
      );
    } finally {
      setSubmitting(false);
    }
  }

  const parsedLat = parseFloat(values.latitude);
  const parsedLon = parseFloat(values.longitude);
  const validCoords = !isNaN(parsedLat) && !isNaN(parsedLon);

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6 rise-in">
      <FieldShell label="Produtor Responsável" htmlFor="producer_id" error={errors.producer_id}>
        <select
          id="producer_id"
          className={fieldClass}
          value={values.producer_id}
          onChange={(e) => update("producer_id", e.target.value)}
          disabled={loadingProducers}
        >
          {producers.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name} &middot; {p.city}/{p.state}
            </option>
          ))}
        </select>
      </FieldShell>

      {/* LOCALIZAÇÃO E MAPA */}
      <div className="flex flex-col gap-3 rounded border-2 border-ink/20 bg-paper-deep/40 p-4">
        <div className="flex items-center justify-between">
          <span className="font-mono text-xs font-semibold uppercase tracking-wide text-ink">
            Localização e Coordenadas
          </span>
          <button
            type="button"
            onClick={handleGetCurrentLocation}
            disabled={gettingLocation}
            className="inline-flex items-center gap-1.5 border border-ink bg-paper px-2.5 py-1 font-mono text-xs font-semibold uppercase text-ink hover:bg-accent hover:text-paper"
          >
            <GpsIcon className="h-3.5 w-3.5 shrink-0 text-accent" />
            <span>{gettingLocation ? "Buscando GPS..." : "Usar GPS Atual"}</span>
          </button>
        </div>

        {locationStatus && (
          <p className="font-mono text-xs text-ink">{locationStatus.text}</p>
        )}

        <div className="overflow-hidden rounded border border-ink/40">
          <LocationPickerClient
            latitude={validCoords ? parsedLat : null}
            longitude={validCoords ? parsedLon : null}
            onChange={handleMapPick}
          />
        </div>

        <div className="grid grid-cols-2 gap-3 pt-1">
          <FieldShell label="Latitude" htmlFor="latitude" error={errors.latitude}>
            <input
              id="latitude"
              className={`${fieldClass} tabular text-xs`}
              value={values.latitude}
              onChange={(e) => update("latitude", e.target.value)}
              inputMode="decimal"
            />
          </FieldShell>

          <FieldShell label="Longitude" htmlFor="longitude" error={errors.longitude}>
            <input
              id="longitude"
              className={`${fieldClass} tabular text-xs`}
              value={values.longitude}
              onChange={(e) => update("longitude", e.target.value)}
              inputMode="decimal"
            />
          </FieldShell>
        </div>
      </div>

      {/* DADOS DA LAVOURA */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <FieldShell label="Cultura" htmlFor="crop" error={errors.crop}>
          <select
            id="crop"
            className={fieldClass}
            value={values.crop}
            onChange={(e) => update("crop", e.target.value)}
          >
            {CROP_OPTIONS.map((c) => (
              <option key={c.value} value={c.value}>
                {c.label}
              </option>
            ))}
          </select>
        </FieldShell>

        <FieldShell label="Tipo de Solo" htmlFor="soil_type" error={errors.soil_type}>
          <select
            id="soil_type"
            className={fieldClass}
            value={values.soil_type}
            onChange={(e) => update("soil_type", e.target.value)}
          >
            {SOIL_TYPE_OPTIONS.map((s) => (
              <option key={s.value} value={s.value}>
                {s.label}
              </option>
            ))}
          </select>
        </FieldShell>

        <FieldShell label="Fase do Cultivo" htmlFor="crop_stage" error={errors.crop_stage}>
          <select
            id="crop_stage"
            className={fieldClass}
            value={values.crop_stage}
            onChange={(e) => update("crop_stage", e.target.value)}
          >
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

      <div className="flex items-center gap-3 pt-2">
        <button
          type="submit"
          disabled={submitting}
          className="border-2 border-ink bg-accent px-5 py-2.5 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink disabled:opacity-50"
        >
          {submitting ? "Salvando..." : "Salvar Alterações"}
        </button>

        <Link
          href={`/properties/${property.id}`}
          className="border border-line px-4 py-2 font-mono text-xs font-semibold uppercase text-ink hover:border-ink"
        >
          Cancelar
        </Link>
      </div>
    </form>
  );
}

