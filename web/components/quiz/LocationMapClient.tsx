"use client";

import dynamic from "next/dynamic";

// Leaflet touches `window`, so the map is only ever loaded in the browser.
const LocationMapClient = dynamic(() => import("./LocationMap"), {
  ssr: false,
  loading: () => (
    <div className="flex h-full w-full items-center justify-center bg-paper-deep font-mono text-xs text-ink-soft">
      Carregando mapa...
    </div>
  ),
});

export default LocationMapClient;
