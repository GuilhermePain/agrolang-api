CREATE TABLE notification_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    alert_id UUID REFERENCES alerts(id) ON DELETE SET NULL,
    producer_id UUID NOT NULL REFERENCES producers(id) ON DELETE CASCADE,
    property_id UUID REFERENCES properties(id) ON DELETE CASCADE,
    phone TEXT NOT NULL,
    message TEXT NOT NULL,
    status TEXT NOT NULL, -- 'sent', 'delivered', 'read', 'failed'
    sent_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    delivered_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    error_message TEXT
);

CREATE INDEX notification_logs_producer_id_idx ON notification_logs (producer_id);
CREATE INDEX notification_logs_sent_at_idx ON notification_logs (sent_at DESC);
CREATE INDEX notification_logs_phone_idx ON notification_logs (phone);

