import type { CropStage } from "../types";

export type { CropStage };

export type Profile = "producer" | "technician";
export type WorkStyle = "family" | "multi_area" | "technician";
export type AreasBucket = "1" | "2-5" | "6+";
export type ProducersBucket = "ate-10" | "11-40" | "mais-40";
export type Concern = "frost" | "severe_water_stress" | "heat_wave" | "heavy_rain";
export type SoilType = "argiloso" | "arenoso" | "misto" | "nao_informado";

/** Everything the quiz collects. The password is deliberately not part of it, so it can never be persisted. */
export type QuizAnswers = {
  work_style: WorkStyle | null;
  areas_bucket: AreasBucket | null;
  producers_bucket: ProducersBucket | null;
  name: string;
  /** What the person typed (DDD + number); normalized only when building the payload. */
  phone: string;
  organization: string;
  /** One of CROP_OPTIONS or "outra"; the typed name lives in crop_custom. */
  crop_choice: string;
  crop_custom: string;
  crop_stage: CropStage | null;
  city: string;
  state: string;
  latitude: number | null;
  longitude: number | null;
  soil_type: SoilType | null;
  concerns: Concern[];
  email: string;
};
