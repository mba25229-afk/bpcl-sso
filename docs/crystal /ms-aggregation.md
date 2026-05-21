# MS Aggregation Rule — Locked
**Status:** LOCKED

---

## Rule

**MS (Motor Spirit / Petrol) total = Regular Petrol + Speed + Speed 100**

Speed and Speed 100 are product variants of petrol. They are NOT separate metrics from MS — they are sub-products that roll up into total MS volume.

```
daily_ms_total(cc_code, date) = 
    daily_ms.kl               -- Regular petrol (sourced from MS sheet)
    + daily_speed.kl          -- Speed + Speed100 combined (sourced from SPEED sheet)
```

## Source Data Mapping

| Excel Sheet | Column | DB Table | Notes |
|---|---|---|---|
| MS | daily KL | `daily_ms` | Regular petrol only |
| SPEED | daily KL | `daily_speed` | Sum of Speed + Speed100 per dealer per day |

The SPEED sheet in MAY_DATA.xlsx has one row per dealer (Speed and Speed100 already summed at source — confirmed no separate Speed100 row after consolidation instruction).

## What Goes Where in Scoring

| Metric | Uses |
|---|---|
| `ms_absolute_vol` (max 10 marks) | `daily_ms.kl + daily_speed.kl` — full petrol volume |
| `speed_vol` (max 5 marks) | `daily_speed.kl` only — Speed product premium metric |

Speed volume is scored **twice**: once as part of total MS, once as its own metric. This is intentional — it incentivises stocking and pushing the premium Speed product.

## DB Computation

```sql
-- MTD MS total for scoring
SELECT 
  cc_code,
  SUM(kl) AS ms_total_kl
FROM daily_ms
WHERE month_year = :month_year
GROUP BY cc_code

UNION ALL

SELECT
  cc_code,
  SUM(kl) AS speed_kl  
FROM daily_speed
WHERE month_year = :month_year
GROUP BY cc_code
```

Aggregated server-side before ranking. Never store the combined total — always compute from source tables.

## LY (Last Year) for Growth Computation

MS Growth % = `(current_month_ms_total - ly_ms_total) / ly_ms_total × 100`

LY values are stored in `monthly_targets.ms_ly` and `monthly_targets.hsd_ly` (from TARGETS sheet). These are manually entered targets from the area manager — not computed from historical daily data.
