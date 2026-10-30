import { redirect } from "next/navigation";
import { API_URL } from "./config";
import { getSessionToken } from "./session";
import type {
  Alert,
  AuthUser,
  NewProducer,
  NewProperty,
  Producer,
  Property,
  WeatherReading,
} from "./types";

export type {
  Alert,
  AuthUser,
  CropStage,
  NewProducer,
  NewProperty,
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
      // response body wasn't JSON; keep the generic message
    }
    throw new ApiError(message, res.status, fields);
  }

  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export function listProperties(): Promise<Property[]> {
  return apiRequest<Property[]>("/properties");
}

export function listProducers(): Promise<Producer[]> {
  return apiRequest<Producer[]>("/producers");
}

export function createProducer(producer: NewProducer): Promise<Producer> {
  return apiRequest<Producer>("/producers", { method: "POST", body: producer });
}

export function createProperty(property: NewProperty): Promise<Property> {
  return apiRequest<Property>("/properties", { method: "POST", body: property });
}

export function getPropertyAlerts(propertyId: string): Promise<Alert[]> {
  return apiRequest<Alert[]>(`/properties/${propertyId}/alerts`);
}

export function getPropertyForecast(propertyId: string): Promise<WeatherReading[]> {
  return apiRequest<WeatherReading[]>(`/properties/${propertyId}/forecast`);
}

export function listRecentAlerts(): Promise<Alert[]> {
  return apiRequest<Alert[]>("/alerts");
}

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
