export type CityResult = {
  name: string;
  state: string;
  latitude: number;
  longitude: number;
  label: string;
};

// Open-Meteo returns the state name (admin1), not the UF code.
const STATE_UF: Record<string, string> = {
  Acre: "AC",
  Alagoas: "AL",
  Amapá: "AP",
  Amazonas: "AM",
  Bahia: "BA",
  Ceará: "CE",
  "Distrito Federal": "DF",
  "Espírito Santo": "ES",
  Goiás: "GO",
  Maranhão: "MA",
  "Mato Grosso": "MT",
  "Mato Grosso do Sul": "MS",
  "Minas Gerais": "MG",
  Pará: "PA",
  Paraíba: "PB",
  Paraná: "PR",
  Pernambuco: "PE",
  Piauí: "PI",
  "Rio de Janeiro": "RJ",
  "Rio Grande do Norte": "RN",
  "Rio Grande do Sul": "RS",
  Rondônia: "RO",
  Roraima: "RR",
  "Santa Catarina": "SC",
  "São Paulo": "SP",
  Sergipe: "SE",
  Tocantins: "TO",
};

const GEOCODING_URL = "https://geocoding-api.open-meteo.com/v1/search";

export function parseGeocodingResponse(json: unknown): CityResult[] {
  if (typeof json !== "object" || json === null) return [];
  const results = (json as { results?: unknown }).results;
  if (!Array.isArray(results)) return [];

  const cities: CityResult[] = [];
  for (const item of results) {
    if (typeof item !== "object" || item === null) continue;
    const r = item as Record<string, unknown>;
    if (r.country_code !== "BR") continue;
    if (typeof r.name !== "string" || typeof r.admin1 !== "string") continue;
    if (typeof r.latitude !== "number" || typeof r.longitude !== "number") continue;
    const state = STATE_UF[r.admin1];
    if (!state) continue;
    cities.push({
      name: r.name,
      state,
      latitude: r.latitude,
      longitude: r.longitude,
      label: `${r.name}, ${state}`,
    });
  }
  return cities;
}

export async function searchCities(query: string, fetchImpl: typeof fetch = fetch): Promise<CityResult[]> {
  const name = query.trim();
  if (name.length < 2) return [];

  const url = `${GEOCODING_URL}?name=${encodeURIComponent(name)}&count=8&language=pt&countryCode=BR`;
  const res = await fetchImpl(url);
  if (!res.ok) throw new Error(`geocoding failed: ${res.status}`);
  return parseGeocodingResponse(await res.json());
}
