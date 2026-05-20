-- Crystal v1 schema: cr_ prefix to coexist with legacy schema
-- BPCL Central Delhi Sales Area dealer scoring system

CREATE TABLE cr_dealers (
  cc_code        VARCHAR(10) PRIMARY KEY,
  ro_name        TEXT NOT NULL,
  area           TEXT NOT NULL DEFAULT 'Central Delhi',
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  dealer_email   TEXT,
  cc_email       TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cr_dealers_active ON cr_dealers (is_active) WHERE is_active = TRUE;

CREATE TABLE cr_competition_periods (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name           TEXT NOT NULL,
  month_year     DATE NOT NULL,
  total_slots    INT NOT NULL DEFAULT 40,
  is_active      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (month_year)
);

CREATE TABLE cr_monthly_targets (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code          VARCHAR(10) NOT NULL REFERENCES cr_dealers(cc_code),
  month_year       DATE NOT NULL,
  ufill_target     INT,
  qoc_target       INT,
  speed_kl         NUMERIC(10,3),
  ms_kl            NUMERIC(10,3),
  hsd_kl           NUMERIC(10,3),
  ms_ly            NUMERIC(10,3),
  hsd_ly           NUMERIC(10,3),
  dsw_available    BOOLEAN,
  nitrogen         BOOLEAN,
  mak_ge_target    NUMERIC(10,3),
  darpan_target    INT,
  coolant_lube_kl  NUMERIC(10,3),
  remarks          TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, month_year),
  FOREIGN KEY (month_year) REFERENCES cr_competition_periods(month_year)
);

CREATE INDEX idx_cr_targets_month ON cr_monthly_targets (month_year);

CREATE TABLE cr_daily_ufill (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES cr_dealers(cc_code),
  txn_date DATE NOT NULL,
  count    INT NOT NULL CHECK (count >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_cr_ufill_date ON cr_daily_ufill (txn_date);
CREATE INDEX idx_cr_ufill_cc_month ON cr_daily_ufill (cc_code, txn_date);

CREATE TABLE cr_daily_qoc (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES cr_dealers(cc_code),
  txn_date DATE NOT NULL,
  count    INT NOT NULL CHECK (count >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_cr_qoc_date ON cr_daily_qoc (txn_date);

CREATE TABLE cr_daily_ms (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES cr_dealers(cc_code),
  txn_date DATE NOT NULL,
  kl       NUMERIC(10,3) NOT NULL CHECK (kl >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_cr_ms_date ON cr_daily_ms (txn_date);

CREATE TABLE cr_daily_speed (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES cr_dealers(cc_code),
  txn_date DATE NOT NULL,
  kl       NUMERIC(10,3) NOT NULL CHECK (kl >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_cr_speed_date ON cr_daily_speed (txn_date);

CREATE TABLE cr_daily_hsd (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES cr_dealers(cc_code),
  txn_date DATE NOT NULL,
  kl       NUMERIC(10,3) NOT NULL CHECK (kl >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_cr_hsd_date ON cr_daily_hsd (txn_date);

CREATE TABLE cr_mak_ge_readings (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code        VARCHAR(10) NOT NULL REFERENCES cr_dealers(cc_code),
  reading_date   DATE NOT NULL,
  meter_reading  NUMERIC(12,2) NOT NULL CHECK (meter_reading >= 0),
  entered_by     TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, reading_date)
);

CREATE INDEX idx_cr_mak_ge_cc_month ON cr_mak_ge_readings (cc_code, reading_date);

CREATE TABLE cr_google_ratings (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code       VARCHAR(10) NOT NULL REFERENCES cr_dealers(cc_code),
  snapshot_date DATE NOT NULL,
  rating        NUMERIC(3,1) NOT NULL CHECK (rating BETWEEN 1.0 AND 5.0),
  review_count  INT NOT NULL CHECK (review_count >= 0),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, snapshot_date)
);

CREATE TABLE cr_scoring_params (
  id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  month_year               DATE NOT NULL REFERENCES cr_competition_periods(month_year),
  metric_key               TEXT NOT NULL,
  display_name             TEXT NOT NULL,
  max_marks                NUMERIC(5,2) NOT NULL,
  negative_scale_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
  is_active                BOOLEAN NOT NULL DEFAULT TRUE,
  sort_order               INT NOT NULL,
  UNIQUE (month_year, metric_key)
);

CREATE TABLE cr_dealer_scores (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code         VARCHAR(10) NOT NULL REFERENCES cr_dealers(cc_code),
  month_year      DATE NOT NULL,
  metric_key      TEXT NOT NULL,
  actual_value    NUMERIC(15,4),
  rank_in_group   INT,
  marks_scored    NUMERIC(6,3),
  max_marks       NUMERIC(5,2),
  computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, month_year, metric_key),
  FOREIGN KEY (month_year) REFERENCES cr_competition_periods(month_year)
);

CREATE INDEX idx_cr_scores_month ON cr_dealer_scores (month_year);
CREATE INDEX idx_cr_scores_cc ON cr_dealer_scores (cc_code, month_year);

CREATE TABLE cr_sangam_data (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code      VARCHAR(10) NOT NULL REFERENCES cr_dealers(cc_code),
  month_year   DATE NOT NULL,
  cert_count   INT NOT NULL DEFAULT 0,
  status       TEXT,
  remarks      TEXT,
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, month_year)
);

CREATE VIEW cr_dealer_total_scores AS
SELECT
  ds.cc_code,
  d.ro_name,
  ds.month_year,
  SUM(ds.marks_scored)   AS total_marks,
  RANK() OVER (
    PARTITION BY ds.month_year
    ORDER BY SUM(ds.marks_scored) DESC
  )                      AS rank_overall,
  ds.computed_at
FROM cr_dealer_scores ds
JOIN cr_dealers d ON d.cc_code = ds.cc_code
WHERE ds.marks_scored IS NOT NULL
GROUP BY ds.cc_code, d.ro_name, ds.month_year, ds.computed_at;
