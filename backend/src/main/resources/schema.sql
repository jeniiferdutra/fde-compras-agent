CREATE TABLE IF NOT EXISTS requisitions (
    id BIGSERIAL PRIMARY KEY,
    request_text TEXT NOT NULL,
    response_text TEXT NOT NULL,
    total NUMERIC(12,2) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS requisition_items (
    id BIGSERIAL PRIMARY KEY,
    requisition_id BIGINT NOT NULL REFERENCES requisitions(id) ON DELETE CASCADE,
    product_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price NUMERIC(12,2) NOT NULL,
    line_total NUMERIC(12,2) NOT NULL
);
