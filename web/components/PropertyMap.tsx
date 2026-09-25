"use client";

import { MapContainer, TileLayer, CircleMarker, Popup } from "react-leaflet";
import Link from "next/link";
import "leaflet/dist/leaflet.css";
import type { Property, Alert } from "@/lib/api";
import { currentRiskLevel, riskColors, riskLabels } from "@/lib/risk";

type PropertyWithAlerts = {
  property: Property;
  alerts: Alert[];
};

const DEFAULT_CENTER: [number, number] = [-14.235, -51.9253];

export default function PropertyMap({ items }: { items: PropertyWithAlerts[] }) {
  const center: [number, number] =
    items.length > 0 ? [items[0].property.latitude, items[0].property.longitude] : DEFAULT_CENTER;

  return (
    <MapContainer center={center} zoom={items.length > 0 ? 8 : 4} className="h-full w-full">
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      {items.map(({ property, alerts }) => {
        const level = currentRiskLevel(alerts);
        return (
          <CircleMarker
            key={property.id}
            center={[property.latitude, property.longitude]}
            radius={11}
            pathOptions={{
              color: "#2a2420",
              weight: 2,
              fillColor: riskColors[level],
              fillOpacity: 0.85,
            }}
          >
            <Popup>
              <div className="flex flex-col gap-1 font-mono">
                <span className="font-display text-base italic text-ink">{property.crop}</span>
                <span className="text-xs text-ink-soft">Risco atual: {riskLabels[level]}</span>
                <Link
                  href={`/properties/${property.id}`}
                  className="mt-1 text-xs font-semibold uppercase tracking-wide text-accent underline"
                >
                  Ver detalhes &rarr;
                </Link>
              </div>
            </Popup>
          </CircleMarker>
        );
      })}
    </MapContainer>
  );
}
