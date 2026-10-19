CREATE TABLE suppliers (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    lead_days INTEGER NOT NULL,
    active BOOLEAN NOT NULL,
    CHECK(lead_days BETWEEN 0 AND 365)
);

CREATE INDEX suppliers_owner_page ON suppliers(owner_id,created_at,id);

CREATE TABLE orders (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    product_id UUID NOT NULL,
    supplier_id UUID NOT NULL,
    quantity BIGINT NOT NULL,
    unit_price BIGINT NOT NULL,
    currency TEXT NOT NULL,
    status TEXT NOT NULL,
    expected_at TIMESTAMPTZ NOT NULL,
    CHECK(status IN ('draft','ordered','receiving','received','cancelled')),
    CHECK(quantity BETWEEN 1 AND 1000000000),
    CHECK(unit_price BETWEEN 0 AND 1000000000),
    FOREIGN KEY(supplier_id) REFERENCES suppliers(id)
);

CREATE INDEX orders_owner_page ON orders(owner_id,created_at,id);

CREATE INDEX orders_receiving ON orders(created_at) WHERE status='receiving';
