CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE producers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    whatsapp_phone TEXT NOT NULL,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE crop_stage AS ENUM ('germination', 'flowering', 'harvest');

CREATE TABLE properties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    producer_id UUID NOT NULL REFERENCES producers(id) ON DELETE CASCADE,
    location GEOGRAPHY(POINT, 4326) NOT NULL,
    crop TEXT NOT NULL,
    soil_type TEXT NOT NULL,
    crop_stage crop_stage NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX properties_location_idx ON properties USING GIST (location);
CREATE INDEX properties_producer_id_idx ON properties (producer_id);
