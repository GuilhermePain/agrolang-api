CREATE INDEX alerts_active_lookup_idx ON alerts (property_id, alert_type) WHERE resolved_at IS NULL;
