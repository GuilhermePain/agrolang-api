"use client";

import { useState, type FormEvent } from "react";
import Link from "next/link";
import { createPropertyAction } from "@/app/actions/data";
import type { CropStage, Producer } from "@/lib/types";
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

export default function PropertyForm({
  initialProducerId,
  producers,
}: {
  initialProducerId?: string;
  producers: Producer[];
}) {
  const [values, setValues] = useState<PropertyFormValues>({
    producer_id: initialProducerId ?? producers[0]?.id ?? "",
    latitude: "",
    longitude: "",
    crop: "milho",
    soil_type: "medio",
    crop_stage: "flowering",
  });
  const [errors, setErrors] = useState<PropertyFormErrors>({});
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [created, setCreated] = useState<string | null>(null);

  // Estados da geolocalização e busca
  const [gettingLocation, setGettingLocation] = useState(false);
  const [locationStatus, setLocationStatus] = useState<{
    type: "success" | "error" | "info";
    text: string;
  } | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchingLocation, setSearchingLocation] = useState(false);

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
        text: "Geolocalização não é suportada pelo seu navegador.",
      });
      return;
    }

    setGettingLocation(true);
    setLocationStatus({
      type: "info",
      text: "Acessando GPS do dispositivo...",
    });

    navigator.geolocation.getCurrentPosition(
      (pos) => {
        const lat = Number(pos.coords.latitude.toFixed(6));
        const lon = Number(pos.coords.longitude.toFixed(6));
        update("latitude", lat.toString());
        update("longitude", lon.toString());
        setGettingLocation(false);
        const accuracy = Math.round(pos.coords.accuracy);
        setLocationStatus({
          type: "success",
          text: `Localização obtida com sucesso! Precisão de ~${accuracy}m.`,
        });
        setErrors((prev) => ({ ...prev, latitude: undefined, longitude: undefined }));
      },
      (err) => {
        setGettingLocation(false);
        let msg = "Não foi possível obter a localização.";
        if (err.code === err.PERMISSION_DENIED) {
          msg = "Permissão de localização negada pelo usuário.";
        } else if (err.code === err.POSITION_UNAVAILABLE) {
          msg = "Sinal de GPS/localização indisponível no momento.";
        } else if (err.code === err.TIMEOUT) {
          msg = "Tempo esgotado ao buscar sinal de GPS.";
        }
        setLocationStatus({ type: "error", text: msg });
      },
      {
        enableHighAccuracy: true,
        timeout: 15000,
        maximumAge: 10000,
      }
    );
  }

  async function handleSearchCityOrZip() {
    if (!searchQuery.trim()) return;
    setSearchingLocation(true);
    setLocationStatus(null);

    try {
      // Se for CEP brasileiro (8 dígitos com ou sem hífen)
      const cleanCep = searchQuery.replace(/\D/g, "");
      let query = searchQuery;
      if (cleanCep.length === 8) {
        const viacepRes = await fetch(`https://viacep.com.br/ws/${cleanCep}/json/`);
        if (viacepRes.ok) {
          const viaData = await viacepRes.json();
          if (!viaData.erro) {
            query = `${viaData.logradouro || ""}, ${viaData.localidade}, ${viaData.uf}, Brasil`;
          }
        }
      }

      // Nominatim / OpenStreetMap Geocoding
      const res = await fetch(
        `https://nominatim.openstreetmap.org/search?format=json&q=${encodeURIComponent(query)}&limit=1`
      );
      if (!res.ok) throw new Error("Erro na busca de coordenadas");
      const results = await res.json();
      if (results && results.length > 0) {
        const lat = Number(parseFloat(results[0].lat).toFixed(6));
        const lon = Number(parseFloat(results[0].lon).toFixed(6));
        update("latitude", lat.toString());
        update("longitude", lon.toString());
        setLocationStatus({
          type: "success",
          text: `Encontrado: ${results[0].display_name.slice(0, 70)}...`,
        });
        setErrors((prev) => ({ ...prev, latitude: undefined, longitude: undefined }));
      } else {
        setLocationStatus({
          type: "error",
          text: "Endereço ou CEP não localizado. Tente clicar no mapa ou usar o GPS.",
        });
      }
    } catch {
      setLocationStatus({
        type: "error",
        text: "Falha ao consultar serviço de mapas.",
      });
    } finally {
      setSearchingLocation(false);
    }
  }

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    const validation = validateProperty(values);
    setErrors(validation);
    if (Object.keys(validation).length > 0) return;

    setSubmitting(true);
    setSubmitError(null);
    try {
      const result = await createPropertyAction({
        producer_id: values.producer_id,
        latitude: Number(values.latitude),
        longitude: Number(values.longitude),
        crop: values.crop,
        soil_type: values.soil_type,
        crop_stage: values.crop_stage as CropStage,
      });
      if (result.ok) setCreated(result.data.id);
      else setSubmitError(result.error);
    } catch {
      setSubmitError("Não foi possível cadastrar a propriedade.");
    } finally {
      setSubmitting(false);
    }
  }

  if (created) {
    return (
      <div className="rise-in border-2 border-risk-low bg-risk-low-bg px-5 py-6 font-mono text-sm text-risk-low">
        <p className="font-display text-lg italic text-ink">Propriedade cadastrada com sucesso!</p>
        <p className="mt-2 text-ink-soft">
          O talhão já foi registrado e entrará na próxima varredura de risco climático.
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

  const parsedLat = parseFloat(values.latitude);
  const parsedLon = parseFloat(values.longitude);
  const validCoords = !isNaN(parsedLat) && !isNaN(parsedLon);

  return (
    <form onSubmit={handleSubmit} className="flex flex-col gap-6 rise-in">
      {/* Seleção do Produtor */}
      <FieldShell label="Produtor Responsável" htmlFor="producer_id" error={errors.producer_id}>
        <select
          id="producer_id"
          className={fieldClass}
          value={values.producer_id}
          onChange={(e) => update("producer_id", e.target.value)}
        >
          <option value="">Selecione um produtor</option>
          {producers.map((p) => (
            <option key={p.id} value={p.id}>
              {p.name} &middot; {p.city}/{p.state}
            </option>
          ))}
        </select>
        {producers.length === 0 && (
          <p className="font-mono text-xs text-ink-soft">
            Nenhum produtor cadastrado ainda.{" "}
            <Link href="/producers/new" className="underline text-accent">
              Cadastrar o primeiro
            </Link>
          </p>
        )}
      </FieldShell>

      {/* BLOCO DE LOCALIZAÇÃO INTELIGENTE (GPS + MAPA + BUSCA) */}
      <div className="flex flex-col gap-3 rounded border-2 border-ink/20 bg-paper-deep/40 p-4">
        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <span className="font-mono text-xs font-semibold uppercase tracking-wide text-ink">
              Localização do Talhão
            </span>
            <p className="text-xs text-ink-soft">
              Capture pelo GPS, clique no mapa ou digite cidade/CEP.
            </p>
          </div>

          <button
            type="button"
            onClick={handleGetCurrentLocation}
            disabled={gettingLocation}
            className="touch flex items-center justify-center gap-2 border-2 border-ink bg-paper px-3 py-1.5 font-mono text-xs font-semibold uppercase tracking-wide text-ink shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] transition hover:bg-accent hover:text-paper active:translate-x-0.5 active:translate-y-0.5"
          >
            <GpsIcon className="h-4 w-4 shrink-0 text-accent group-hover:text-paper" />
            <span>{gettingLocation ? "Buscando GPS..." : "Usar Minha Localização Atual"}</span>
          </button>
        </div>

        {/* Busca por Cidade ou CEP */}
        <div className="flex gap-2">
          <input
            type="text"
            className={`${fieldClass} py-1.5 text-xs`}
            placeholder="Ou busque por CEP ou Cidade (Ex: 13083-852 ou Piracicaba, SP)"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") {
                e.preventDefault();
                handleSearchCityOrZip();
              }
            }}
          />
          <button
            type="button"
            onClick={handleSearchCityOrZip}
            disabled={searchingLocation || !searchQuery.trim()}
            className="shrink-0 border-2 border-ink bg-paper px-3 py-1.5 font-mono text-xs font-semibold uppercase text-ink transition hover:bg-ink hover:text-paper disabled:opacity-50"
          >
            {searchingLocation ? "Buscando..." : "Buscar"}
          </button>
        </div>

        {locationStatus && (
          <p
            className={`font-mono text-xs ${
              locationStatus.type === "error"
                ? "text-risk-critical"
                : locationStatus.type === "success"
                ? "text-risk-low"
                : "text-ink"
            }`}
          >
            {locationStatus.text}
          </p>
        )}

        {/* Mapa Interativo de Seleção */}
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
              placeholder="-22.9099"
              inputMode="decimal"
            />
          </FieldShell>

          <FieldShell label="Longitude" htmlFor="longitude" error={errors.longitude}>
            <input
              id="longitude"
              className={`${fieldClass} tabular text-xs`}
              value={values.longitude}
              onChange={(e) => update("longitude", e.target.value)}
              placeholder="-47.0626"
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

      <button
        type="submit"
        disabled={submitting}
        className="w-full border-2 border-ink bg-accent py-3 font-mono text-sm font-semibold uppercase tracking-wide text-paper transition hover:bg-ink disabled:opacity-50 sm:w-fit sm:px-8"
      >
        {submitting ? "Salvando..." : "Cadastrar Propriedade"}
      </button>
    </form>
  );
}
