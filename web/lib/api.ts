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
  cloud_cover_pct?: number;
  wind_speed_kmh?: number;
  weather_code?: number;
  apparent_temp_c?: number;
  rain_prob_pct?: number;
};

export type CropStage = "germination" | "flowering" | "harvest";

export type NewProducer = {
  name: string;
  whatsapp_phone: string;
  city: string;
  state: string;
};

export type NewProperty = {
  producer_id: string;
  latitude: number;
  longitude: number;
  crop: string;
  soil_type: string;
  crop_stage: CropStage;
};

export type NotificationStatus = "sent" | "delivered" | "read" | "failed";

export type NotificationLog = {
  id: string;
  alert_id?: string;
  producer_id: string;
  property_id?: string;
  phone: string;
  message: string;
  status: NotificationStatus;
  sent_at: string;
  delivered_at?: string | null;
  read_at?: string | null;
  error_message?: string | null;
};

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number
  ) {
    super(message);
    this.name = "ApiError";
  }
}

async function apiFetch<T>(path: string): Promise<T> {
  const res = await fetch(`${API_URL}${path}`, { cache: "no-store" });
  if (!res.ok) {
    throw new ApiError(`API ${path} failed: ${res.status}`, res.status);
  }
  return res.json() as Promise<T>;
}

async function apiMutate<TBody, TResult>(
  path: string,
  method: "POST" | "PUT" | "DELETE",
  body?: TBody
): Promise<TResult> {
  const options: RequestInit = {
    method,
    headers: { "Content-Type": "application/json" },
    cache: "no-store",
  };
  if (body !== undefined) {
    options.body = JSON.stringify(body);
  }

  const res = await fetch(`${API_URL}${path}`, options);
  if (!res.ok) {
    let message = `API ${path} failed: ${res.status}`;
    try {
      const payload = (await res.json()) as { error?: string };
      if (payload.error) message = payload.error;
    } catch {
      // response body wasn't JSON; keep generic
    }
    throw new ApiError(message, res.status);
  }
  return res.json() as Promise<TResult>;
}

export function listProperties(): Promise<Property[]> {
  return apiFetch<Property[]>("/properties");
}

export function getProperty(id: string): Promise<Property> {
  return apiFetch<Property>(`/properties/${id}`);
}

export function createProperty(property: NewProperty): Promise<Property> {
  return apiMutate<NewProperty, Property>("/properties", "POST", property);
}

export function updateProperty(id: string, property: NewProperty): Promise<Property> {
  return apiMutate<NewProperty, Property>(`/properties/${id}`, "PUT", property);
}

export function deleteProperty(id: string): Promise<{ status: string; id: string }> {
  return apiMutate<undefined, { status: string; id: string }>(`/properties/${id}`, "DELETE");
}

export function listProducers(): Promise<Producer[]> {
  return apiFetch<Producer[]>("/producers");
}

export function getProducer(id: string): Promise<Producer> {
  return apiFetch<Producer>(`/producers/${id}`);
}

export function createProducer(producer: NewProducer): Promise<Producer> {
  return apiMutate<NewProducer, Producer>("/producers", "POST", producer);
}

export function updateProducer(id: string, producer: NewProducer): Promise<Producer> {
  return apiMutate<NewProducer, Producer>(`/producers/${id}`, "PUT", producer);
}

export function deleteProducer(id: string): Promise<{ status: string; id: string }> {
  return apiMutate<undefined, { status: string; id: string }>(`/producers/${id}`, "DELETE");
}

export function listProducerProperties(producerId: string): Promise<Property[]> {
  return apiFetch<Property[]>(`/producers/${producerId}/properties`);
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

export function resolveAlert(alertId: string): Promise<{ status: string; id: string }> {
  return apiMutate<Record<string, never>, { status: string; id: string }>(
    `/alerts/${alertId}/resolve`,
    "POST",
    {}
  );
}

export function triggerScan(): Promise<{ status: string }> {
  return apiMutate<Record<string, never>, { status: string }>("/scan/run", "POST", {});
}

export function listRecentNotifications(): Promise<NotificationLog[]> {
  return apiFetch<NotificationLog[]>("/notifications");
}

export function listProducerNotifications(producerId: string): Promise<NotificationLog[]> {
  return apiFetch<NotificationLog[]>(`/producers/${producerId}/notifications`);
}
