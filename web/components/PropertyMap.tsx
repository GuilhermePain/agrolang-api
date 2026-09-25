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
            radius={10}
            pathOptions={{ color: riskColors[level], fillColor: riskColors[level], fillOpacity: 0.8 }}
          >
            <Popup>
              <div className="flex flex-col gap-1">
                <span className="font-semibold">{property.crop}</span>
                <span>Risco atual: {riskLabels[level]}</span>
                <Link href={`/properties/${property.id}`} className="text-blue-600 underline">
                  Ver detalhes
                </Link>
              </div>
            </Popup>
          </CircleMarker>
        );
      })}
    </MapContainer>
  );
}
