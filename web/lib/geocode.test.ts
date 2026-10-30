import { describe, expect, it, vi } from "vitest";
import { parseGeocodingResponse, searchCities } from "./geocode";

const response = {
  results: [
    { name: "Campinas", admin1: "São Paulo", latitude: -22.90556, longitude: -47.06083, country_code: "BR" },
    { name: "Campinas", admin1: "Distrito do Porto", latitude: 41.16667, longitude: -8.65, country_code: "PT" },
    { name: "Lugar Sem Estado", admin1: "Estado Inexistente", latitude: 1, longitude: 2, country_code: "BR" },
  ],
};

describe("parseGeocodingResponse", () => {
  it("keeps only Brazilian results whose state maps to a UF", () => {
    expect(parseGeocodingResponse(response)).toEqual([
      {
        name: "Campinas",
        state: "SP",
        latitude: -22.90556,
        longitude: -47.06083,
        label: "Campinas, SP",
      },
    ]);
  });

  it("returns an empty list for unexpected shapes", () => {
    expect(parseGeocodingResponse(null)).toEqual([]);
    expect(parseGeocodingResponse({})).toEqual([]);
    expect(parseGeocodingResponse({ results: "x" })).toEqual([]);
  });
});

describe("searchCities", () => {
  it("does not call the API for very short queries", async () => {
    const fetchImpl = vi.fn();
    expect(await searchCities("c", fetchImpl as unknown as typeof fetch)).toEqual([]);
    expect(fetchImpl).not.toHaveBeenCalled();
  });

  it("queries Open-Meteo restricted to Brazil and parses the result", async () => {
    const fetchImpl = vi.fn().mockResolvedValue({ ok: true, json: async () => response });

    const cities = await searchCities("Campinas", fetchImpl as unknown as typeof fetch);

    const url = String(fetchImpl.mock.calls[0][0]);
    expect(url).toContain("geocoding-api.open-meteo.com/v1/search");
    expect(url).toContain("name=Campinas");
    expect(url).toContain("countryCode=BR");
    expect(cities).toHaveLength(1);
  });

  it("throws when the API fails", async () => {
    const fetchImpl = vi.fn().mockResolvedValue({ ok: false, status: 500 });
    await expect(searchCities("Campinas", fetchImpl as unknown as typeof fetch)).rejects.toThrow();
  });
});
