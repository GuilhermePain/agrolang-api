"use client";

import { useState, useEffect } from "react";
import { MapContainer, TileLayer, CircleMarker, Popup, LayersControl, useMap } from "react-leaflet";
import Link from "next/link";
import "leaflet/dist/leaflet.css";
import type { Property, Alert } from "@/lib/types";
import { currentRiskLevel, riskColors, riskLabels, cropLabels, soilLabels } from "@/lib/risk";
import { MountainIcon, LayersIcon, WindIcon } from "./Icons";

type PropertyWithAlerts = {
  property: Property;
  alerts: Alert[];
};

const DEFAULT_CENTER: [number, number] = [-14.235, -51.9253];

function MapViewController({
  targetCenter,
  targetZoom,
}: {
  targetCenter: [number, number] | null;
  targetZoom: number | null;
}) {
  const map = useMap();
  useEffect(() => {
    if (targetCenter && targetZoom) {
      map.setView(targetCenter, targetZoom, { animate: true });
    }
  }, [targetCenter, targetZoom, map]);
  return null;
}

export default function PropertyMap({ items }: { items: PropertyWithAlerts[] }) {
  const [activeTab, setActiveTab] = useState<"map" | "winds">("map");
  const [viewTarget, setViewTarget] = useState<{
    center: [number, number];
    zoom: number;
  } | null>(null);

  // Se houver propriedades no Ceará, iniciamos com visão geral do Ceará ou Brasil
  const hasCeara = items.some(
    (i) => i.property.latitude > -8.0 && i.property.latitude < -2.5 && i.property.longitude > -42.0 && i.property.longitude < -37.0
  );

  const initialCenter: [number, number] = hasCeara
    ? [-5.3, -39.3]
    : items.length > 0
    ? [items[0].property.latitude, items[0].property.longitude]
    : DEFAULT_CENTER;

  const initialZoom = hasCeara ? 7 : items.length > 0 ? 6 : 4;

  return (
    <div className="relative h-full w-full flex flex-col">
      {/* BARRA SUPERIOR: ALTERNÂNCIA MAPA TOPOGRÁFICO VS RADAR DE VENTOS */}
      <div className="z-[1000] flex flex-wrap items-center justify-between gap-2 border-b-2 border-ink bg-paper px-3 py-2">
        {/* ABAS: MAPA VS VENTOS AO VIVO */}
        <div className="flex rounded border border-ink p-0.5 bg-card">
          <button
            type="button"
            onClick={() => setActiveTab("map")}
            className={`flex items-center gap-1.5 px-2.5 py-1 font-mono text-[11px] font-bold transition ${
              activeTab === "map"
                ? "bg-ink text-paper"
                : "text-ink hover:bg-line/40"
            }`}
          >
            <MountainIcon className="h-3.5 w-3.5" />
            <span>Mapa Topográfico & Talhões</span>
          </button>
          <button
            type="button"
            onClick={() => setActiveTab("winds")}
            className={`flex items-center gap-1.5 px-2.5 py-1 font-mono text-[11px] font-bold transition ${
              activeTab === "winds"
                ? "bg-accent text-paper"
                : "text-ink hover:bg-line/40"
            }`}
          >
            <WindIcon className="h-3.5 w-3.5" />
            <span>Visualizador de Ventos ao Vivo</span>
          </button>
        </div>

        {/* ATALHOS DE REGIÃO (visíveis na aba mapa) */}
        {activeTab === "map" && (
          <div className="flex items-center gap-1">
            <span className="font-mono text-[10px] uppercase tracking-wider text-ink-soft hidden sm:inline">
              Foco:
            </span>
            <button
              type="button"
              onClick={() => setViewTarget({ center: [-5.3, -39.3], zoom: 7 })}
              className="rounded border border-ink bg-card px-2 py-0.5 font-mono text-[10px] font-semibold text-ink transition hover:bg-ink hover:text-paper"
            >
              Ceará
            </button>
            <button
              type="button"
              onClick={() => setViewTarget({ center: [-22.5, -47.1], zoom: 7 })}
              className="rounded border border-ink bg-card px-2 py-0.5 font-mono text-[10px] font-semibold text-ink transition hover:bg-ink hover:text-paper"
            >
              Sudeste
            </button>
            <button
              type="button"
              onClick={() => setViewTarget({ center: DEFAULT_CENTER, zoom: 4 })}
              className="rounded border border-ink bg-card px-2 py-0.5 font-mono text-[10px] font-semibold text-ink transition hover:bg-ink hover:text-paper"
            >
              Brasil
            </button>
          </div>
        )}
      </div>

      {/* CONTEÚDO: MAPA LEAFLET OU VISUALIZADOR DE VENTOS WINDY */}
      <div className="relative flex-1 w-full overflow-hidden">
        {activeTab === "winds" ? (
          /* ================= VISUALIZADOR METEOROLÓGICO DE VENTOS AO VIVO ================= */
          <div className="relative h-full w-full bg-slate-900">
            <iframe
              width="100%"
              height="100%"
              src="https://embed.windy.com/embed2.html?lat=-5.3&lon=-39.3&detailLat=-5.3&detailLon=-39.3&width=650&height=450&zoom=6&level=surface&overlay=wind&product=ecmwf&menu=&message=&marker=&calendar=now&pressure=&type=map&location=coordinates&detail=&metricWind=km%2Fh&metricTemp=%C2%B0C&radarRange=-1"
              frameBorder="0"
              className="h-full w-full border-none"
              title="Visualizador Meteorológico de Ventos e Correntes de Ar"
            />
            <div className="pointer-events-none absolute bottom-2 left-2 z-10 rounded border border-ink bg-paper/90 px-2.5 py-1 font-mono text-[11px] text-ink shadow">
              <span className="font-bold">Correntes de Vento ECMWF (Tempo Real):</span> Animação dinâmica das rajadas e frentes de vento sobre o Brasil e Ceará.
            </div>
          </div>
        ) : (
          /* ================= MAPA LEAFLET TOPOGRÁFICO & TALHÕES ================= */
          <MapContainer center={initialCenter} zoom={initialZoom} className="h-full w-full">
            {viewTarget && (
              <MapViewController targetCenter={viewTarget.center} targetZoom={viewTarget.zoom} />
            )}

            <LayersControl position="topright">
              {/* CAMADA BASE 1: TOPOGRÁFICA COM CURVAS DE NÍVEL E ALTIMETRIA */}
              <LayersControl.BaseLayer checked name="Topográfico / Curvas de Nível (OpenTopoMap)">
                <TileLayer
                  attribution='Mapa: &copy; <a href="https://opentopomap.org">OpenTopoMap</a> (<a href="https://creativecommons.org/licenses/by-sa/3.0/">CC-BY-SA</a>), SRTM'
                  url="https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png"
                  maxZoom={17}
                />
              </LayersControl.BaseLayer>

              {/* CAMADA BASE 2: SATÉLITE ORBITAL DE ALTA RESOLUÇÃO */}
              <LayersControl.BaseLayer name="Satélite Orbital & Superfície (Esri Imagery)">
                <TileLayer
                  attribution="Tiles &copy; Esri, Maxar, Earthstar Geographics"
                  url="https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}"
                  maxZoom={18}
                />
              </LayersControl.BaseLayer>

              {/* CAMADA BASE 3: RELEVO SOMBREADO & FÍSICO */}
              <LayersControl.BaseLayer name="Relevo Sombreado & Físico (Esri Topo)">
                <TileLayer
                  attribution="Tiles &copy; Esri &mdash; DeLorme, NAVTEQ, TomTom, Intermap, iPC, USGS, FAO, NPS"
                  url="https://server.arcgisonline.com/ArcGIS/rest/services/World_Topo_Map/MapServer/tile/{z}/{y}/{x}"
                  maxZoom={18}
                />
              </LayersControl.BaseLayer>

              {/* CAMADA BASE 4: PADRÃO RODOVIÁRIO */}
              <LayersControl.BaseLayer name="Padrão Rodoviário (OpenStreetMap)">
                <TileLayer
                  attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
                  url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
                />
              </LayersControl.BaseLayer>
            </LayersControl>

            {/* MARCADORES DAS PROPRIEDADES / TALHÕES */}
            {items.map(({ property, alerts }) => {
              const level = currentRiskLevel(alerts);
              const cropName = cropLabels[property.crop.toLowerCase()] || property.crop;
              const soilName = soilLabels[property.soil_type.toLowerCase()] || property.soil_type;

              return (
                <CircleMarker
                  key={property.id}
                  center={[property.latitude, property.longitude]}
                  radius={11}
                  pathOptions={{
                    color: "#2a2420",
                    weight: 2,
                    fillColor: riskColors[level],
                    fillOpacity: 0.9,
                  }}
                >
                  <Popup>
                    <div className="flex flex-col gap-1 font-mono">
                      <span className="font-display text-base italic text-ink">
                        {cropName}
                      </span>
                      <span className="text-xs text-ink-soft">
                        Solo: <span className="font-semibold text-ink">{soilName}</span>
                      </span>
                      <span className="text-xs text-ink-soft">
                        Risco Atual: <span className="font-semibold text-ink">{riskLabels[level]}</span>
                      </span>
                      <span className="text-[10px] text-ink-soft">
                        {property.latitude.toFixed(4)}, {property.longitude.toFixed(4)}
                      </span>
                      <Link
                        href={`/properties/${property.id}`}
                        className="mt-1.5 text-xs font-semibold uppercase tracking-wide text-accent underline"
                      >
                        Ver detalhes e clima &rarr;
                      </Link>
                    </div>
                  </Popup>
                </CircleMarker>
              );
            })}
          </MapContainer>
        )}
      </div>

      {/* LEGENDA INFORMATIVA: RELEVO */}
      {activeTab === "map" && (
        <div className="pointer-events-none absolute bottom-3 left-3 z-[1000] hidden rounded border-2 border-ink bg-paper/95 px-3 py-1.5 font-mono text-[11px] shadow-[var(--shadow-hard)] backdrop-blur-sm sm:block">
          <div className="flex items-center gap-1.5 font-bold text-ink">
            <LayersIcon className="h-3.5 w-3.5 text-accent" />
            <span>Relevo e Curvas de Nível Ativas</span>
          </div>
          <p className="text-[10px] text-ink-soft">
            Linhas de contorno altimétrico e bacias hidrográficas.
          </p>
        </div>
      )}
    </div>
  );
}
