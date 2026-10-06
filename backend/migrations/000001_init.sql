CREATE TABLE IF NOT EXISTS items (
    id BIGSERIAL PRIMARY KEY,
    sku VARCHAR(64) NOT NULL,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    unit VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_items_sku_active
    ON items (sku)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_items_category
    ON items (category)
    WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS locations (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(64) NOT NULL UNIQUE,
    zone VARCHAR(100) NOT NULL,
    type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_locations_zone
    ON locations (zone);

CREATE TABLE IF NOT EXISTS stock (
    id BIGSERIAL PRIMARY KEY,
    item_id BIGINT NOT NULL REFERENCES items(id),
    location_id BIGINT NOT NULL REFERENCES locations(id),
    qty INTEGER NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_stock_qty_non_negative CHECK (qty >= 0),
    CONSTRAINT uq_stock_item_location UNIQUE (item_id, location_id)
);

CREATE INDEX IF NOT EXISTS idx_stock_item_id
    ON stock (item_id);

CREATE INDEX IF NOT EXISTS idx_stock_location_id
    ON stock (location_id);
