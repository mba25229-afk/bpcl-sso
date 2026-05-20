CREATE TABLE performance_records (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cc_number        VARCHAR(20) NOT NULL REFERENCES retail_outlets(cc_number),
    product_id       SMALLINT    NOT NULL REFERENCES products(id),
    period           DATE        NOT NULL,  -- stored as first day of month
    achieved         NUMERIC(14,2),
    last_year        NUMERIC(14,2),
    volume_kl        NUMERIC(14,3),         -- fuel only (MS/HSD/SPEED)
    source           TEXT NOT NULL DEFAULT 'manual'
                     CHECK (source IN ('manual','excel_upload','api_sync')),
    uploaded_file_id UUID,                  -- FK added after uploaded_files table exists
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cc_number, product_id, period)
);

CREATE INDEX idx_perf_cc_period    ON performance_records(cc_number, period);
CREATE INDEX idx_perf_product      ON performance_records(product_id);
CREATE INDEX idx_perf_period       ON performance_records(period);
