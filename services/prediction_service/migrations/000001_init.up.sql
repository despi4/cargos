CREATE TABLE predictions (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    window_days INTEGER NOT NULL,
    recommendations JSONB NOT NULL,
    notification_sent BOOLEAN NOT NULL,
    CHECK(window_days BETWEEN 1 AND 365)
);

CREATE INDEX predictions_owner_page ON predictions(owner_id,created_at,id);

CREATE INDEX predictions_pending ON predictions(created_at) WHERE NOT notification_sent;
