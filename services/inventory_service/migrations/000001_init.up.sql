CREATE TABLE products (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    sku TEXT NOT NULL,
    name TEXT NOT NULL,
    unit TEXT NOT NULL,
    stock BIGINT NOT NULL,
    min_stock BIGINT NOT NULL,
    target_stock BIGINT NOT NULL,
    active BOOLEAN NOT NULL,
    UNIQUE(owner_id,sku),
    CHECK(stock BETWEEN 0 AND 1000000000),
    CHECK(min_stock >= 0 AND target_stock >= min_stock AND target_stock <= 1000000000)
);

CREATE INDEX products_owner_page ON products(owner_id,created_at,id);

CREATE TABLE movements (
    id UUID PRIMARY KEY,
    owner_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    product_id UUID NOT NULL,
    quantity BIGINT NOT NULL,
    reason TEXT NOT NULL,
    reference TEXT NOT NULL,
    balance BIGINT NOT NULL,
    UNIQUE(owner_id,reference),
    FOREIGN KEY(product_id) REFERENCES products(id),
    CHECK(quantity != 0),
    CHECK(balance >= 0)
);

CREATE INDEX movements_owner_page ON movements(owner_id,created_at,id);
