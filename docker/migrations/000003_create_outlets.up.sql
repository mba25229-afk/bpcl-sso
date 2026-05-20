CREATE TABLE trading_areas (
    id              SERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    district        TEXT NOT NULL,
    state           TEXT NOT NULL DEFAULT 'Delhi'
);

CREATE TABLE retail_outlets (
    cc_number       VARCHAR(20) PRIMARY KEY,
    name            TEXT NOT NULL,
    location        TEXT,
    district        TEXT,
    state           TEXT DEFAULT 'Delhi',
    rank            CHAR(2),
    territory_code  TEXT,
    trading_area_id INTEGER REFERENCES trading_areas(id),
    outlet_type     TEXT NOT NULL DEFAULT 'regular'
                    CHECK (outlet_type IN ('regular','adhoc','coco','dealer_owned')),
    ro_manager_id   UUID REFERENCES users(id),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_outlets_territory    ON retail_outlets(territory_code);
CREATE INDEX idx_outlets_type         ON retail_outlets(outlet_type);
CREATE INDEX idx_outlets_trading_area ON retail_outlets(trading_area_id);
