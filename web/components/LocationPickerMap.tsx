"use client";

import { useEffect } from "react";
import { MapContainer, TileLayer, CircleMarker, useMapEvents, useMap, LayersControl } from "react-leaflet";
import "leaflet/dist/leaflet.css";
import { MountainIcon } from "./Icons";

interface LocationPickerMapProps {
  latitude?: number | null;
  longitude?: number | null;
  onChange: (lat: number, lon: number) => void;
}

const DEFAULT_CENTER: [number, number] = [-5.3, -39.3]; // Centralizado no Ceará / Nordeste por padrão ou Brasil

function MapController({ center, zoom }: { center: [number, number]; zoom?: number }) {
  const map = useMap();
  useEffect(() => {
    map.setView(center, zoom ?? map.getZoom());
  }, [center, zoom, map]);
  return null;
}

function ClickHandler({ onPick }: { onPick: (lat: number, lon: number) => void }) {
  useMapEvents({
    click(e) {
      onPick(Number(e.latlng.lat.toFixed(6)), Number(e.latlng.lng.toFixed(6)));
    },
  });
  return null;
}

export default function LocationPickerMap({
  latitude,
  longitude,
  onChange,
}: LocationPickerMapProps) {
  const hasCoords =
    typeof latitude === "number" && !isNaN(latitude) && typeof longitude === "number" && !isNaN(longitude);
  const center: [number, number] = hasCoords ? [latitude, longitude] : DEFAULT_CENTER;
  const zoom = hasCoords ? 14 : 6;

  return (
    <div className="relative h-72 w-full overflow-hidden border-2 border-ink bg-card">
      <MapContainer
        center={center}
        zoom={zoom}
        className="h-full w-full cursor-crosshair"
      >
        <LayersControl position="topright">
          <LayersControl.BaseLayer checked name="Topográfico / Curvas de Nível (OpenTopoMap)">
            <TileLayer
              attribution='Mapa: &copy; <a href="https://opentopomap.org">OpenTopoMap</a> (<a href="https://creativecommons.org/licenses/by-sa/3.0/">CC-BY-SA</a>), SRTM'
              url="https://{s}.tile.opentopomap.org/{z}/{x}/{y}.png"
              maxZoom={17}
            />
          </LayersControl.BaseLayer>

          <LayersControl.BaseLayer name="Satélite Orbital & Superfície (Esri)">
            <TileLayer
              attribution="Tiles &copy; Esri, Maxar"
              url="https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}"
              maxZoom={18}
            />
          </LayersControl.BaseLayer>

          <LayersControl.BaseLayer name="Relevo Sombreado Esri">
            <TileLayer
              attribution="Tiles &copy; Esri &mdash; USGS, Esri"
              url="https://server.arcgisonline.com/ArcGIS/rest/services/World_Topo_Map/MapServer/tile/{z}/{y}/{x}"
              maxZoom={18}
            />
          </LayersControl.BaseLayer>

          <LayersControl.BaseLayer name="Padrão (OpenStreetMap)">
            <TileLayer
              attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
              url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
            />
          </LayersControl.BaseLayer>
        </LayersControl>

        <ClickHandler onPick={onChange} />
        {hasCoords && (
          <>
            <MapController center={center} zoom={14} />
            <CircleMarker
              center={center}
              radius={10}
              pathOptions={{
                color: "#1a1614",
                weight: 3,
                fillColor: "#059669",
                fillOpacity: 0.9,
              }}
            />
          </>
        )}
      </MapContainer>
      <div className="pointer-events-none absolute bottom-2 left-2 z-[400] flex items-center gap-2 rounded bg-paper/95 px-2.5 py-1 font-mono text-[11px] font-medium text-ink shadow border border-ink">
        <MountainIcon className="h-3.5 w-3.5 text-accent" />
        <span>
          {hasCoords
            ? `Lat: ${latitude.toFixed(4)} | Lon: ${longitude.toFixed(4)} (clique para reposicionar)`
            : "Clique no mapa para marcar a coordenada exata"}
        </span>
      </div>
    </div>
  );
}
