-- Crystal Chunk 1 verification queries — all 7 acceptance criteria

\echo ''
\echo '=== AC1: All 13 cr_ tables exist ==='
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

\echo ''
\echo '=== AC2: cr_dealer_total_scores view exists ==='
SELECT viewname FROM pg_views WHERE viewname = 'cr_dealer_total_scores';
-- Expected: 1 row

\echo ''
\echo '=== AC3: 40 active dealers seeded ==='
SELECT COUNT(*) AS dealer_count FROM cr_dealers WHERE is_active = TRUE;
-- Expected: 40

\echo ''
\echo '=== AC4: May 2026 scoring_params — 15 total, 10 active ==='
SELECT
  COUNT(*)                              AS total_params,
  COUNT(*) FILTER (WHERE is_active)    AS active_params
FROM cr_scoring_params
WHERE month_year = '2026-05-01';
-- Expected: total_params=15, active_params=10

\echo ''
\echo '=== AC5: 6-day daily data row counts (should all be 240) ==='
SELECT 'cr_daily_ufill'  AS tbl, COUNT(*) AS rows
  FROM cr_daily_ufill  WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06'
UNION ALL
SELECT 'cr_daily_qoc',   COUNT(*) FROM cr_daily_qoc   WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06'
UNION ALL
SELECT 'cr_daily_ms',    COUNT(*) FROM cr_daily_ms    WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06'
UNION ALL
SELECT 'cr_daily_speed', COUNT(*) FROM cr_daily_speed WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06'
UNION ALL
SELECT 'cr_daily_hsd',   COUNT(*) FROM cr_daily_hsd   WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06';
-- Expected: all 5 rows = 240

\echo ''
\echo '=== AC6: MTD MS + Speed sums — top 10 dealers (cross-check with Excel) ==='
SELECT
  m.cc_code,
  d.ro_name,
  ROUND(SUM(m.kl), 3)                   AS mtd_ms_kl,
  ROUND(SUM(sp.kl), 3)                  AS mtd_speed_kl,
  ROUND(SUM(m.kl) + SUM(sp.kl), 3)     AS mtd_ms_plus_speed_kl,
  ROUND(SUM(h.kl), 3)                   AS mtd_hsd_kl
FROM cr_daily_ms m
JOIN cr_dealers d    ON d.cc_code = m.cc_code
JOIN cr_daily_speed sp ON sp.cc_code = m.cc_code AND sp.txn_date = m.txn_date
JOIN cr_daily_hsd h    ON h.cc_code  = m.cc_code AND h.txn_date  = m.txn_date
WHERE m.txn_date BETWEEN '2026-05-01' AND '2026-05-06'
GROUP BY m.cc_code, d.ro_name
ORDER BY mtd_ms_plus_speed_kl DESC
LIMIT 10;

\echo ''
\echo '=== AC7: dealer_total_scores view — populate sample and rank ==='
INSERT INTO cr_dealer_scores
  (cc_code, month_year, metric_key, actual_value, rank_in_group, marks_scored, max_marks)
SELECT
  sub.cc_code,
  '2026-05-01',
  'ms_absolute_vol',
  sub.mtd_kl,
  RANK() OVER (ORDER BY sub.mtd_kl DESC),
  ROUND(
    (1.0 - (RANK() OVER (ORDER BY sub.mtd_kl DESC) - 1.0) / 40.0) * 10.0,
    3
  ),
  10
FROM (
  SELECT cc_code, SUM(kl) AS mtd_kl
  FROM cr_daily_ms
  WHERE txn_date BETWEEN '2026-05-01' AND '2026-05-06'
  GROUP BY cc_code
) sub
ON CONFLICT (cc_code, month_year, metric_key) DO NOTHING;

SELECT cc_code, ro_name, ROUND(total_marks, 3) AS total_marks, rank_overall
FROM cr_dealer_total_scores
WHERE month_year = '2026-05-01'
ORDER BY rank_overall
LIMIT 10;
-- Expected: 40 rows total with ranks 1–40
