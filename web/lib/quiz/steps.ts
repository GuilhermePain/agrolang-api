import { CROP_OTHER } from "./options";
import type { Profile, QuizAnswers, WorkStyle } from "./types";

export type StepId =
  | "style"
  | "areas"
  | "identity"
  | "crop"
  | "location"
  | "soil"
  | "concerns"
  | "producers"
  | "account";

const PRODUCER_STEPS: StepId[] = ["style", "areas", "identity", "crop", "location", "soil", "concerns", "account"];
const TECHNICIAN_STEPS: StepId[] = ["style", "identity", "producers", "concerns", "account"];

export function profileOf(workStyle: WorkStyle | null): Profile | null {
  if (workStyle === null) return null;
  return workStyle === "technician" ? "technician" : "producer";
}

export function stepsFor(workStyle: WorkStyle | null): StepId[] {
  const profile = profileOf(workStyle);
  if (profile === null) return ["style"];
  return profile === "technician" ? TECHNICIAN_STEPS : PRODUCER_STEPS;
}

/** DDD + number (10 or 11 digits), with or without the 55 country code. Returns digits starting with 55, or null. */
export function normalizePhone(raw: string): string | null {
  const digits = raw.replace(/\D/g, "");
  if (/^55\d{10,11}$/.test(digits)) return digits;
  if (/^\d{10,11}$/.test(digits)) return `55${digits}`;
  return null;
}

export function cropValue(a: QuizAnswers): string {
  return a.crop_choice === CROP_OTHER ? a.crop_custom.trim() : a.crop_choice.trim();
}

/** Returns the message to show when the step can't advance yet, or null when it's complete. */
export function validateStep(id: StepId, a: QuizAnswers): string | null {
  switch (id) {
    case "style":
      return a.work_style === null ? "Escolha uma das opções." : null;
    case "areas":
      return a.areas_bucket === null ? "Escolha uma das opções." : null;
    case "producers":
      return a.producers_bucket === null ? "Escolha uma das opções." : null;
    case "identity":
      if (a.name.trim().length < 2) return "Digite seu nome.";
      if (normalizePhone(a.phone) === null) return "Digite o DDD e o número, só números. Ex.: 19 99999-9999.";
      if (a.work_style === "technician" && a.organization.trim() === "") {
        return "Digite o nome da sua organização.";
      }
      return null;
    case "crop":
      if (cropValue(a) === "") return "Escolha ou digite a cultura.";
      if (a.crop_stage === null) return "Escolha em que fase está.";
      return null;
    case "location":
      if (a.latitude === null || a.longitude === null) return "Marque sua terra no mapa.";
      if (a.city.trim() === "" || !/^[A-Z]{2}$/.test(a.state)) return "Escolha a cidade na busca.";
      return null;
    case "soil":
      return a.soil_type === null ? "Escolha o tipo de solo." : null;
    case "concerns":
    case "account":
      return null;
  }
}

/** Index of the first step that still has an error; the last step when all are complete. */
export function firstIncompleteIndex(steps: StepId[], a: QuizAnswers): number {
  const index = steps.findIndex((id) => validateStep(id, a) !== null);
  return index === -1 ? steps.length - 1 : index;
}

const FIELD_TO_STEP: Record<string, StepId> = {
  profile: "style",
  onboarding: "style",
  name: "identity",
  whatsapp_phone: "identity",
  organization: "identity",
  email: "account",
  password: "account",
  concerns: "concerns",
  producer: "location",
  "producer.city": "location",
  "producer.state": "location",
  property: "location",
  "property.latitude": "location",
  "property.longitude": "location",
  "property.crop": "crop",
  "property.crop_stage": "crop",
  "property.soil_type": "soil",
};

export function stepForField(field: string): StepId | null {
  return FIELD_TO_STEP[field] ?? null;
}
