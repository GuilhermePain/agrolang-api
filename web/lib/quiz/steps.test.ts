import { describe, expect, it } from "vitest";
import {
  cropList,
  firstIncompleteIndex,
  normalizePhone,
  profileOf,
  stepForField,
  stepsFor,
  validateStep,
} from "./steps";
import { initialAnswers } from "./state";
import type { QuizAnswers } from "./types";

const answers = (patch: Partial<QuizAnswers> = {}): QuizAnswers => ({ ...initialAnswers, ...patch });

describe("stepsFor / profileOf", () => {
  it("only offers the first step until a work style is chosen", () => {
    expect(stepsFor(null)).toEqual(["style"]);
    expect(profileOf(null)).toBeNull();
  });

  it("gives producers 8 steps ending in the account", () => {
    expect(stepsFor("family")).toEqual([
      "style", "areas", "identity", "crop", "location", "soil", "concerns", "account",
    ]);
    expect(stepsFor("multi_area")).toHaveLength(8);
    expect(profileOf("family")).toBe("producer");
    expect(profileOf("multi_area")).toBe("producer");
  });

  it("gives technicians 5 steps", () => {
    expect(stepsFor("technician")).toEqual(["style", "identity", "producers", "concerns", "account"]);
    expect(profileOf("technician")).toBe("technician");
  });
});

describe("normalizePhone", () => {
  it("adds the Brazil country code to DDD + number", () => {
    expect(normalizePhone("19 99999-9999")).toBe("5519999999999");
    expect(normalizePhone("(19) 3333-4444")).toBe("551933334444");
  });

  it("keeps a number that already has the country code", () => {
    expect(normalizePhone("+55 19 99999-9999")).toBe("5519999999999");
  });

  it("rejects numbers of the wrong length", () => {
    expect(normalizePhone("123")).toBeNull();
    expect(normalizePhone("")).toBeNull();
  });
});

describe("cropList", () => {
  it("returns every chosen crop with its own stage, in order", () => {
    const list = cropList(
      answers({
        crops: [
          { choice: "Milho", stage: "flowering" },
          { choice: "Café", stage: "harvest" },
        ],
      })
    );
    expect(list).toEqual([
      { crop: "Milho", stage: "flowering" },
      { crop: "Café", stage: "harvest" },
    ]);
  });

  it("uses the typed name for 'outra'", () => {
    const list = cropList(answers({ crops: [{ choice: "outra", stage: "harvest" }], crop_custom: "  Mandioca " }));
    expect(list).toEqual([{ crop: "Mandioca", stage: "harvest" }]);
  });
});

describe("validateStep", () => {
  it("requires a choice on single-choice steps", () => {
    expect(validateStep("style", answers())).toBe("Escolha uma das opções.");
    expect(validateStep("style", answers({ work_style: "family" }))).toBeNull();
    expect(validateStep("areas", answers())).toBe("Escolha uma das opções.");
    expect(validateStep("producers", answers())).toBe("Escolha uma das opções.");
    expect(validateStep("soil", answers())).toBe("Escolha o tipo de solo.");
  });

  it("validates the identity step and asks for an organization only from technicians", () => {
    const base = { name: "José", phone: "19 99999-9999" };
    expect(validateStep("identity", answers({ ...base, work_style: "family" }))).toBeNull();
    expect(validateStep("identity", answers({ ...base, work_style: "technician" }))).toBe(
      "Digite o nome da sua organização."
    );
    expect(
      validateStep("identity", answers({ ...base, work_style: "technician", organization: "Coop" }))
    ).toBeNull();
    expect(validateStep("identity", answers({ name: "J", phone: base.phone }))).toBe("Digite seu nome.");
    expect(validateStep("identity", answers({ name: "José", phone: "12" }))).toBe(
      "Digite o DDD e o número, só números. Ex.: 19 99999-9999."
    );
  });

  it("validates the crop step", () => {
    expect(validateStep("crop", answers())).toBe("Escolha pelo menos uma cultura.");
    expect(validateStep("crop", answers({ crops: [{ choice: "Milho", stage: null }] }))).toBe(
      "Escolha em que fase está cada cultura."
    );
    expect(validateStep("crop", answers({ crops: [{ choice: "Milho", stage: "flowering" }] }))).toBeNull();
    expect(
      validateStep(
        "crop",
        answers({
          crops: [
            { choice: "Milho", stage: "flowering" },
            { choice: "Café", stage: null },
          ],
        })
      )
    ).toBe("Escolha em que fase está cada cultura.");
    expect(
      validateStep(
        "crop",
        answers({
          crops: [
            { choice: "Milho", stage: "flowering" },
            { choice: "Café", stage: "harvest" },
          ],
        })
      )
    ).toBeNull();
  });

  it("asks for the name of 'outra' and rejects a repeat of a listed crop", () => {
    const outra = { choice: "outra", stage: "harvest" as const };
    expect(validateStep("crop", answers({ crops: [outra] }))).toBe("Digite o nome da outra cultura.");
    expect(validateStep("crop", answers({ crops: [outra], crop_custom: "Mandioca" }))).toBeNull();
    expect(
      validateStep(
        "crop",
        answers({ crops: [{ choice: "Milho", stage: "harvest" }, outra], crop_custom: " milho " })
      )
    ).toBe("Essa cultura já está marcada.");
  });

  it("validates the location step", () => {
    expect(validateStep("location", answers())).toBe("Marque sua terra no mapa.");
    expect(validateStep("location", answers({ latitude: -22.9, longitude: -47.06 }))).toBe(
      "Escolha a cidade na busca."
    );
    expect(
      validateStep("location", answers({ latitude: -22.9, longitude: -47.06, city: "Campinas", state: "SP" }))
    ).toBeNull();
  });

  it("never blocks the optional concerns step", () => {
    expect(validateStep("concerns", answers())).toBeNull();
  });
});

describe("firstIncompleteIndex", () => {
  it("points at the first step that still has an error", () => {
    const steps = stepsFor("family");
    expect(firstIncompleteIndex(steps, answers({ work_style: "family" }))).toBe(1);
    expect(firstIncompleteIndex(steps, answers({ work_style: "family", areas_bucket: "1" }))).toBe(2);
  });

  it("returns the last step when everything is filled", () => {
    const steps = stepsFor("technician");
    const full = answers({
      work_style: "technician", name: "Ana", phone: "19 98888-7777",
      organization: "Coop", producers_bucket: "11-40",
    });
    expect(firstIncompleteIndex(steps, full)).toBe(steps.length - 1);
  });
});

describe("stepForField", () => {
  it("maps API field names back to the quiz step", () => {
    expect(stepForField("name")).toBe("identity");
    expect(stepForField("whatsapp_phone")).toBe("identity");
    expect(stepForField("email")).toBe("account");
    expect(stepForField("producer.city")).toBe("location");
    expect(stepForField("property.latitude")).toBe("location");
    expect(stepForField("property.crops")).toBe("crop");
    expect(stepForField("property.soil_type")).toBe("soil");
    expect(stepForField("concerns")).toBe("concerns");
    expect(stepForField("unknown")).toBeNull();
  });
});
