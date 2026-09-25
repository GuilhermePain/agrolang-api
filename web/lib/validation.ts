import type { CropStage } from "./api";

const CROP_STAGES: CropStage[] = ["germination", "flowering", "harvest"];

export type ProducerFormValues = {
  name: string;
  whatsapp_phone: string;
  city: string;
  state: string;
};

export type ProducerFormErrors = Partial<Record<keyof ProducerFormValues, string>>;

export function validateProducer(values: ProducerFormValues): ProducerFormErrors {
  const errors: ProducerFormErrors = {};
  if (!values.name.trim()) errors.name = "Informe o nome do produtor.";
  if (!values.whatsapp_phone.trim()) errors.whatsapp_phone = "Informe o telefone do WhatsApp.";
  if (!values.city.trim()) errors.city = "Informe a cidade.";
  if (!values.state.trim()) errors.state = "Informe o estado.";
  return errors;
}

export type PropertyFormValues = {
  producer_id: string;
  latitude: string;
  longitude: string;
  crop: string;
  soil_type: string;
  crop_stage: string;
};

export type PropertyFormErrors = Partial<Record<keyof PropertyFormValues, string>>;

export function validateProperty(values: PropertyFormValues): PropertyFormErrors {
  const errors: PropertyFormErrors = {};

  if (!values.producer_id.trim()) errors.producer_id = "Selecione um produtor.";

  const lat = Number(values.latitude);
  if (values.latitude.trim() === "" || Number.isNaN(lat)) {
    errors.latitude = "Latitude deve ser um número.";
  } else if (lat < -90 || lat > 90) {
    errors.latitude = "Latitude deve estar entre -90 e 90.";
  }

  const lon = Number(values.longitude);
  if (values.longitude.trim() === "" || Number.isNaN(lon)) {
    errors.longitude = "Longitude deve ser um número.";
  } else if (lon < -180 || lon > 180) {
    errors.longitude = "Longitude deve estar entre -180 e 180.";
  }

  if (!values.crop.trim()) errors.crop = "Informe a cultura.";
  if (!values.soil_type.trim()) errors.soil_type = "Informe o tipo de solo.";
  if (!CROP_STAGES.includes(values.crop_stage as CropStage)) {
    errors.crop_stage = "Selecione uma fase válida.";
  }

  return errors;
}
