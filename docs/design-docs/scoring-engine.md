# Scoring Engine Design — Boost and Win Competition
## 100-Point Dealer Scoring System

---

## Overview

Monthly competition scoring 39 BPCL Delhi dealers across fuel volume, market share,
non-fuel revenue, compliance, and digital adoption. March 2026 is the first live edition.

Competition name: **Boost and Win**
Period: Monthly (March 2026 onward)
Dealers: ~39 active BPCL outlets in Delhi territory
Score range: Can go negative (observed: -5.65 to +56.37 in March 2026)

---

## The 15 Parameters (100 marks total)

| # | Parameter | Column | Max Marks | Negative Allowed | Source |
|---|---|---|---|---|---|
| 1 | MS Volume (KL) | ms_vol_kl | 10 | No | performance_records |
| 2 | MS Growth % | ms_growth_pct | 5 | Yes | performance_records vs last_year |
| 3 | MS Market Share Gain % | ms_ta_ms_gain_pct | 10 | Yes | market_share_data (Delhi Master) |
| 4 | HSD Volume (KL) | hsd_vol_kl | 10 | No | performance_records |
| 5 | HSD Growth % | hsd_growth_pct | 5 | Yes | performance_records vs last_year |
| 6 | HSD Market Share Gain % | hsd_ta_ms_gain_pct | 10 | Yes | market_share_data (Delhi Master) |
| 7 | Oil Change Count | qoc_count | 10 | No | performance_records (QOC product) |
| 8 | MAK/GE Sales (₹) | lubricants_value | 10 | No | performance_records (Lubricants product) |
| 9 | UFill Transactions | ufill_count | 5 | No | performance_records (UFill product) |
| 10 | Speed Volume (KL) | speed_vol_kl | 5 | No | performance_records (SPEED product) |
| 11 | Cleanliness Audit | cleanliness_grade | 5 | No | dealer_audit_scores |
| 12 | Sangam Certifications | sangam_count | 5 | No | dealer_audit_scores |
| 13 | Google Rating | google_rating | 5 | No | dealer_audit_scores |
| 14 | IPS % | ips_pct | 5 | No | dealer_audit_scores |
| 15 | Bonus | bonus_marks | 10 | No | competition_bonus (requires remarks) |

---

## Scoring Logic Per Parameter

### Volume parameters (MS Vol, HSD Vol, Speed Vol, QOC, UFill, Lubricants, SBI)
```
score = (dealer_value / max_value_in_group) * max_marks
```
Relative to the highest performer in the group. Zero if dealer_value is NULL (not blank=zero).

### Growth % parameters (MS Growth, HSD Growth)
```
growth_pct = (current_period - last_year_period) / last_year_period * 100
score = CASE
  WHEN growth_pct >= 15  THEN max_marks      -- full score
  WHEN growth_pct >= 0   THEN (growth_pct / 15) * max_marks
  WHEN growth_pct >= -15 THEN (growth_pct / 15) * max_marks  -- negative score
  ELSE -max_marks                              -- floor at -max_marks
END
```

### Market Share Gain % (MS TA Gain, HSD TA Gain)
```
dealer_share_current = dealer_ms_vol / trading_area_total_ms_vol (current period)
dealer_share_last    = dealer_ms_vol_ly / trading_area_total_ms_vol_ly (last year)
gain_pct = dealer_share_current - dealer_share_last  (in percentage points)

score = CASE
  WHEN gain_pct >= 2   THEN max_marks
  WHEN gain_pct >= 0   THEN (gain_pct / 2) * max_marks
  WHEN gain_pct >= -2  THEN (gain_pct / 2) * max_marks  -- negative
  ELSE -max_marks
END
```

### Cleanliness Audit
```
Excellent     = 5 marks
Good          = 4 marks
Average       = 3 marks
Below Average = 1 mark
Poor          = 0 marks
NOT_AUDITED   = excluded from scoring (see Bug #2)
```

### Google Rating
```
score = ((rating - 1) / 4) * 5   -- maps 1.0–5.0 stars to 0–5 marks
```

### IPS %
```
score = (ips_pct / 100) * 5
```

---

## ⚠️ BUG #1 — MS/HSD Market Share Gain Missing for 35/39 Dealers (CRITICAL)

**Status:** Unresolved in Excel template. Must be fixed in DB implementation.
**Impact:** 20/100 marks unscored for 90% of dealers. Rankings are invalid without this.

**Root cause:** The Delhi Master trading area totals were never computed and joined into
the scoring template. The `ms_ta_ms_gain_pct` column is blank for 35 of 39 dealers.

**Fix:** The `market_share_data` table (populated from Delhi_Master.xlsx) contains all
BPCL + HPCL + IOCL monthly volumes by trading area. The scoring engine must JOIN to this
table to compute `trading_area_total` and derive gain %.

**Required before any competition result is considered final.**

---

## ⚠️ BUG #2 — Blank Audit Score ≠ Zero Score (DATA INTEGRITY)

**Status:** Must be enforced at DB and service layer.
**Impact:** Dealers who were never audited score 0 for Cleanliness, IPS, Google Rating —
same as a dealer who was audited and scored Poor. This is unfair and incorrect.

**Fix:** The `dealer_audit_scores` table uses a `NOT_AUDITED` sentinel value (NULL for
the score, with a separate `audit_status` column = 'not_audited' | 'audited').

**Scoring rule:**
- `audit_status = 'audited'` AND `score = 0` → 0 marks (genuine poor performance)
- `audit_status = 'not_audited'` → EXCLUDED from parameter denominator
- Never treat NULL as 0 in any scoring function

---

## ⚠️ BUG #3 — Relative Volume Scoring Allows Mediocre Absolute Scores (DESIGN)

**Status:** Known design limitation. Not a bug per se, but document for future.
**Impact:** If all 39 dealers have low volumes, the best of a weak group still scores 10/10.

**Current behavior:** Intentional for intra-group competition. Document in competition rules.
**Future improvement:** Add absolute benchmark thresholds in Phase 2.

---

## ⚠️ BUG #4 — Bonus Marks Have No Audit Trail (GOVERNANCE)

**Status:** Must be enforced at DB level.
**Impact:** 10-point discretionary allocation currently has no justification recorded.
Only SANJEEV FILLING STATION received bonus in March 2026 with no documented reason.

**Fix:** The `competition_bonus` table has a mandatory `remarks` column (NOT NULL).
The API endpoint for setting bonus marks MUST reject requests with blank/null remarks.
Service layer: `if len(remarks) < 10 { return ErrBonusRemarksRequired }`

---

## ⚠️ BUG #5 — ADHOC Dealers in Main Ranking (DATA CLASSIFICATION)

**Status:** Must be handled at DB level.
**Impact:** SAKSHAM MOTORS ADHOC (CC: 259713) is an adhoc/transitional outlet included
in the main 39-dealer ranking, artificially inflating the competitive field.

**Fix:** `retail_outlets` table has `outlet_type` column:
- `'regular'` — participates in main Boost and Win ranking
- `'adhoc'` — excluded from main ranking, shown in separate view
- `'coco'` — company-owned, separate reporting

**Scoring engine MUST filter:** `WHERE outlet_type = 'regular'` for main leaderboard.

---

## March 2026 Results Reference (Top 10)

| Rank | Dealer | CC | Score |
|---|---|---|---|
| 1 | M.L. SETHI SERVICE STATION | 112847 | 56.37 |
| 2 | SAI FILLING STATION | 108923 | 55.35 |
| 3 | AUTO CARE | 115634 | 48.17 |
| 4 | GARG ROAD LINES | 109251 | 45.85 |
| 5 | SAHAS FILLING STATION | 113782 | 45.73 |
| 6 | VAIBHAV FILLING STATION | 107634 | 43.68 |
| 7 | SHANKAR FILLING STATION | 114521 | 41.50 |
| 8 | SAKSHAM MOTORS | 116890 | 37.02 |
| 9 | SANJEEV FILLING STATION | 111234 | 35.60 |
| 10 | LINK ROAD PETROL F.STN. | 108765 | 34.20 |

Bottom: MAHADEV FILLING STATION (CC: 115012) scored -5.65 — extreme decline, zero across all non-fuel.
