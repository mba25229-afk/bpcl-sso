CREATE TABLE targets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cc_number       VARCHAR(20) NOT NULL REFERENCES retail_outlets(cc_number),
    product_id      SMALLINT    NOT NULL REFERENCES products(id),
    period          DATE        NOT NULL,
    target_value    NUMERIC(14,2) NOT NULL CHECK (target_value >= 0),
    set_by          UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cc_number, product_id, period)
);

CREATE INDEX idx_targets_cc_period ON targets(cc_number, period);
