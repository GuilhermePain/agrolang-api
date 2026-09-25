import { describe, expect, it } from "vitest";
import { validateProducer, validateProperty } from "./validation";

describe("validateProducer", () => {
  it("returns no errors for a fully filled producer", () => {
    const errors = validateProducer({
      name: "Joao",
      whatsapp_phone: "5511999999999",
      city: "Campinas",
      state: "SP",
    });
    expect(errors).toEqual({});
  });

  it("returns an error for a missing name", () => {
    const errors = validateProducer({
      name: "",
      whatsapp_phone: "5511999999999",
      city: "Campinas",
      state: "SP",
    });
    expect(errors.name).toBe("Informe o nome do produtor.");
  });

  it("returns an error for a missing whatsapp phone", () => {
    const errors = validateProducer({
      name: "Joao",
      whatsapp_phone: "  ",
      city: "Campinas",
      state: "SP",
    });
    expect(errors.whatsapp_phone).toBe("Informe o telefone do WhatsApp.");
  });

  it("returns errors for missing city and state", () => {
    const errors = validateProducer({
      name: "Joao",
      whatsapp_phone: "5511999999999",
      city: "",
      state: "",
    });
    expect(errors.city).toBe("Informe a cidade.");
    expect(errors.state).toBe("Informe o estado.");
  });
});

describe("validateProperty", () => {
  const valid = {
    producer_id: "prod-1",
    latitude: "-22.9",
    longitude: "-47.1",
    crop: "Milho",
    soil_type: "Argiloso",
    crop_stage: "flowering",
  };

  it("returns no errors for a fully filled, valid property", () => {
    expect(validateProperty(valid)).toEqual({});
  });

  it("returns an error when no producer is selected", () => {
    const errors = validateProperty({ ...valid, producer_id: "" });
    expect(errors.producer_id).toBe("Selecione um produtor.");
  });

  it("returns an error when latitude is not a number", () => {
    const errors = validateProperty({ ...valid, latitude: "abc" });
    expect(errors.latitude).toBe("Latitude deve ser um número.");
  });

  it("returns an error when latitude is out of range", () => {
    const errors = validateProperty({ ...valid, latitude: "120" });
    expect(errors.latitude).toBe("Latitude deve estar entre -90 e 90.");
  });

  it("returns an error when longitude is not a number", () => {
    const errors = validateProperty({ ...valid, longitude: "abc" });
    expect(errors.longitude).toBe("Longitude deve ser um número.");
  });

  it("returns an error when longitude is out of range", () => {
    const errors = validateProperty({ ...valid, longitude: "200" });
    expect(errors.longitude).toBe("Longitude deve estar entre -180 e 180.");
  });

  it("returns errors for missing crop and soil type", () => {
    const errors = validateProperty({ ...valid, crop: "", soil_type: "" });
    expect(errors.crop).toBe("Informe a cultura.");
    expect(errors.soil_type).toBe("Informe o tipo de solo.");
  });

  it("returns an error for an invalid crop stage", () => {
    const errors = validateProperty({ ...valid, crop_stage: "bloom" });
    expect(errors.crop_stage).toBe("Selecione uma fase válida.");
  });
});
