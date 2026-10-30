"use client";

import { useEffect, useRef, useState } from "react";
import { searchCities, type CityResult } from "@/lib/geocode";
import { bigFieldClass } from "./styles";

export default function CitySearch({
  initialLabel,
  onPick,
}: {
  initialLabel: string;
  onPick: (city: CityResult) => void;
}) {
  const [query, setQuery] = useState(initialLabel);
  const [picked, setPicked] = useState(initialLabel !== "");
  const [results, setResults] = useState<CityResult[]>([]);
  const [status, setStatus] = useState<"idle" | "loading" | "error">("idle");
  const requestId = useRef(0);

  useEffect(() => {
    const q = query.trim();
    if (picked || q.length < 2) {
      requestId.current += 1; // drop any in-flight answer
      // eslint-disable-next-line react-hooks/set-state-in-effect -- clear stale results when the search is no longer active
      setResults([]);
      setStatus("idle");
      return;
    }

    const id = ++requestId.current;
    const timer = setTimeout(async () => {
      setStatus("loading");
      try {
        const found = await searchCities(q);
        if (id !== requestId.current) return;
        setResults(found);
        setStatus("idle");
      } catch {
        if (id !== requestId.current) return;
        setResults([]);
        setStatus("error");
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [query, picked]);

  function choose(city: CityResult) {
    setQuery(city.label);
    setPicked(true);
    setResults([]);
    setStatus("idle");
    onPick(city);
  }

  return (
    <div className="flex flex-col gap-2">
      <label
        htmlFor="city-search"
        className="font-mono text-[11px] font-semibold uppercase tracking-wide text-ink-soft"
      >
        Cidade
      </label>
      <input
        id="city-search"
        className={bigFieldClass}
        value={query}
        onChange={(e) => {
          setQuery(e.target.value);
          setPicked(false);
        }}
        autoComplete="off"
        placeholder="Digite o nome da cidade"
      />

      {status === "loading" && <p className="font-mono text-xs text-ink-soft">Buscando...</p>}
      {status === "error" && (
        <p className="font-mono text-xs text-risk-critical">
          Não deu para buscar agora. Use &ldquo;Digitar à mão&rdquo; abaixo.
        </p>
      )}
      {!picked && status === "idle" && query.trim().length >= 2 && results.length === 0 && (
        <p className="font-mono text-xs text-ink-soft">Nenhuma cidade encontrada.</p>
      )}

      {results.length > 0 && (
        <ul className="flex flex-col border-2 border-ink bg-card">
          {results.map((city) => (
            <li key={`${city.label}-${city.latitude}`}>
              <button
                type="button"
                onClick={() => choose(city)}
                className="min-h-12 w-full border-b border-line px-4 py-2 text-left font-mono text-sm text-ink last:border-b-0 hover:bg-paper-deep"
              >
                {city.label}
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
