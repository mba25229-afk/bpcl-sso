# Scoring Formula — Locked
**Source:** Dealer_Competition_Scoring_Template_v4.xlsx  
**Verified:** 40/40 dealers match on all volume metrics  
**Status:** LOCKED — do not change without updating this doc + Params table in DB

---

## Core Formula

```
score_for_metric = ((n - rank) / (n - 1)) × max_marks
```

| Variable | Definition |
|---|---|
| `n` | Total competition slots — **fixed at 40** (includes empty/blank slots) |
| `rank` | Dealer's rank for this metric, descending (rank 1 = highest value = best) |
| `max_marks` | Maximum marks for this metric — stored in `scoring_params` table |

**This is peer-relative ranking, not target-vs-actual.** A dealer's score depends entirely on where they sit among the 40 enrolled dealers. Rank 1 always earns `max_marks`. Rank 40 always earns 0. A dealer improving their absolute number can still lose marks if peers improve more.

---

## Metric Weights (Max Marks)

| metric_key | display_name | max_marks | data_source | notes |
|---|---|---|---|---|
| `ms_absolute_vol` | MS Volume (KL) | 10 | daily_ms MTD sum | MS = Regular Petrol + Speed + Speed100 combined |
| `ms_growth_pct` | MS Growth % | 5 | computed vs LY | Can go negative — see Negative Scale below |
| `ms_ms_gain_pct` | MS Market Share Gain % | 10 | territory-level MS% vs LY | Data not yet active — returns NULL until enabled |
| `hsd_absolute_vol` | HSD Volume (KL) | 10 | daily_hsd MTD sum | |
| `hsd_growth_pct` | HSD Growth % | 5 | computed vs LY | Can go negative |
| `hsd_ms_gain_pct` | HSD Market Share Gain % | 10 | territory-level HSD MS% vs LY | Not yet active |
| `oil_change_count` | Oil Change (QOC) | 10 | daily_qoc MTD sum | |
| `mak_ge_sales` | MAK / GE Sales | 10 | mak_ge_readings delta | total_sales = last_reading − first_reading for month |
| `ufill_txns` | UFill Transactions | 5 | daily_ufill MTD sum | |
| `speed_vol` | Speed Volume (KL) | 5 | daily_speed MTD sum | Speed product only — separate from MS total |
| `cleanliness_audit` | Cleanliness Audit | 5 | cleanliness_grade — fixed lookup | Grade-based, not rank-based (see below) |
| `sangam_certs` | Sangam Certificates | 5 | sangam_cert_count | |
| `google_rating` | Google Rating | 5 | google_ratings snapshot | Rank by rating × review_count product |
| `ips_pct` | IPS % | 5 | ips_pct field | Not yet active — NULL until ALP data enabled |
| `bonus` | Bonus | 10 | manual_bonus_marks | Discretionary — admin entry only |

**Total possible: 115 marks**

---

## Negative Scale Behaviour (Growth Metrics)

Toggle: **"Use Negative Scale for Growth? = Yes"** (applies to `ms_growth_pct` and `hsd_growth_pct`)  
Toggle: **"Use Negative Scale for MS Gain? = Yes"** (applies to `ms_ms_gain_pct` and `hsd_ms_gain_pct`)

When negative scale is ON, the rank-based formula still applies but the score range extends below 0:

```
score = ((n - rank) / (n - 1) × max_marks × 2) − max_marks
```

| rank | n=40, max=5 | result |
|---|---|---|
| 1 (best growth) | 39/39 × 10 − 5 | **+5.0** |
| 20 (median) | 20/39 × 10 − 5 | **+0.13** |
| 40 (worst) | 0/39 × 10 − 5 | **−5.0** |

A dealer with the worst growth in the group can subtract up to `max_marks` from their total.

---

## Cleanliness — Grade Lookup (not rank-based)

| grade | marks |
|---|---|
| Excellent | 5 |
| Good | 4 |
| Average | 3 |
| Below Average | 1 |
| Poor | 0 |
| NULL / not audited | 0 |

---

## Google Rating Scoring

Input data from `google_ratings` table: `rating` (1–5 stars) and `review_count`.  
Scoring rank is determined by: `rating × review_count` (composite engagement score).  
Apply standard rank formula with `max_marks = 5`.

---

## MAK GE Scoring

Input: `mak_ge_readings` — weekly meter readings per dealer.  
`total_sales_kl = last_reading_of_month − first_reading_of_month`  
If only one reading exists: `total_sales_kl = 0` (insufficient data).  
Apply standard rank formula with `max_marks = 10`.

---

## Competition Period

The `n = 40` slot count is fixed per competition period, not dynamic.  
When a new competition period is created, admin sets `total_slots = 40`.  
Empty slots (non-enrolled or inactive dealers) always score 0 and rank last.  
Rank ties: `method = 'first'` (first occurrence in dataset retains better rank).

---

## Metrics Currently NOT Scored (NULL output)

- `ms_ms_gain_pct` — requires territory-level market share data, not yet fed
- `hsd_ms_gain_pct` — same
- `ips_pct` — ALP module not in scope (Chunk 4+)
- `cleanliness_audit` — requires field audit input, no automated source

These return `NULL` in `dealer_scores.marks_scored` and are excluded from `total_marks` sum.
