CREATE TABLE demand_records (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    product_id UUID NOT NULL,
    quantity BIGINT NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    note TEXT NOT NULL,
    CHECK(quantity BETWEEN 1 AND 1000000000)
);

CREATE INDEX demand_records_owner_page ON demand_records(owner_id,created_at,id);
