"use client";

import { useEffect } from "react";
import L from "leaflet";
import { MapContainer, Marker, TileLayer, useMap, useMapEvents } from "react-leaflet";
import "leaflet/dist/leaflet.css";

const BRAZIL_CENTER: [number, number] = [-14.235, -51.9253];

const pinIcon = L.divIcon({
  className: "",
  html: '<div style="width:22px;height:22px;border:3px solid #2a2420;background:#8a4a2f;border-radius:50% 50% 50% 0;transform:rotate(-45deg);box-shadow:3px 3px 0 #2a2420"></div>',
  iconSize: [22, 22],
  iconAnchor: [11, 22],
});

function Recenter({ pin }: { pin: [number, number] | null }) {
  const map = useMap();
  const lat = pin?.[0];
  const lon = pin?.[1];
  useEffect(() => {
    if (lat !== undefined && lon !== undefined) map.setView([lat, lon], Math.max(map.getZoom(), 11));
  }, [lat, lon, map]);
  return null;
}

function ClickToPlace({ onChange }: { onChange: (lat: number, lon: number) => void }) {
  useMapEvents({
    click(e) {
      onChange(e.latlng.lat, e.latlng.lng);
    },
  });
  return null;
}

export default function LocationMap({
  latitude,
  longitude,
  onChange,
}: {
  latitude: number | null;
  longitude: number | null;
  onChange: (lat: number, lon: number) => void;
}) {
  const pin: [number, number] | null = latitude !== null && longitude !== null ? [latitude, longitude] : null;

  return (
    <MapContainer center={pin ?? BRAZIL_CENTER} zoom={pin ? 12 : 4} className="h-full w-full">
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />
      <Recenter pin={pin} />
      <ClickToPlace onChange={onChange} />
      {pin && (
        <Marker
          position={pin}
          icon={pinIcon}
          draggable
          eventHandlers={{
            dragend(e) {
              const p = (e.target as L.Marker).getLatLng();
              onChange(p.lat, p.lng);
            },
          }}
        />
      )}
    </MapContainer>
  );
}
