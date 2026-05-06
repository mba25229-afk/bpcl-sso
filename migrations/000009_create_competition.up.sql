-- Competition periods (e.g. "Boost and Win March 2026")
CREATE TABLE competition_periods (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    period          DATE NOT NULL UNIQUE,   -- first day of competition month
    territory_code  TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'draft'
                    CHECK (status IN ('draft','active','published','archived')),
    published_at    TIMESTAMPTZ,
    created_by      UUID REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Parameter scores per dealer per competition
CREATE TABLE competition_scores (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    competition_id      UUID NOT NULL REFERENCES competition_periods(id),
    cc_number           VARCHAR(20) NOT NULL REFERENCES retail_outlets(cc_number),
    -- Raw input values
    ms_vol_kl           NUMERIC(14,3),
    ms_growth_pct       NUMERIC(8,4),
    ms_ta_gain_pp       NUMERIC(8,4),    -- percentage points, NOT percent
    hsd_vol_kl          NUMERIC(14,3),
    hsd_growth_pct      NUMERIC(8,4),
    hsd_ta_gain_pp      NUMERIC(8,4),
    qoc_count           NUMERIC(10,0),
    lubricants_value    NUMERIC(14,2),
    ufill_count         NUMERIC(10,0),
    speed_vol_kl        NUMERIC(14,3),
    -- Computed scores per parameter
    score_ms_vol        NUMERIC(6,3),
    score_ms_growth     NUMERIC(6,3),
    score_ms_ta_gain    NUMERIC(6,3),
    score_hsd_vol       NUMERIC(6,3),
    score_hsd_growth    NUMERIC(6,3),
    score_hsd_ta_gain   NUMERIC(6,3),
    score_qoc           NUMERIC(6,3),
    score_lubricants    NUMERIC(6,3),
    score_ufill         NUMERIC(6,3),
    score_speed         NUMERIC(6,3),
    score_cleanliness   NUMERIC(6,3),
    score_sangam        NUMERIC(6,3),
    score_google        NUMERIC(6,3),
    score_ips           NUMERIC(6,3),
    score_bonus         NUMERIC(6,3) DEFAULT 0,
    total_score         NUMERIC(8,3),
    rank                INTEGER,
    computed_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (competition_id, cc_number)
);

-- Audit scores (Cleanliness, Sangam, Google, IPS) — separate from performance
CREATE TABLE dealer_audit_scores (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cc_number           VARCHAR(20) NOT NULL REFERENCES retail_outlets(cc_number),
    period              DATE NOT NULL,
    cleanliness_grade   TEXT CHECK (cleanliness_grade IN
                            ('Excellent','Good','Average','Below Average','Poor')),
    cleanliness_status  TEXT NOT NULL DEFAULT 'not_audited'
                        CHECK (cleanliness_status IN ('audited','not_audited')),
    sangam_count        INTEGER,
    google_rating       NUMERIC(3,1) CHECK (google_rating BETWEEN 1.0 AND 5.0),
    ips_pct             NUMERIC(5,2) CHECK (ips_pct BETWEEN 0 AND 100),
    audit_status        TEXT NOT NULL DEFAULT 'not_audited'
                        CHECK (audit_status IN ('audited','not_audited','partial')),
    audited_by          UUID REFERENCES users(id),
    audited_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (cc_number, period)
);

-- Bonus marks with mandatory remarks (Bug #4 fix)
CREATE TABLE competition_bonus (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    competition_id  UUID NOT NULL REFERENCES competition_periods(id),
    cc_number       VARCHAR(20) NOT NULL REFERENCES retail_outlets(cc_number),
    bonus_marks     NUMERIC(5,2) NOT NULL CHECK (bonus_marks >= 0 AND bonus_marks <= 10),
    remarks         TEXT NOT NULL CHECK (length(trim(remarks)) >= 10),  -- Bug #4 enforcement
    awarded_by      UUID NOT NULL REFERENCES users(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (competition_id, cc_number)
);

-- Delhi Master market share data (all OMCs)
CREATE TABLE market_share_data (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    outlet_name     TEXT NOT NULL,
    cc_number       VARCHAR(20),           -- NULL for HPCL/IOCL outlets
    omc             TEXT NOT NULL CHECK (omc IN ('BPCL','HPCL','IOCL')),
    trading_area_id INTEGER NOT NULL REFERENCES trading_areas(id),
    period          DATE NOT NULL,
    ms_vol_kl       NUMERIC(14,3) DEFAULT 0,
    hsd_vol_kl      NUMERIC(14,3) DEFAULT 0,
    source          TEXT NOT NULL DEFAULT 'delhi_master',
    uploaded_file_id UUID REFERENCES uploaded_files(id),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (outlet_name, omc, trading_area_id, period)
);

-- Materialized totals per trading area per period
CREATE TABLE trading_area_totals (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trading_area_id INTEGER NOT NULL REFERENCES trading_areas(id),
    period          DATE NOT NULL,
    total_ms_kl     NUMERIC(14,3) DEFAULT 0,
    total_hsd_kl    NUMERIC(14,3) DEFAULT 0,
    bpcl_ms_kl      NUMERIC(14,3) DEFAULT 0,
    hpcl_ms_kl      NUMERIC(14,3) DEFAULT 0,
    iocl_ms_kl      NUMERIC(14,3) DEFAULT 0,
    bpcl_hsd_kl     NUMERIC(14,3) DEFAULT 0,
    hpcl_hsd_kl     NUMERIC(14,3) DEFAULT 0,
    iocl_hsd_kl     NUMERIC(14,3) DEFAULT 0,
    computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (trading_area_id, period)
);

CREATE INDEX idx_comp_scores_comp   ON competition_scores(competition_id);
CREATE INDEX idx_comp_scores_cc     ON competition_scores(cc_number);
CREATE INDEX idx_audit_scores_cc    ON dealer_audit_scores(cc_number, period);
CREATE INDEX idx_msd_trading_period ON market_share_data(trading_area_id, period);
CREATE INDEX idx_msd_cc             ON market_share_data(cc_number);
CREATE INDEX idx_tat_period         ON trading_area_totals(trading_area_id, period);
