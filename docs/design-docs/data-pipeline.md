# Data Pipeline Design — Delhi Master → Market Share
## How raw OMC volumes become dealer competition scores

---

## Source: Delhi_Master.xlsx

Contains monthly MS and HSD volumes for ALL Delhi petrol stations across 3 OMCs:
- BPC (Bharat Petroleum) — the dealers we score
- HPC (Hindustan Petroleum) — competitor
- IOC (Indian Oil) — competitor

Coverage: ~200+ outlets, April 2024 – February 2026 (22 months)
Grouping: Each outlet assigned to a **trading area** (micro-market cluster)
Geography: 9 Delhi districts

---

## Trading Areas (27 clusters across Delhi)

| District | Trading Areas |
|---|---|
| West Delhi | Kirti Nagar, Rajouri Garden, Tilak Nagar, Janakpuri |
| South Delhi | Saket, Lajpat Nagar, Hauz Khas, Greater Kailash |
| North West Delhi | Rohini, Pitampura, Shalimar Bagh |
| East Delhi | Preet Vihar, Mayur Vihar, Laxmi Nagar |
| South West Delhi | Dwarka, Uttam Nagar, Palam |
| Central Delhi | Connaught Place, Karol Bagh, Paharganj |
| North East Delhi | Shahdara, Dilshad Garden |
| New Delhi | Chanakyapuri, RK Puram |
| Shahdara | Shahdara North, Shahdara South |

---

## Market Share Gain Computation

### Step 1 — Monthly trading area totals (all OMCs)
```sql
-- For each trading area + period:
ta_total_ms  = SUM(ms_vol_kl) WHERE trading_area_id = X AND period = Y
ta_total_hsd = SUM(hsd_vol_kl) WHERE trading_area_id = X AND period = Y
```

### Step 2 — Dealer market share for current period
```sql
dealer_share_ms_current = dealer_ms_vol_kl / ta_total_ms  (for period)
dealer_share_ms_ly      = dealer_ms_vol_kl_ly / ta_total_ms_ly  (for same period last year)
```

### Step 3 — Market share gain (percentage points, not relative %)
```
ms_gain_pp  = dealer_share_ms_current - dealer_share_ms_ly
hsd_gain_pp = dealer_share_hsd_current - dealer_share_hsd_ly
```

### Step 4 — Score mapping (see scoring-engine.md parameter #3 and #6)
```
+2pp or more  → full 10 marks
0 to +2pp     → proportional positive
-2pp to 0pp   → proportional negative
Below -2pp    → -10 marks (floor)
```

---

## Database Tables That Support This

### `trading_areas`
```
id, name, district, state
```

### `market_share_data`
```
outlet_id, trading_area_id, omc (bpcl|hpcl|iocl), period,
ms_vol_kl, hsd_vol_kl, source (delhi_master|manual)
```

### `trading_area_totals` (computed/materialized)
```
trading_area_id, period, total_ms_vol_kl, total_hsd_vol_kl,
bpcl_ms_vol, hpcl_ms_vol, iocl_ms_vol,
bpcl_hsd_vol, hpcl_hsd_vol, iocl_hsd_vol
```

---

## Excel Upload Format (Delhi Master)

The `parser/excel.go` needs a SECOND parse mode for Delhi Master format.
This is DIFFERENT from the regular performance upload format.

Delhi Master row structure:
```
Column A:  Outlet Name
Column B:  CC Number (BPC only, blank for HPCL/IOCL)
Column C:  OMC (BPC/HPC/IOC)
Column D:  District
Column E:  Trading Area
Columns F+: Monthly volumes (one column per month, header = "Apr-24", "May-24" etc.)
```

The parser must detect which format is being uploaded based on column headers in row 1.
- If col A header = "Outlet Name" → Delhi Master mode
- If col A header = "product_code" → Performance upload mode

---

## Refresh Cadence

Delhi Master is updated monthly (typically by 15th of following month).
When a new Delhi Master is uploaded:
1. Ingest into `market_share_data` table
2. Recompute `trading_area_totals` for all affected periods
3. Recompute competition scores for the relevant competition period
4. Update `competition_scores` and rerank dealers
