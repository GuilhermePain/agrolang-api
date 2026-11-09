import { describe, expect, it } from "vitest";
import { buildSignupPayload } from "./payload";
import { initialAnswers } from "./state";
import type { QuizAnswers } from "./types";

const producer: QuizAnswers = {
  ...initialAnswers,
  work_style: "family",
  areas_bucket: "2-5",
  name: " José da Silva ",
  phone: "19 99999-9999",
  crops: [
    { choice: "Milho", stage: "flowering" },
    { choice: "outra", stage: "harvest" },
  ],
  crop_custom: "Mandioca",
  city: "Campinas",
  state: "SP",
  latitude: -22.9099,
  longitude: -47.0626,
  soil_type: "argiloso",
  concerns: ["frost", "severe_water_stress"],
  email: " Jose@Exemplo.com ",
  organization: "ignorado",
  producers_bucket: "11-40",
};

const technician: QuizAnswers = {
  ...initialAnswers,
  work_style: "technician",
  name: "Ana",
  phone: "(19) 98888-7777",
  organization: " Coopera Campo ",
  producers_bucket: "11-40",
  concerns: ["heavy_rain"],
  email: "ana@coop.com",
  city: "ignorada",
  latitude: 1,
  longitude: 2,
};

describe("buildSignupPayload", () => {
  it("builds the producer payload with producer and property", () => {
    expect(buildSignupPayload(producer, "segredo123")).toEqual({
      profile: "producer",
      name: "José da Silva",
      whatsapp_phone: "5519999999999",
      email: "jose@exemplo.com",
      password: "segredo123",
      organization: null,
      concerns: ["frost", "severe_water_stress"],
      onboarding: { work_style: "family", areas_bucket: "2-5" },
      producer: { city: "Campinas", state: "SP" },
      property: {
        latitude: -22.9099,
        longitude: -47.0626,
        soil_type: "argiloso",
        crops: [
          { crop: "Milho", crop_stage: "flowering" },
          { crop: "Mandioca", crop_stage: "harvest" },
        ],
      },
    });
  });

  it("builds the technician payload without producer or property", () => {
    const payload = buildSignupPayload(technician, "segredo123");
    expect(payload).toEqual({
      profile: "technician",
      name: "Ana",
      whatsapp_phone: "5519988887777",
      email: "ana@coop.com",
      password: "segredo123",
      organization: "Coopera Campo",
      concerns: ["heavy_rain"],
      onboarding: { work_style: "technician", producers_bucket: "11-40" },
    });
    expect("producer" in payload).toBe(false);
    expect("property" in payload).toBe(false);
  });

  it("throws when the answers are incomplete", () => {
    expect(() => buildSignupPayload(initialAnswers, "segredo123")).toThrow("incomplete quiz answers");
    expect(() => buildSignupPayload({ ...producer, latitude: null }, "segredo123")).toThrow(
      "incomplete quiz answers"
    );
    expect(() => buildSignupPayload({ ...producer, crops: [] }, "segredo123")).toThrow("incomplete quiz answers");
    expect(() =>
      buildSignupPayload({ ...producer, crops: [{ choice: "Milho", stage: null }] }, "segredo123")
    ).toThrow("incomplete quiz answers");
    expect(() => buildSignupPayload({ ...technician, producers_bucket: null }, "segredo123")).toThrow(
      "incomplete quiz answers"
    );
  });
});
