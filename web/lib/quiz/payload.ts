import { cropList, normalizePhone, profileOf } from "./steps";
import type { Concern, CropStage, Profile, QuizAnswers, SoilType } from "./types";

export type SignupPayload = {
  profile: Profile;
  name: string;
  whatsapp_phone: string;
  email: string;
  password: string;
  organization: string | null;
  concerns: Concern[];
  onboarding: Record<string, string>;
  producer?: { city: string; state: string };
  property?: {
    latitude: number;
    longitude: number;
    soil_type: SoilType;
    /** One property is created per crop, sharing the location and soil. */
    crops: { crop: string; crop_stage: CropStage }[];
  };
};

const INCOMPLETE = "incomplete quiz answers";

/** Turns the answers into the POST /auth/signup body; answers of the other profile are ignored. */
export function buildSignupPayload(a: QuizAnswers, password: string): SignupPayload {
  const profile = profileOf(a.work_style);
  const phone = normalizePhone(a.phone);
  if (profile === null || a.work_style === null || phone === null) throw new Error(INCOMPLETE);

  const base = {
    profile,
    name: a.name.trim(),
    whatsapp_phone: phone,
    email: a.email.trim().toLowerCase(),
    password,
    concerns: a.concerns,
  };

  if (profile === "technician") {
    if (a.producers_bucket === null) throw new Error(INCOMPLETE);
    return {
      ...base,
      organization: a.organization.trim(),
      onboarding: { work_style: a.work_style, producers_bucket: a.producers_bucket },
    };
  }

  const crops: { crop: string; crop_stage: CropStage }[] = [];
  for (const { crop, stage } of cropList(a)) {
    if (crop === "" || stage === null) throw new Error(INCOMPLETE);
    crops.push({ crop, crop_stage: stage });
  }
  if (
    a.areas_bucket === null ||
    crops.length === 0 ||
    a.soil_type === null ||
    a.latitude === null ||
    a.longitude === null
  ) {
    throw new Error(INCOMPLETE);
  }

  return {
    ...base,
    organization: null,
    onboarding: { work_style: a.work_style, areas_bucket: a.areas_bucket },
    producer: { city: a.city.trim(), state: a.state },
    property: {
      latitude: a.latitude,
      longitude: a.longitude,
      soil_type: a.soil_type,
      crops,
    },
  };
}
