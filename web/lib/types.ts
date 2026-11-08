export type RiskLevel = "low" | "medium" | "critical";

export type CropStage = "germination" | "flowering" | "harvest";

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

export type AuthUser = {
  id: string;
  email: string;
  profile: "producer" | "technician";
  name: string;
  whatsapp_phone: string;
  organization: string | null;
  concerns: string[];
  onboarding: Record<string, string>;
  created_at: string;
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

export type DeleteResult = { status: string; id: string };
