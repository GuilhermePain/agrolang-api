"use client";

import dynamic from "next/dynamic";

const LocationPickerMap = dynamic(() => import("./LocationPickerMap"), {
  ssr: false,
  loading: () => (
    <div className="flex h-64 w-full items-center justify-center border-2 border-line bg-paper-deep font-mono text-xs text-ink-soft">
      Carregando mapa interativo...
    </div>
  ),
});

interface LocationPickerClientProps {
  latitude?: number | null;
  longitude?: number | null;
  onChange: (lat: number, lon: number) => void;
}

export default function LocationPickerClient(props: LocationPickerClientProps) {
  return <LocationPickerMap {...props} />;
}

