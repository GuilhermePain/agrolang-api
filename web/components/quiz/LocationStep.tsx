"use client";

import { useState } from "react";
import type { CityResult } from "@/lib/geocode";
import type { QuizAnswers } from "@/lib/quiz/types";
import { FieldShell } from "../FieldShell";
import CitySearch from "./CitySearch";
import LocationMapClient from "./LocationMapClient";
import { bigFieldClass, optionClass } from "./styles";

const round5 = (n: number) => Math.round(n * 1e5) / 1e5;

export default function LocationStep({
  answers,
  onChange,
}: {
  answers: QuizAnswers;
  onChange: (patch: Partial<QuizAnswers>) => void;
}) {
  const [manual, setManual] = useState(false);
  const [gpsMessage, setGpsMessage] = useState<string | null>(null);

  function pickCity(city: CityResult) {
    onChange({
      city: city.name,
      state: city.state,
      latitude: round5(city.latitude),
      longitude: round5(city.longitude),
    });
  }

  function setPin(lat: number, lon: number) {
    setGpsMessage(null);
    onChange({ latitude: round5(lat), longitude: round5(lon) });
  }

  function useMyLocation() {
    if (!("geolocation" in navigator)) {
      setGpsMessage("Seu aparelho não deixa usar a localização. Arraste o pino no mapa.");
      return;
    }
    navigator.geolocation.getCurrentPosition(
      (pos) => setPin(pos.coords.latitude, pos.coords.longitude),
      () => setGpsMessage("Não deu para pegar a localização. Arraste o pino no mapa."),
      { enableHighAccuracy: true, timeout: 10000 }
    );
  }

  return (
    <>
      <CitySearch
        initialLabel={answers.city && answers.state ? `${answers.city}, ${answers.state}` : ""}
        onPick={pickCity}
      />

      <button
        type="button"
        onClick={() => setManual((m) => !m)}
        className="w-fit font-mono text-xs text-accent underline"
      >
        {manual ? "Voltar para a busca" : "Não achei, digitar à mão"}
      </button>

      {manual && (
        <div className="grid grid-cols-[1fr_5rem] gap-3">
          <FieldShell label="Cidade" htmlFor="manual-city">
            <input
              id="manual-city"
              className={bigFieldClass}
              value={answers.city}
              onChange={(e) => onChange({ city: e.target.value })}
            />
          </FieldShell>
          <FieldShell label="UF" htmlFor="manual-state">
            <input
              id="manual-state"
              className={bigFieldClass}
              value={answers.state}
              maxLength={2}
              onChange={(e) => onChange({ state: e.target.value.toUpperCase() })}
            />
          </FieldShell>
        </div>
      )}

      <div className="h-72 border-2 border-ink shadow-[var(--shadow-hard)]">
        <LocationMapClient latitude={answers.latitude} longitude={answers.longitude} onChange={setPin} />
      </div>
      <p className="font-mono text-xs text-ink-soft">
        Toque no mapa ou arraste o pino até a sua área.
      </p>

      <button type="button" onClick={useMyLocation} className={optionClass(false)}>
        <span className="font-semibold">Usar minha localização</span>
        <span className="mt-0.5 text-xs text-ink-soft">Funciona se você estiver na sua terra agora</span>
      </button>
      {gpsMessage && <p className="font-mono text-xs text-risk-critical">{gpsMessage}</p>}
    </>
  );
}
