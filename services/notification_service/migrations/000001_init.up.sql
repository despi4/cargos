CREATE TABLE notifications (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    event_id UUID NOT NULL,
    title TEXT NOT NULL,
    message TEXT NOT NULL,
    is_read BOOLEAN NOT NULL,
    UNIQUE(owner_id,event_id)
);

CREATE INDEX notifications_owner_page ON notifications(owner_id,created_at,id);
