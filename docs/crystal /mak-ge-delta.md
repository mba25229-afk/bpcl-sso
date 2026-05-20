# MAK GE Delta Rule — Locked
**Status:** LOCKED

---

## Rule

**MAK GE total_sales = last meter reading of month − first meter reading of month**

MAK GE readings are cumulative meter values recorded weekly (approximately 1st, 8th, 15th, 22nd, 31st of month). The monthly sales volume is the delta between the last and first reading.

```
total_sales_kl = reading_at_month_end - reading_at_month_start
```

If `total_sales_kl < 0`: data entry error — flag for admin review, do not score.  
If only one reading exists for the month: `total_sales_kl = NULL`, exclude from scoring.  
If no readings exist: `total_sales_kl = 0`, dealer scores 0 for this metric.

## DB Table

```sql
mak_ge_readings (
  id, cc_code, reading_date DATE, meter_reading DECIMAL(10,2),
  total_sales DECIMAL(10,2) GENERATED -- computed on fetch, not stored
)
```

## Computation

```sql
-- Monthly MAK GE sales per dealer
WITH ranked AS (
  SELECT
    cc_code,
    meter_reading,
    ROW_NUMBER() OVER (PARTITION BY cc_code ORDER BY reading_date ASC)  AS rn_first,
    ROW_NUMBER() OVER (PARTITION BY cc_code ORDER BY reading_date DESC) AS rn_last,
    COUNT(*) OVER (PARTITION BY cc_code) AS reading_count
  FROM mak_ge_readings
  WHERE DATE_TRUNC('month', reading_date) = :target_month
)
SELECT
  cc_code,
  MAX(CASE WHEN rn_last = 1 THEN meter_reading END) -
  MAX(CASE WHEN rn_first = 1 THEN meter_reading END) AS total_sales_kl
FROM ranked
WHERE reading_count >= 2
GROUP BY cc_code
```

## Weekly Schedule

Readings are entered on a weekly basis — approximately:
- Week 1: Reading as of 1st of month
- Week 2: Reading as of 8th
- Week 3: Reading as of 15th
- Week 4: Reading as of 22nd
- Week 5: Reading as of 31st (or last day of month)

Admin enters these via admin portal → MAK GE Reading entry form.
