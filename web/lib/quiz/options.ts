import type {
  AreasBucket,
  Concern,
  CropStage,
  ProducersBucket,
  SoilType,
  WorkStyle,
} from "./types";

export type Option<T extends string> = { value: T; label: string; hint?: string };

export const STYLE_OPTIONS: Option<WorkStyle>[] = [
  { value: "family", label: "Planto na minha terra, com a família" },
  { value: "multi_area", label: "Tenho várias áreas e contrato gente" },
  { value: "technician", label: "Sou agrônomo ou técnico de cooperativa" },
];

export const AREAS_OPTIONS: Option<AreasBucket>[] = [
  { value: "1", label: "Uma só" },
  { value: "2-5", label: "De 2 a 5" },
  { value: "6+", label: "Mais de 5" },
];

export const PRODUCERS_OPTIONS: Option<ProducersBucket>[] = [
  { value: "ate-10", label: "Até 10 produtores" },
  { value: "11-40", label: "De 11 a 40" },
  { value: "mais-40", label: "Mais de 40" },
];

export const CROP_OPTIONS = ["Milho", "Tomate", "Café", "Soja", "Feijão"] as const;
export const CROP_OTHER = "outra";

export const STAGE_OPTIONS: Option<CropStage>[] = [
  { value: "germination", label: "Acabei de plantar", hint: "Germinação" },
  { value: "flowering", label: "Está florindo", hint: "Floração" },
  { value: "harvest", label: "Perto de colher", hint: "Colheita" },
];

export const SOIL_OPTIONS: Option<SoilType>[] = [
  { value: "argiloso", label: "Argiloso", hint: "Terra vermelha ou barrenta, gruda quando molhada" },
  { value: "arenoso", label: "Arenoso", hint: "Solta e clara, a água escorre rápido" },
  { value: "misto", label: "Misto", hint: "Meio a meio, ou não é bem de um tipo só" },
  { value: "nao_informado", label: "Não sei", hint: "Tudo bem, dá para ajustar depois" },
];

export const CONCERN_OPTIONS: Option<Concern>[] = [
  { value: "frost", label: "Geada", hint: "Frio forte que queima a planta" },
  { value: "severe_water_stress", label: "Seca", hint: "Falta de chuva e ar muito seco" },
  { value: "heat_wave", label: "Calor forte", hint: "Dias seguidos muito quentes" },
  { value: "heavy_rain", label: "Chuva forte", hint: "Temporal, alagamento nas áreas baixas" },
];
