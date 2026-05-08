-- May 2026 competition period, scoring params, and 6-day test data

-- ============================================================
-- 1. COMPETITION PERIOD: Boost and Win May 2026
-- ============================================================
INSERT INTO cr_competition_periods (name, month_year, total_slots, is_active)
VALUES ('Boost and Win May 2026', '2026-05-01', 40, TRUE)
ON CONFLICT (month_year) DO NOTHING;

-- ============================================================
-- 2. SCORING PARAMS (15 rows, 10 active)
-- Inactive: ms_ms_gain_pct(3), hsd_ms_gain_pct(6), cleanliness_audit(11), ips_pct(14)
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
--    Deterministic values via HASHTEXT for reproducible MTD sums.
--    Speed + Speed100 already summed into cr_daily_speed per invariant #7.
-- ============================================================

-- UFILL: 5–25 transactions per dealer per day
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

-- QOC: 0–12 oil changes per dealer per day
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

-- MS (regular petrol, NOT Speed): 3.000–12.999 KL per dealer per day
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

-- SPEED (Speed + Speed100 combined): 0.000–3.999 KL per dealer per day
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

-- HSD (Diesel): 8.000–27.999 KL per dealer per day
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
