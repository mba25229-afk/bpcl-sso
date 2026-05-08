# Crystal Chunk 1 — Database Schema Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the Crystal v1 database schema (dealers, daily metrics, scoring, view) as migration 000010, with dealer seed and May 2026 test data that satisfies all acceptance criteria.

**Architecture:** New parallel schema on top of existing bpcl_portal DB. Crystal tables use `cc_code VARCHAR(10)` as canonical dealer key and `month_year DATE` (first-of-month) as period key — distinct from the legacy `cc_number`/`period` column names so both schemas coexist without collision. The scoring system is normalized: one row per metric per dealer per month in `dealer_scores`, aggregated by `dealer_total_scores` view.

**Tech Stack:** PostgreSQL 15+ (local Supabase), golang-migrate for migration runner, `psql` for seed/verify scripts.

---

## File Map

| Action | Path | Responsibility |
|--------|------|----------------|
| Create | `migrations/000010_crystal_schema.up.sql` | All 10 tables + 1 view + all indexes |
| Create | `migrations/000010_crystal_schema.down.sql` | Exact rollback (drop view + tables in FK order) |
| Create | `scripts/seed_crystal_dealers.sql` | 40 active dealers with cc_code/ro_name/email |
| Create | `scripts/seed_crystal_may2026.sql` | competition_period + 15 scoring_params + 6 days test data |
| Create | `scripts/verify_crystal_chunk1.sql` | Verification queries for all 7 acceptance criteria |

---

## Task 1: Create the up migration

**Files:**
- Create: `migrations/000010_crystal_schema.up.sql`

- [ ] **Step 1: Write the migration file**

```sql
-- migrations/000010_crystal_schema.up.sql
-- Crystal v1 schema: dealer scoring system for BPCL Central Delhi Sales Area

CREATE TABLE dealers (
  cc_code        VARCHAR(10) PRIMARY KEY,
  ro_name        TEXT NOT NULL,
  area           TEXT NOT NULL DEFAULT 'Central Delhi',
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  dealer_email   TEXT,
  cc_email       TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_dealers_active ON dealers (is_active) WHERE is_active = TRUE;

-- -------------------------------------------------------

CREATE TABLE competition_periods (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name           TEXT NOT NULL,
  month_year     DATE NOT NULL,
  total_slots    INT NOT NULL DEFAULT 40,
  is_active      BOOLEAN NOT NULL DEFAULT FALSE,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (month_year)
);

-- -------------------------------------------------------

CREATE TABLE monthly_targets (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code          VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
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
  FOREIGN KEY (month_year) REFERENCES competition_periods(month_year)
);

CREATE INDEX idx_targets_month ON monthly_targets (month_year);

-- -------------------------------------------------------

CREATE TABLE daily_ufill (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  txn_date DATE NOT NULL,
  count    INT NOT NULL CHECK (count >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_ufill_date ON daily_ufill (txn_date);
CREATE INDEX idx_ufill_cc_month ON daily_ufill (cc_code, DATE_TRUNC('month', txn_date));

-- -------------------------------------------------------

CREATE TABLE daily_qoc (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  txn_date DATE NOT NULL,
  count    INT NOT NULL CHECK (count >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_qoc_date ON daily_qoc (txn_date);

-- -------------------------------------------------------

CREATE TABLE daily_ms (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  txn_date DATE NOT NULL,
  kl       NUMERIC(10,3) NOT NULL CHECK (kl >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_ms_date ON daily_ms (txn_date);

-- -------------------------------------------------------

CREATE TABLE daily_speed (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  txn_date DATE NOT NULL,
  kl       NUMERIC(10,3) NOT NULL CHECK (kl >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_speed_date ON daily_speed (txn_date);

-- -------------------------------------------------------

CREATE TABLE daily_hsd (
  id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code  VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  txn_date DATE NOT NULL,
  kl       NUMERIC(10,3) NOT NULL CHECK (kl >= 0),
  UNIQUE (cc_code, txn_date)
);

CREATE INDEX idx_hsd_date ON daily_hsd (txn_date);

-- -------------------------------------------------------

CREATE TABLE mak_ge_readings (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code        VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  reading_date   DATE NOT NULL,
  meter_reading  NUMERIC(12,2) NOT NULL CHECK (meter_reading >= 0),
  entered_by     TEXT,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, reading_date)
);

CREATE INDEX idx_mak_ge_cc_month ON mak_ge_readings (cc_code, DATE_TRUNC('month', reading_date));

-- -------------------------------------------------------

CREATE TABLE google_ratings (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code       VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  snapshot_date DATE NOT NULL,
  rating        NUMERIC(3,1) NOT NULL CHECK (rating BETWEEN 1.0 AND 5.0),
  review_count  INT NOT NULL CHECK (review_count >= 0),
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, snapshot_date)
);

-- -------------------------------------------------------

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

-- -------------------------------------------------------

CREATE TABLE dealer_scores (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  cc_code         VARCHAR(10) NOT NULL REFERENCES dealers(cc_code),
  month_year      DATE NOT NULL,
  metric_key      TEXT NOT NULL,
  actual_value    NUMERIC(15,4),
  rank_in_group   INT,
  marks_scored    NUMERIC(6,3),
  max_marks       NUMERIC(5,2),
  computed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (cc_code, month_year, metric_key),
  FOREIGN KEY (month_year) REFERENCES competition_periods(month_year)
);

CREATE INDEX idx_scores_month ON dealer_scores (month_year);
CREATE INDEX idx_scores_cc ON dealer_scores (cc_code, month_year);

-- -------------------------------------------------------

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

-- -------------------------------------------------------

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

- [ ] **Step 2: Run and verify it applies cleanly**

```bash
cd /Users/meynesh/Documents/bpcl-sso
bash scripts/migrate.sh up
```

Expected: `000010/u crystal_schema OK` and no errors.

If DB is not running, start it first:
```bash
docker compose up -d db
sleep 3
bash scripts/migrate.sh up
```

- [ ] **Step 3: Verify tables exist**

```bash
psql "$BPCL_DB_URL" -c "\dt dealers daily_ufill daily_qoc daily_ms daily_speed daily_hsd mak_ge_readings google_ratings scoring_params dealer_scores sangam_data competition_periods"
```

Expected: 12 rows (11 tables + 1 base competition_periods already existed — but the crystal version is the new one since they don't conflict on name... wait).

**IMPORTANT NOTE:** The existing schema already has a `competition_periods` table from migration 000009. The crystal spec defines a NEW `competition_periods` table with a different schema (`month_year` column instead of `period`, no `territory_code`). These conflict.

Resolution: Name the crystal competition table `crystal_competition_periods` in the migration, OR use a separate schema. Use prefix `cr_` for all crystal tables to avoid collision with existing tables.

**Re-write the migration with `cr_` prefix:**

```sql
-- migrations/000010_crystal_schema.up.sql
-- Crystal v1 schema: cr_ prefix to coexist with legacy schema

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
CREATE INDEX idx_cr_ufill_cc_month ON cr_daily_ufill (cc_code, DATE_TRUNC('month', txn_date));

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

CREATE INDEX idx_cr_mak_ge_cc_month ON cr_mak_ge_readings (cc_code, DATE_TRUNC('month', reading_date));

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
```

- [ ] **Step 4: Commit the up migration**

```bash
git add migrations/000010_crystal_schema.up.sql
git commit -m "feat: add Crystal v1 schema migration (000010)"
```

---

## Task 2: Create the down migration (rollback)

**Files:**
- Create: `migrations/000010_crystal_schema.down.sql`

- [ ] **Step 1: Write rollback**

```sql
-- migrations/000010_crystal_schema.down.sql
DROP VIEW IF EXISTS cr_dealer_total_scores;
DROP TABLE IF EXISTS cr_dealer_scores CASCADE;
DROP TABLE IF EXISTS cr_scoring_params CASCADE;
DROP TABLE IF EXISTS cr_sangam_data CASCADE;
DROP TABLE IF EXISTS cr_google_ratings CASCADE;
DROP TABLE IF EXISTS cr_mak_ge_readings CASCADE;
DROP TABLE IF EXISTS cr_daily_hsd CASCADE;
DROP TABLE IF EXISTS cr_daily_speed CASCADE;
DROP TABLE IF EXISTS cr_daily_ms CASCADE;
DROP TABLE IF EXISTS cr_daily_qoc CASCADE;
DROP TABLE IF EXISTS cr_daily_ufill CASCADE;
DROP TABLE IF EXISTS cr_monthly_targets CASCADE;
DROP TABLE IF EXISTS cr_competition_periods CASCADE;
DROP TABLE IF EXISTS cr_dealers CASCADE;
```

- [ ] **Step 2: Test rollback on a clean state**

```bash
cd /Users/meynesh/Documents/bpcl-sso
bash scripts/migrate.sh down 1
```

Expected: `000010/d crystal_schema OK` — no errors.

- [ ] **Step 3: Re-apply the up migration**

```bash
bash scripts/migrate.sh up
```

Expected: `000010/u crystal_schema OK`.

- [ ] **Step 4: Commit**

```bash
git add migrations/000010_crystal_schema.down.sql
git commit -m "feat: add Crystal v1 rollback migration"
```

---

## Task 3: Create the dealers seed

**Files:**
- Create: `scripts/seed_crystal_dealers.sql`

The 40 dealers come from the existing `retail_outlets` seed (same BPCL Central Delhi Sales Area cc_codes). `dealer_email` and `cc_email` are NULL until `Dealer_Email_Map` sheet is imported — that is a separate data operation.

- [ ] **Step 1: Write the dealers seed**

```sql
-- scripts/seed_crystal_dealers.sql
-- 40 active BPCL Central Delhi Sales Area dealers
-- dealer_email / cc_email to be filled from Dealer_Email_Map sheet

INSERT INTO cr_dealers (cc_code, ro_name, area, is_active) VALUES
('112847', 'M.L. SETHI SERVICE STATION',    'Central Delhi', TRUE),
('108923', 'SAI FILLING STATION',            'Central Delhi', TRUE),
('115634', 'AUTO CARE',                      'Central Delhi', TRUE),
('109251', 'GARG ROAD LINES',                'Central Delhi', TRUE),
('113782', 'SAHAS FILLING STATION',          'Central Delhi', TRUE),
('107634', 'VAIBHAV FILLING STATION',        'Central Delhi', TRUE),
('114521', 'SHANKAR FILLING STATION',        'Central Delhi', TRUE),
('116890', 'SAKSHAM MOTORS',                 'Central Delhi', TRUE),
('111234', 'SANJEEV FILLING STATION',        'Central Delhi', TRUE),
('108765', 'LINK ROAD PETROL F.STN.',        'Central Delhi', TRUE),
('112345', 'GUPTA SERVICE STATION',          'Central Delhi', TRUE),
('119876', 'NATIONAL FILLING STATION',       'Central Delhi', TRUE),
('113456', 'DELHI PETROLEUM',                'Central Delhi', TRUE),
('117234', 'SUNRISE FUEL STATION',           'Central Delhi', TRUE),
('115789', 'KRISHNA PETROLEUM',              'Central Delhi', TRUE),
('118901', 'METRO FILLING STATION',          'Central Delhi', TRUE),
('110234', 'SHARMA PETROL PUMP',             'Central Delhi', TRUE),
('116123', 'BALAJI SERVICE STATION',         'Central Delhi', TRUE),
('119012', 'CAPITAL FUELS',                  'Central Delhi', TRUE),
('113234', 'ANAND FILLING STATION',          'Central Delhi', TRUE),
('117890', 'PUNJAB PETROLEUM',               'Central Delhi', TRUE),
('114678', 'RAJDHANI FUELS',                 'Central Delhi', TRUE),
('112789', 'SHIV SHAKTI FILLING STN.',       'Central Delhi', TRUE),
('118345', 'OLYMPIC FILLING STATION',        'Central Delhi', TRUE),
('110567', 'LOTUS PETROLEUM',                'Central Delhi', TRUE),
('115901', 'HARI OM FUELS',                  'Central Delhi', TRUE),
('119456', 'STAR PETROL PUMP',               'Central Delhi', TRUE),
('113890', 'NEW INDIA FILLING STN.',         'Central Delhi', TRUE),
('116456', 'BHATIA PETROLEUM',               'Central Delhi', TRUE),
('118012', 'DIAMOND FILLING STATION',        'Central Delhi', TRUE),
('111789', 'RISHI PETROLEUM',                'Central Delhi', TRUE),
('114234', 'SUNRISE AUTO CENTRE',            'Central Delhi', TRUE),
('117567', 'JUPITER FILLING STATION',        'Central Delhi', TRUE),
('116789', 'BHAGWATI FILLING STATION',       'Central Delhi', TRUE),
('112012', 'MODERN FUELS',                   'Central Delhi', TRUE),
('118567', 'BP-GOLDEN PARK',                 'Central Delhi', TRUE),
('111456', 'FAST TRACK FILLING STATION',     'Central Delhi', TRUE),
('109678', 'KRITI NANAK FILLING STN.',       'Central Delhi', TRUE),
('259713', 'SAKSHAM MOTORS ADHOC',           'Central Delhi', TRUE),
('115012', 'MAHADEV FILLING STATION',        'Central Delhi', TRUE)
ON CONFLICT (cc_code) DO NOTHING;
```

- [ ] **Step 2: Run the seed**

```bash
psql "$BPCL_DB_URL" -f scripts/seed_crystal_dealers.sql
```

Expected output: `INSERT 0 40`

- [ ] **Step 3: Verify count**

```bash
psql "$BPCL_DB_URL" -c "SELECT COUNT(*) FROM cr_dealers WHERE is_active = TRUE;"
```

Expected: `40`

- [ ] **Step 4: Commit**

```bash
git add scripts/seed_crystal_dealers.sql
git commit -m "seed: 40 Crystal dealers for Central Delhi Sales Area"
```

---

## Task 4: Create the May 2026 competition + scoring_params + test data seed

**Files:**
- Create: `scripts/seed_crystal_may2026.sql`

- [ ] **Step 1: Write the seed file**

```sql
-- scripts/seed_crystal_may2026.sql
-- May 2026 competition period + scoring params + 6 days of test data

-- ============================================================
-- 1. COMPETITION PERIOD: Boost and Win May 2026
-- ============================================================
INSERT INTO cr_competition_periods (name, month_year, total_slots, is_active)
VALUES ('Boost and Win May 2026', '2026-05-01', 40, TRUE)
ON CONFLICT (month_year) DO NOTHING;

-- ============================================================
-- 2. SCORING PARAMS (15 rows, 10 active)
-- ============================================================
INSERT INTO cr_scoring_params
  (month_year, metric_key, display_name, max_marks, negative_scale_enabled, is_active, sort_order)
VALUES
  ('2026-05-01', 'ms_absolute_vol',   'MS Volume (KL)',             10, FALSE, TRUE,  1),
  ('2026-05-01', 'ms_growth_pct',     'MS Growth %',                 5, TRUE,  TRUE,  2),
  ('2026-05-01', 'ms_ms_gain_pct',    'MS Market Share Gain %',     10, TRUE,  FALSE, 3),
  ('2026-05-01', 'hsd_absolute_vol',  'HSD Volume (KL)',            10, FALSE, TRUE,  4),
  ('2026-05-01', 'hsd_growth_pct',    'HSD Growth %',                5, TRUE,  TRUE,  5),
  ('2026-05-01', 'hsd_ms_gain_pct',   'HSD Market Share Gain %',   10, TRUE,  FALSE, 6),
  ('2026-05-01', 'oil_change_count',  'Oil Change (QOC)',           10, FALSE, TRUE,  7),
  ('2026-05-01', 'mak_ge_sales',      'MAK / GE Sales',            10, FALSE, TRUE,  8),
  ('2026-05-01', 'ufill_txns',        'UFill Transactions',          5, FALSE, TRUE,  9),
  ('2026-05-01', 'speed_vol',         'Speed Volume (KL)',           5, FALSE, TRUE,  10),
  ('2026-05-01', 'cleanliness_audit', 'Cleanliness Audit',           5, FALSE, FALSE, 11),
  ('2026-05-01', 'sangam_certs',      'Sangam Certifications',       5, FALSE, TRUE,  12),
  ('2026-05-01', 'google_rating',     'Google Rating',               5, FALSE, TRUE,  13),
  ('2026-05-01', 'ips_pct',           'IPS %',                       5, FALSE, FALSE, 14),
  ('2026-05-01', 'bonus',             'Bonus',                      10, FALSE, TRUE,  15)
ON CONFLICT (month_year, metric_key) DO NOTHING;

-- ============================================================
-- 3. SIX DAYS OF TEST DATA (May 1–6, 2026)
--    Values are deterministic per dealer using HASHTEXT so
--    MTD sums are reproducible for verification queries.
--    All 40 dealers × 5 metric tables × 6 days = 1,200 rows each group.
-- ============================================================

-- UFILL: counts 5-25 per day
INSERT INTO cr_daily_ufill (cc_code, txn_date, count)
SELECT
  d.cc_code,
  s.txn_date,
  5 + ABS(HASHTEXT(d.cc_code || s.txn_date::TEXT)) % 21
FROM cr_dealers d
CROSS JOIN (
  SELECT generate_series(
    '2026-05-01'::DATE,
    '2026-05-06'::DATE,
    '1 day'::INTERVAL
  )::DATE AS txn_date
) s
ON CONFLICT (cc_code, txn_date) DO UPDATE SET count = EXCLUDED.count;

-- QOC: counts 0-12 per day
INSERT INTO cr_daily_qoc (cc_code, txn_date, count)
SELECT
  d.cc_code,
  s.txn_date,
  ABS(HASHTEXT(d.cc_code || 'qoc' || s.txn_date::TEXT)) % 13
FROM cr_dealers d
CROSS JOIN (
  SELECT generate_series(
    '2026-05-01'::DATE,
    '2026-05-06'::DATE,
    '1 day'::INTERVAL
  )::DATE AS txn_date
) s
ON CONFLICT (cc_code, txn_date) DO UPDATE SET count = EXCLUDED.count;

-- MS: 3.000 – 12.999 KL per day (regular petrol, not Speed)
INSERT INTO cr_daily_ms (cc_code, txn_date, kl)
SELECT
  d.cc_code,
  s.txn_date,
  ROUND((3.0 + (ABS(HASHTEXT(d.cc_code || 'ms' || s.txn_date::TEXT)) % 10000) / 1000.0)::NUMERIC, 3)
FROM cr_dealers d
CROSS JOIN (
  SELECT generate_series(
    '2026-05-01'::DATE,
    '2026-05-06'::DATE,
    '1 day'::INTERVAL
  )::DATE AS txn_date
) s
ON CONFLICT (cc_code, txn_date) DO UPDATE SET kl = EXCLUDED.kl;

-- SPEED: 0.000 – 3.999 KL per day (Speed + Speed100 already summed)
INSERT INTO cr_daily_speed (cc_code, txn_date, kl)
SELECT
  d.cc_code,
  s.txn_date,
  ROUND((ABS(HASHTEXT(d.cc_code || 'spd' || s.txn_date::TEXT)) % 4000) / 1000.0::NUMERIC, 3)
FROM cr_dealers d
CROSS JOIN (
  SELECT generate_series(
    '2026-05-01'::DATE,
    '2026-05-06'::DATE,
    '1 day'::INTERVAL
  )::DATE AS txn_date
) s
ON CONFLICT (cc_code, txn_date) DO UPDATE SET kl = EXCLUDED.kl;

-- HSD: 8.000 – 27.999 KL per day
INSERT INTO cr_daily_hsd (cc_code, txn_date, kl)
SELECT
  d.cc_code,
  s.txn_date,
  ROUND((8.0 + (ABS(HASHTEXT(d.cc_code || 'hsd' || s.txn_date::TEXT)) % 20000) / 1000.0)::NUMERIC, 3)
FROM cr_dealers d
CROSS JOIN (
  SELECT generate_series(
    '2026-05-01'::DATE,
    '2026-05-06'::DATE,
    '1 day'::INTERVAL
  )::DATE AS txn_date
) s
ON CONFLICT (cc_code, txn_date) DO UPDATE SET kl = EXCLUDED.kl;
```

- [ ] **Step 2: Run the seed**

```bash
psql "$BPCL_DB_URL" -f scripts/seed_crystal_may2026.sql
```

Expected output (in order):
```
INSERT 0 1        -- competition_period
INSERT 0 15       -- scoring_params
INSERT 0 240      -- ufill (40 dealers × 6 days)
INSERT 0 240      -- qoc
INSERT 0 240      -- ms
INSERT 0 240      -- speed
INSERT 0 240      -- hsd
```

- [ ] **Step 3: Verify scoring_params count and active count**

```bash
psql "$BPCL_DB_URL" -c "
SELECT
  COUNT(*) AS total_params,
  COUNT(*) FILTER (WHERE is_active) AS active_params
FROM cr_scoring_params
WHERE month_year = '2026-05-01';"
```

Expected: `total_params=15, active_params=10`

- [ ] **Step 4: Commit**

```bash
git add scripts/seed_crystal_may2026.sql
git commit -m "seed: May 2026 competition period, scoring params, and 6-day test data"
```

---

## Task 5: Create the verification script

**Files:**
- Create: `scripts/verify_crystal_chunk1.sql`

- [ ] **Step 1: Write the verification queries**

```sql
-- scripts/verify_crystal_chunk1.sql
-- Verifies all 7 acceptance criteria for Crystal Chunk 1

\echo '=== AC1: Migration ran clean (tables exist) ==='
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_name IN (
    'cr_dealers','cr_competition_periods','cr_monthly_targets',
    'cr_daily_ufill','cr_daily_qoc','cr_daily_ms','cr_daily_speed',
    'cr_daily_hsd','cr_mak_ge_readings','cr_google_ratings',
    'cr_scoring_params','cr_dealer_scores','cr_sangam_data'
  )
ORDER BY table_name;
-- Expected: 13 rows

\echo '=== AC2: View exists ==='
SELECT viewname FROM pg_views WHERE viewname = 'cr_dealer_total_scores';
-- Expected: 1 row

\echo '=== AC3: 40 dealers seeded ==='
SELECT COUNT(*) AS dealer_count FROM cr_dealers WHERE is_active = TRUE;
-- Expected: 40

\echo '=== AC4: May 2026 scoring_params — 15 rows, 10 active ==='
SELECT
  COUNT(*)                              AS total,
  COUNT(*) FILTER (WHERE is_active)    AS active
FROM cr_scoring_params WHERE month_year = '2026-05-01';
-- Expected: total=15, active=10

\echo '=== AC5: 6 days of daily data inserted for all 40 dealers ==='
SELECT
  'ufill'  AS tbl, COUNT(*) AS rows FROM cr_daily_ufill  WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06'
UNION ALL SELECT
  'qoc',          COUNT(*)          FROM cr_daily_qoc    WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06'
UNION ALL SELECT
  'ms',           COUNT(*)          FROM cr_daily_ms     WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06'
UNION ALL SELECT
  'speed',        COUNT(*)          FROM cr_daily_speed  WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06'
UNION ALL SELECT
  'hsd',          COUNT(*)          FROM cr_daily_hsd    WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06';
-- Expected: all rows = 240 (40 dealers × 6 days)

\echo '=== AC6: MTD sums sample — top 5 dealers by MS volume ==='
SELECT
  cc_code,
  SUM(kl)                              AS mtd_ms_kl,
  SUM(kl) + COALESCE((
    SELECT SUM(s.kl) FROM cr_daily_speed s
    WHERE s.cc_code = m.cc_code
      AND s.txn_date BETWEEN '2026-05-01' AND '2026-05-06'
  ), 0)                                AS mtd_ms_plus_speed_kl
FROM cr_daily_ms m
WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06'
GROUP BY cc_code
ORDER BY mtd_ms_kl DESC
LIMIT 5;
-- Compare with Excel MTD column — should be within ±0.01 KL
-- (Cross-check manually when Excel data is available)

\echo '=== AC7: dealer_total_scores view returns ranks ==='
-- Populate sample scores to test the view
INSERT INTO cr_dealer_scores (cc_code, month_year, metric_key, actual_value, rank_in_group, marks_scored, max_marks)
SELECT
  d.cc_code,
  '2026-05-01',
  'ms_absolute_vol',
  SUM(m.kl),
  RANK() OVER (ORDER BY SUM(m.kl) DESC),
  ROUND((RANK() OVER (ORDER BY SUM(m.kl) DESC)::NUMERIC / 40) * 10, 3),
  10
FROM cr_dealers d
JOIN cr_daily_ms m ON m.cc_code = d.cc_code
WHERE m.txn_date BETWEEN '2026-05-01' AND '2026-05-06'
GROUP BY d.cc_code
ON CONFLICT (cc_code, month_year, metric_key) DO NOTHING;

SELECT cc_code, ro_name, total_marks, rank_overall
FROM cr_dealer_total_scores
WHERE month_year = '2026-05-01'
ORDER BY rank_overall
LIMIT 10;
-- Expected: 40 rows total, ranks 1-40, no ties at rank 1
```

- [ ] **Step 2: Run verification**

```bash
psql "$BPCL_DB_URL" -f scripts/verify_crystal_chunk1.sql
```

Review output against expected values noted in each section.

- [ ] **Step 3: Commit**

```bash
git add scripts/verify_crystal_chunk1.sql
git commit -m "test: Crystal Chunk 1 verification queries"
```

---

## Task 6: Run migration end-to-end and confirm all AC pass

- [ ] **Step 1: Full rollback + re-apply cycle**

```bash
cd /Users/meynesh/Documents/bpcl-sso
bash scripts/migrate.sh down 1
bash scripts/migrate.sh up
```

Both commands must complete without error.

- [ ] **Step 2: Re-seed and re-verify**

```bash
psql "$BPCL_DB_URL" -f scripts/seed_crystal_dealers.sql
psql "$BPCL_DB_URL" -f scripts/seed_crystal_may2026.sql
psql "$BPCL_DB_URL" -f scripts/verify_crystal_chunk1.sql
```

All verification checks must pass.

- [ ] **Step 3: Final commit**

```bash
git add -A
git status
git commit -m "chore: Crystal Chunk 1 schema — all AC verified"
```

---

## Self-Review: Spec Coverage Check

| Spec requirement | Task/Step |
|---|---|
| `dealers` table with all columns + partial index | Task 1 |
| `competition_periods` with `UNIQUE(month_year)` | Task 1 |
| `monthly_targets` with FK to both dealers + competition_periods | Task 1 |
| `daily_ufill`, `daily_qoc`, `daily_ms`, `daily_speed`, `daily_hsd` | Task 1 |
| `mak_ge_readings` with meter reading delta model | Task 1 |
| `google_ratings` with rating CHECK constraint | Task 1 |
| `scoring_params` configurable per period | Task 1 |
| `dealer_scores` normalized, UNIQUE(cc,month,metric) | Task 1 |
| `dealer_total_scores` view with RANK() OVER | Task 1 |
| `sangam_data` table | Task 1 |
| Rollback script | Task 2 |
| 40 dealers seeded | Task 3 |
| May 2026 competition_period seeded | Task 4 |
| 15 scoring_params (10 active) | Task 4 |
| 6 days of daily rows for all 40 dealers | Task 4 |
| MTD sum verification | Task 5 AC6 |
| dealer_total_scores view ranks | Task 5 AC7 |

**Gaps identified:**
- `Dealer_Email_Map` emails: `dealer_email`/`cc_email` columns are NULL — to be backfilled from Excel in a separate data operation (this is not a schema gap, just a data gap acknowledged in the plan)
- Monthly targets: seed does not include `cr_monthly_targets` rows — not required by AC (AC only tests daily data MTD sums and scoring_params)

No placeholder steps. All code is complete.
