# Crystal — Database Schema
> Generated from MAY_DATA.xlsx. Source of truth for all ETL and API contracts.
> Last verified: May 2026. Update when sheet structure changes.

## Dealer Master (41 ROs — Central Delhi Sales Area)

```sql
-- dealers (static, seeded once — only BPCL ops can add/remove)
CREATE TABLE dealers (
  cc_code       INTEGER PRIMARY KEY,   -- e.g. 112385
  ro_name       TEXT NOT NULL,
  dsw_available BOOLEAN DEFAULT FALSE,
  nitrogen      BOOLEAN DEFAULT FALSE,
  alp_enrolled  BOOLEAN DEFAULT FALSE,
  alp_type      TEXT,                  -- 'IPS' | null
  pos_machines  INTEGER DEFAULT 0,
  pos_working   INTEGER DEFAULT 0,
  created_at    TIMESTAMPTZ DEFAULT NOW()
);
```

### All 41 dealers (seed data)
| cc_code | ro_name |
|---|---|
| 112385 | ANAND SUPER SERVICE STN. |
| 112386 | ANIL FILLING STATION |
| 112390 | AUTO CARE |
| 151186 | BHAGWATI FILLING STATION |
| 198868 | BHARAT FILLING STATION |
| 200922 | BP-GOLDEN PARK |
| 147286 | FAST TRACK FILLING STATION |
| 112402 | GANPATI FILLING STATION |
| 144329 | GARG ROAD LINES |
| 168006 | GREEN HEART FILLING STATION |
| 112407 | JAGDISH FILLING STATION |
| 148413 | JALVAYU FILLING STATION |
| 244458 | KAMLA FUELS & MOTORS |
| 112414 | KARTIK AUTO CENTRE |
| 167171 | KRISHNA FILLING STATION |
| 112415 | KRITI NANAK FILLING STN. |
| 112417 | LAKSHMI FILLING STATION |
| 112418 | LINK ROAD PETROL F.STN. |
| 112419 | LINK ROAD PETROL S.STN. |
| 112420 | M.L.SETHI SERVICE STATION |
| 198751 | MAHADEV FILLING STATION |
| 112422 | MANN SERVICE STATION |
| 112424 | MOBILE CENTRE |
| 112431 | NEW HARYANA SERVICE STN |
| 112438 | R.N.MOTORS |
| 112439 | R.S.BHOLA RAM & SONS |
| 112440 | RAIZADA MOTORS |
| 112442 | RAJINDER SERVICE STATION |
| 112446 | ROHINI FILLING STATION |
| 127090 | SAHAS FILLING STATION |
| 163708 | SAI FILLING STATION |
| 112448 | SAKSHAM MOTORS |
| 136152 | SANJEEV FILLING STATION |
| 112450 | SHANKAR FILLING STATION |
| 128472 | SHAURYA BHUSHAN FILLING STATION |
| 112456 | SYALL SERVICE STATION |
| 112458 | VAIBHAV FILLING STATION |
| 112461 | VEEJAY SERVICE STATION |
| 131712 | VIJAY AGRO CENTRE |
| 112465 | WELCOME AUTOS |

> **Total: 40 dealers above.** One dealer (BHARAT FILLING STATION 198868) appears in data but verify active status before seeding.

---

## Monthly Targets

```sql
-- targets (set once per month per dealer, from TARGETS sheet)
CREATE TABLE targets (
  id              SERIAL PRIMARY KEY,
  cc_code         INTEGER REFERENCES dealers(cc_code),
  month           DATE NOT NULL,          -- first day of month: '2026-05-01'
  ufill_target    INTEGER,                -- Nos
  oil_change_target INTEGER,
  speed_target    NUMERIC(10,2),          -- KL
  ms_target       NUMERIC(10,2),          -- KL
  hsd_target      NUMERIC(10,2),          -- KL
  ms_ly           NUMERIC(10,2),          -- Last Year MS
  hsd_ly          NUMERIC(10,2),          -- Last Year HSD
  mak_ge_target   NUMERIC(10,2),
  darpan_target   NUMERIC(10,2),
  coolant_target  NUMERIC(10,2),
  remarks         TEXT,
  UNIQUE(cc_code, month)
);
```

---

## Daily Actuals

```sql
-- daily_actuals (one row per dealer per day per metric)
-- Source: UFILL / QOC / SPEED / MS / HSD sheets (day columns)
CREATE TABLE daily_actuals (
  id         SERIAL PRIMARY KEY,
  cc_code    INTEGER REFERENCES dealers(cc_code),
  date       DATE NOT NULL,
  ufill      INTEGER,          -- Nos (count of UFILL transactions)
  qoc        INTEGER,          -- Quick Oil Change count
  speed_kl   NUMERIC(10,3),   -- SPEED product KL
  ms_kl      NUMERIC(10,3),   -- Motor Spirit KL
  hsd_kl     NUMERIC(10,3),   -- High Speed Diesel KL
  UNIQUE(cc_code, date)
);

-- MTD is always computed: SUM(daily_actuals) WHERE date BETWEEN month_start AND :date
-- NEVER store MTD as a column — it causes double-accounting bugs
```

---

## DSW (Dealer Sales Workers)

```sql
CREATE TABLE dsw (
  id         SERIAL PRIMARY KEY,
  cc_code    INTEGER REFERENCES dealers(cc_code),
  dsw_name   TEXT,
  contact_no TEXT,
  active     BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);
```

---

## Google Ratings

```sql
-- Snapshot per dealer at specific dates (3 snapshots visible in sheet)
CREATE TABLE google_ratings (
  id         SERIAL PRIMARY KEY,
  cc_code    INTEGER REFERENCES dealers(cc_code),
  rated_at   DATE NOT NULL,
  rating     NUMERIC(2,1),   -- e.g. 4.2
  reviews    INTEGER,
  UNIQUE(cc_code, rated_at)
);
```

---

## MAK GE Readings

```sql
CREATE TABLE mak_ge_readings (
  id          SERIAL PRIMARY KEY,
  cc_code     INTEGER REFERENCES dealers(cc_code),
  reading_date DATE NOT NULL,
  reading_value NUMERIC(10,2),
  UNIQUE(cc_code, reading_date)
);
```

---

## Zod Validation Schema (ETL layer)

```typescript
// src/etl/schema.ts
// This is the ONLY place where raw Sheet data is trusted and validated.
// ETL must fail loudly if any field violates this schema.

import { z } from 'zod';

export const DailyActualRow = z.object({
  cc_code:   z.number().int().positive(),
  date:      z.string().regex(/^\d{4}-\d{2}-\d{2}$/),
  ufill:     z.number().int().min(0).nullable(),
  qoc:       z.number().int().min(0).nullable(),
  speed_kl:  z.number().min(0).nullable(),
  ms_kl:     z.number().min(0).nullable(),
  hsd_kl:    z.number().min(0).nullable(),
});

export const TargetRow = z.object({
  cc_code:          z.number().int().positive(),
  month:            z.string().regex(/^\d{4}-\d{2}-01$/),
  ufill_target:     z.number().int().min(0).nullable(),
  oil_change_target:z.number().int().min(0).nullable(),
  speed_target:     z.number().min(0).nullable(),
  ms_target:        z.number().min(0).nullable(),
  hsd_target:       z.number().min(0).nullable(),
  ms_ly:            z.number().min(0).nullable(),
  hsd_ly:           z.number().min(0).nullable(),
  mak_ge_target:    z.number().min(0).nullable(),
  darpan_target:    z.number().min(0).nullable(),
  coolant_target:   z.number().min(0).nullable(),
  remarks:          z.string().nullable(),
});

// Excel serial dates (e.g. 46143) must be converted BEFORE validation
// Use: new Date((serial - 25569) * 86400 * 1000).toISOString().split('T')[0]
export function excelSerialToDate(serial: number): string {
  return new Date((serial - 25569) * 86400 * 1000).toISOString().split('T')[0];
}
```

---

## Critical Notes for ETL
1. Excel date columns are serial numbers (46143 = May 1 2026). Always convert via `excelSerialToDate`.
2. Empty cells = null, NOT zero. Zero and null mean different things (null = not yet entered, zero = genuinely zero transactions).
3. MTD column in sheets is redundant — compute it in SQL, never ingest it.
4. WELCOME AUTOS appears as both "Welcome Autos" and "WELCOME AUTOS" — normalize to uppercase on ingest.
5. BHARAT FILLING STATION (198868) had 0 UFILL — this is valid data, not a missing value.
