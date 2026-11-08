// Server-only: reads the httpOnly session cookie. Browser code goes through
// the server actions in app/actions/*, never through this module.
import { redirect } from "next/navigation";
import { API_URL } from "./config";
import { getSessionToken } from "./session";
import type {
  Alert,
  AuthUser,
  DeleteResult,
  NewProducer,
  NewProperty,
  NotificationLog,
  Producer,
  Property,
  WeatherReading,
} from "./types";

export type {
  Alert,
  AuthUser,
  CropStage,
  DeleteResult,
  NewProducer,
  NewProperty,
  NotificationLog,
  NotificationStatus,
  Producer,
  Property,
  RiskLevel,
  WeatherReading,
} from "./types";

export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public fields?: Record<string, string>
  ) {
    super(message);
    this.name = "ApiError";
  }
}

export type SessionResponse = { token: string; expires_at: string; user: AuthUser };

type RequestOptions = { method?: string; body?: unknown; auth?: boolean };

async function apiRequest<T>(path: string, opts: RequestOptions = {}): Promise<T> {
  const auth = opts.auth !== false;
  const headers: Record<string, string> = {};
  if (opts.body !== undefined) headers["Content-Type"] = "application/json";
  if (auth) {
    const token = await getSessionToken();
    if (token) headers.Authorization = `Bearer ${token}`;
  }

  const res = await fetch(`${API_URL}${path}`, {
    method: opts.method ?? "GET",
    headers,
    body: opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
    cache: "no-store",
  });

  if (res.status === 401 && auth) redirect("/sair");

  if (!res.ok) {
    let message = `API ${path} failed: ${res.status}`;
    let fields: Record<string, string> | undefined;
    try {
      const payload = (await res.json()) as { error?: string; fields?: Record<string, string> };
      if (payload.error) message = payload.error;
      fields = payload.fields;
    } catch {
      // response body wasn't JSON; keep generic
    }
    throw new ApiError(message, res.status, fields);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

// properties

export function listProperties(): Promise<Property[]> {
  return apiRequest<Property[]>("/properties");
}

export function getProperty(id: string): Promise<Property> {
  return apiRequest<Property>(`/properties/${id}`);
}

export function createProperty(property: NewProperty): Promise<Property> {
  return apiRequest<Property>("/properties", { method: "POST", body: property });
}

export function updateProperty(id: string, property: NewProperty): Promise<Property> {
  return apiRequest<Property>(`/properties/${id}`, { method: "PUT", body: property });
}

export function deleteProperty(id: string): Promise<DeleteResult> {
  return apiRequest<DeleteResult>(`/properties/${id}`, { method: "DELETE" });
}

export function getPropertyAlerts(propertyId: string): Promise<Alert[]> {
  return apiRequest<Alert[]>(`/properties/${propertyId}/alerts`);
}

export function getPropertyForecast(propertyId: string): Promise<WeatherReading[]> {
  return apiRequest<WeatherReading[]>(`/properties/${propertyId}/forecast`);
}

// producers

export function listProducers(): Promise<Producer[]> {
  return apiRequest<Producer[]>("/producers");
}

export function getProducer(id: string): Promise<Producer> {
  return apiRequest<Producer>(`/producers/${id}`);
}

export function createProducer(producer: NewProducer): Promise<Producer> {
  return apiRequest<Producer>("/producers", { method: "POST", body: producer });
}

export function updateProducer(id: string, producer: NewProducer): Promise<Producer> {
  return apiRequest<Producer>(`/producers/${id}`, { method: "PUT", body: producer });
}

export function deleteProducer(id: string): Promise<DeleteResult> {
  return apiRequest<DeleteResult>(`/producers/${id}`, { method: "DELETE" });
}

export function listProducerProperties(producerId: string): Promise<Property[]> {
  return apiRequest<Property[]>(`/producers/${producerId}/properties`);
}

export function listProducerNotifications(producerId: string): Promise<NotificationLog[]> {
  return apiRequest<NotificationLog[]>(`/producers/${producerId}/notifications`);
}

// alerts, scan, notifications

export function listRecentAlerts(): Promise<Alert[]> {
  return apiRequest<Alert[]>("/alerts");
}

export function resolveAlert(alertId: string): Promise<DeleteResult> {
  return apiRequest<DeleteResult>(`/alerts/${alertId}/resolve`, { method: "POST", body: {} });
}

export function triggerScan(): Promise<{ status: string }> {
  return apiRequest<{ status: string }>("/scan/run", { method: "POST", body: {} });
}

export function listRecentNotifications(): Promise<NotificationLog[]> {
  return apiRequest<NotificationLog[]>("/notifications");
}

// auth

export function signup(payload: unknown): Promise<SessionResponse> {
  return apiRequest<SessionResponse>("/auth/signup", { method: "POST", body: payload, auth: false });
}

export function login(email: string, password: string): Promise<SessionResponse> {
  return apiRequest<SessionResponse>("/auth/login", {
    method: "POST",
    body: { email, password },
    auth: false,
  });
}

export function logout(): Promise<void> {
  return apiRequest<void>("/auth/logout", { method: "POST" });
}
