const API_URL = process.env.NEXT_PUBLIC_API_URL ?? "http://localhost:8080";

export type RiskLevel = "low" | "medium" | "critical";

export type Producer = {
  id: string;
  name: string;
  whatsapp_phone: string;
  city: string;
  state: string;
  created_at: string;
  updated_at: string;
};

export type Property = {
  id: string;
  producer_id: string;
  latitude: number;
  longitude: number;
  crop: string;
  soil_type: string;
  crop_stage: string;
  created_at: string;
  updated_at: string;
};

export type Alert = {
  id: string;
  property_id: string;
  level: RiskLevel;
  alert_type: string;
  period_start: string;
  period_end: string;
  triggered_at: string;
  resolved_at: string | null;
};

export type WeatherReading = {
  time: string;
  temp_avg_c: number;
  humidity_pct: number;
  precipitation_mm: number;
};

async function apiFetch<T>(path: string): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, { cache: "no-store" });
  if (!res.ok) {
    throw new Error(`API ${path} failed: ${res.status}`);
  }
  return res.json() as Promise<T>;
}

export function listProperties(): Promise<Property[]> {
  return apiFetch<Property[]>("/properties");
}

export function getPropertyAlerts(propertyId: string): Promise<Alert[]> {
  return apiFetch<Alert[]>(`/properties/${propertyId}/alerts`);
}

export function getPropertyForecast(propertyId: string): Promise<WeatherReading[]> {
  return apiFetch<WeatherReading[]>(`/properties/${propertyId}/forecast`);
}

export function listRecentAlerts(): Promise<Alert[]> {
  return apiFetch<Alert[]>("/alerts");
}
