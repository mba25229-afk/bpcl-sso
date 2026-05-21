# Chunk 1 — Database Schema
**Status:** ACTIVE  
**Depends on:** Chunk 0 (business rules locked ✓)  
**Blocks:** Chunk 2 (Ingest API), Chunk 3 (Scoring Engine)

---

## Context

This schema replaces the MAY_DATA.xlsx manual tracking system for BPCL Central Delhi Sales Area. 41 active dealers. Monthly data cycle. Scoring is **peer-relative rank-based**, not target-vs-actual.

Schema is Postgres. All timestamps are UTC. `cc_code` is the canonical dealer identifier — matches BPCL system codes.

---

## Migration

### Table: `dealers`

```sql
CREATE TABLE dealers (
  cc_code        VARCHAR(10) PRIMARY KEY,  -- e.g. '112385'
  ro_name        TEXT NOT NULL,
  area           TEXT NOT NULL DEFAULT 'Central Delhi',
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  dealer_email   TEXT,
  cc_email       TEXT,                     -- Area manager CC email
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dealers_active ON dealers (is_active) WHERE is_active = TRUE;
```

Seed from `Dealer_Email_Map` sheet — 40 active dealers.

---

### Table: `competition_periods`

```sql
CREATE TABLE competition_periods (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name           TEXT NOT NULL,            -- e.g. 'Boost and Win May 2026'
  month_year     DATE NOT NULL,            -- First day of month: 2026-05-01
  total_slots    INT NOT NULL DEFAULT 40,  -- Fixed n for rank formula
  is_active      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (month_year)
);
```

Only one `competition_period` active at a time. Admin activates/closes.

---

### Table: `monthly_targets`

Targets entered once per month per dealer by admin. Prorated daily for MTD comparisons.

```sql
CREATE TABLE monthly_targets (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code          VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  month_year       DATE NOT NULL,          -- First day of month
  ufill_target     INT,                    -- UFill transaction count
  qoc_target       INT,                    -- Quick Oil Change count
  speed_kl         NUMERIC(10,3),          -- Speed product KL
  ms_kl            NUMERIC(10,3),          -- Total MS (Petrol+Speed) KL
  hsd_kl           NUMERIC(10,3),          -- HSD (Diesel) KL
  ms_ly            NUMERIC(10,3),          -- MS Last Year KL (for growth %)
  hsd_ly           NUMERIC(10,3),          -- HSD Last Year KL (for growth %)
  dsw_available    BOOLEAN,               -- DSW (Digital Service Worker) available
  nitrogen         BOOLEAN,               -- Nitrogen tyre inflation available
  mak_ge_target    NUMERIC(10,3),          -- MAK GE sales target KL
  darpan_target    INT,                    -- Darpan target
  coolant_lube_kl  NUMERIC(10,3),          -- Coolant/Other Lube sales target (litres)
  remarks          TEXT,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, month_year),
  FOREIGN KEY (month_year) REFERENCES competition_periods(month_year)
);

CREATE INDEX idx_targets_month ON monthly_targets (month_year);
```

---

### Table: `daily_ufill`

```sql
CREATE TABLE daily_ufill (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  txn_date DATE NOT NULL,
  count    INT NOT NULL CHECK (count >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_ufill_date ON daily_ufill (txn_date);
CREATE INDEX idx_ufill_cc_month ON daily_ufill (cc_code, DATE_TRUNC('month', txn_date));
```

---

### Table: `daily_qoc`

```sql
CREATE TABLE daily_qoc (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  txn_date DATE NOT NULL,
  count    INT NOT NULL CHECK (count >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_qoc_date ON daily_qoc (txn_date);
```

---

### Table: `daily_ms`

Regular petrol (MS) sales — does NOT include Speed product.

```sql
CREATE TABLE daily_ms (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  txn_date DATE NOT NULL,
  kl       NUMERIC(10,3) NOT NULL CHECK (kl >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_ms_date ON daily_ms (txn_date);
```

---

### Table: `daily_speed`

Speed + Speed100 combined (consolidated per business rule). Used for:
1. `speed_vol` scoring metric (5 marks, standalone)
2. Adding to `daily_ms` for total `ms_absolute_vol` scoring (10 marks)

```sql
CREATE TABLE daily_speed (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  txn_date DATE NOT NULL,
  kl       NUMERIC(10,3) NOT NULL CHECK (kl >= 0),  -- Speed + Speed100 sum
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_speed_date ON daily_speed (txn_date);
```

---

### Table: `daily_hsd`

HSD (Diesel) sales.

```sql
CREATE TABLE daily_hsd (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  txn_date DATE NOT NULL,
  kl       NUMERIC(10,3) NOT NULL CHECK (kl >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_hsd_date ON daily_hsd (txn_date);
```

---

### Table: `mak_ge_readings`

Weekly meter readings. Delta = last − first within month = total sales.

```sql
CREATE TABLE mak_ge_readings (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code        VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  reading_date   DATE NOT NULL,
  meter_reading  NUMERIC(12,2) NOT NULL CHECK (meter_reading >= 0),
  entered_by     TEXT,                  -- admin user ID
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, reading_date)
);

CREATE INDEX idx_mak_ge_cc_month ON mak_ge_readings (cc_code, DATE_TRUNC('month', reading_date));
```

---

### Table: `google_ratings`

Periodic snapshots — captured ~3 times per month by admin.

```sql
CREATE TABLE google_ratings (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code       VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  snapshot_date DATE NOT NULL,
  rating        NUMERIC(3,1) NOT NULL CHECK (rating BETWEEN 1.0 AND 5.0),
  review_count  INT NOT NULL CHECK (review_count >= 0),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, snapshot_date)
);
```

Scoring uses latest snapshot within the competition month's date range.

---

### Table: `scoring_params`

Configurable per competition period. Admin can adjust weights without code change.

```sql
CREATE TABLE scoring_params (
  id                       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  month_year               DATE NOT NULL REFERENCES competition_periods(month_year),
  metric_key               TEXT NOT NULL,
  display_name             TEXT NOT NULL,
  max_marks                NUMERIC(5,2) NOT NULL,
  negative_scale_enabled   BOOLEAN NOT NULL DEFAULT FALSE,
  is_active                BOOLEAN NOT NULL DEFAULT TRUE,
  sort_order               INT NOT NULL,
  UNIQUE (month_year, metric_key)
);
```

**Seed data for every new competition period:**

```sql
INSERT INTO scoring_params (month_year, metric_key, display_name, max_marks, negative_scale_enabled, is_active, sort_order) VALUES
  (:month, 'ms_absolute_vol',   'MS Volume (KL)',             10, FALSE, TRUE,  1),
  (:month, 'ms_growth_pct',     'MS Growth %',                 5, TRUE,  TRUE,  2),
  (:month, 'ms_ms_gain_pct',    'MS Market Share Gain %',     10, TRUE,  FALSE, 3),  -- inactive
  (:month, 'hsd_absolute_vol',  'HSD Volume (KL)',            10, FALSE, TRUE,  4),
  (:month, 'hsd_growth_pct',    'HSD Growth %',                5, TRUE,  TRUE,  5),
  (:month, 'hsd_ms_gain_pct',   'HSD Market Share Gain %',   10, TRUE,  FALSE, 6),  -- inactive
  (:month, 'oil_change_count',  'Oil Change (QOC)',           10, FALSE, TRUE,  7),
  (:month, 'mak_ge_sales',      'MAK / GE Sales',            10, FALSE, TRUE,  8),
  (:month, 'ufill_txns',        'UFill Transactions',          5, FALSE, TRUE,  9),
  (:month, 'speed_vol',         'Speed Volume (KL)',           5, FALSE, TRUE,  10),
  (:month, 'cleanliness_audit', 'Cleanliness Audit',           5, FALSE, FALSE, 11), -- inactive (no source)
  (:month, 'sangam_certs',      'Sangam Certifications',       5, FALSE, TRUE,  12),
  (:month, 'google_rating',     'Google Rating',               5, FALSE, TRUE,  13),
  (:month, 'ips_pct',           'IPS %',                       5, FALSE, FALSE, 14), -- inactive (ALP not in scope)
  (:month, 'bonus',             'Bonus',                      10, FALSE, TRUE,  15);
```

---

### Table: `dealer_scores`

Computed scores — written by scoring engine at end of month or on-demand.

```sql
CREATE TABLE dealer_scores (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code         VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  month_year      DATE NOT NULL,
  metric_key      TEXT NOT NULL,
  actual_value    NUMERIC(15,4),       -- Raw input value (KL, count, %, etc.)
  rank_in_group   INT,                 -- 1 = best, 40 = worst
  marks_scored    NUMERIC(6,3),        -- NULL if metric inactive or data missing
  max_marks       NUMERIC(5,2),
  computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, month_year, metric_key),
  FOREIGN KEY (month_year) REFERENCES competition_periods(month_year)
);

CREATE INDEX idx_scores_month ON dealer_scores (month_year);
CREATE INDEX idx_scores_cc ON dealer_scores (cc_code, month_year);
```

---

### View: `dealer_total_scores`

```sql
CREATE VIEW dealer_total_scores AS
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
FROM dealer_scores ds
JOIN dealers d ON d.cc_code = ds.cc_code
WHERE ds.marks_scored IS NOT NULL
GROUP BY ds.cc_code, d.ro_name, ds.month_year, ds.computed_at;
```

---

### Table: `sangam_data`

```sql
CREATE TABLE sangam_data (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code      VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  month_year   DATE NOT NULL,
  cert_count   INT NOT NULL DEFAULT 0,
  status       TEXT,
  remarks      TEXT,
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, month_year)
);
```

---

## Invariants (enforced in application layer + DB constraints)

1. **cc_code must exist in `dealers`** — no ghost entries anywhere
2. **`txn_date` must fall within an active `competition_period`** — validated in ingest API
3. **`daily_*` tables use UPSERT** — duplicate date inserts update, never error
4. **`mak_ge_readings` requires ≥ 2 entries/month** to produce a score
5. **`scoring_params.is_active = FALSE` metrics return `NULL` marks** — excluded from total
6. **`total_slots = 40` in `competition_periods`** — fixed, do not make dynamic
7. **Speed + Speed100 sum BEFORE insert** into `daily_speed` — single row per dealer per day

---

## Rollback

```sql
DROP VIEW IF EXISTS dealer_total_scores;
DROP TABLE IF EXISTS dealer_scores CASCADE;
DROP TABLE IF EXISTS scoring_params CASCADE;
DROP TABLE IF EXISTS sangam_data CASCADE;
DROP TABLE IF EXISTS google_ratings CASCADE;
DROP TABLE IF EXISTS mak_ge_readings CASCADE;
DROP TABLE IF EXISTS daily_hsd CASCADE;
DROP TABLE IF EXISTS daily_speed CASCADE;
DROP TABLE IF EXISTS daily_ms CASCADE;
DROP TABLE IF EXISTS daily_qoc CASCADE;
DROP TABLE IF EXISTS daily_ufill CASCADE;
DROP TABLE IF EXISTS monthly_targets CASCADE;
DROP TABLE IF EXISTS competition_periods CASCADE;
DROP TABLE IF EXISTS dealers CASCADE;
```

---

## Acceptance Criteria

- [ ] Migration runs clean on local Supabase instance
- [ ] Rollback script executes without error on clean state
- [ ] Seed 40 dealers from `Dealer_Email_Map` — all rows insert
- [ ] Seed May 2026 scoring_params — 15 rows, 10 active
- [ ] Insert 6 days of MAY_DATA daily rows (UFILL, QOC, MS, HSD, SPEED) for all 40 dealers
- [ ] MTD sums from DB match Excel MTD column within ±0.01 KL
- [ ] `dealer_total_scores` view returns ranks matching v4 Scores sheet for same input data

---

## What Is NOT in This Chunk

| Feature | Reason |
|---|---|
| ALP / IPS data | Out of scope (Chunk 4+) |
| UFILL Issues | Out of scope |
| DSW contact records | Out of scope |
| Cleanliness audit source | No automated data feed — admin manual later |
| MS Market Share Gain % | Requires territory-level market data not yet available |
| Historical data migration | Separate migration task post-Chunk 2 |
