# MAY DATA ETL Pipeline Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** ETL script that reads `docs/crystal /MAY DATA.xlsx` and upserts May 2026 data into all relevant `cr_` tables.

**Architecture:** Single Python script following the existing `ingest_delhi_master_v2.py` pattern — argparse CLI, openpyxl reader, psycopg2 upserts, row-level error logging, final `cr_etl_log` entry. Runs in 9 ordered passes respecting FK constraints.

**Tech Stack:** Python 3, openpyxl, psycopg2, PostgreSQL

---

## File Map

| Action | Path |
|--------|------|
| Create | `scripts/ingest_may_data.py` |

---

### Task 1: Script Skeleton — CLI, DB, Constants

**Files:**
- Create: `scripts/ingest_may_data.py`

- [ ] **Step 1: Verify dependencies are available**

```bash
python3 -c "import openpyxl, psycopg2; print('OK')"
```

Expected: `OK`. If not: `pip3 install openpyxl psycopg2-binary`

- [ ] **Step 2: Create the script skeleton**

```python
#!/usr/bin/env python3
"""
MAY DATA Excel ingestion.

Reads docs/crystal /MAY DATA.xlsx and upserts May 2026 data into cr_ tables.

Sheets loaded (in FK-safe order):
  TARGETS, UFILL, QOC, SPEED, MS, HSD, MAK GE Reading, SANGAM, GOOGLE RATINGS

Usage:
  python scripts/ingest_may_data.py
  python scripts/ingest_may_data.py --dry-run
  python scripts/ingest_may_data.py --sheet UFILL
"""

import sys
import os
import argparse
import time as _time
import datetime

import openpyxl
import psycopg2
import psycopg2.extras

EXCEL_PATH = os.path.join(os.path.dirname(__file__), '..', 'docs', 'crystal ', 'MAY DATA.xlsx')
DB_URL = os.environ.get('BPCL_DB_URL', 'postgres://bpcl:bpcl@localhost:5433/bpcl_portal')

MONTH_YEAR = datetime.date(2026, 5, 1)  # competition period for all sheets

SHEETS = ['TARGETS', 'UFILL', 'QOC', 'SPEED', 'MS', 'HSD',
          'MAK GE Reading', 'SANGAM', 'GOOGLE RATINGS']


def main():
    parser = argparse.ArgumentParser(description='Ingest MAY DATA.xlsx into bpcl_portal DB')
    parser.add_argument('--dry-run', action='store_true', help='Parse without writing to DB')
    parser.add_argument('--sheet', metavar='SHEET', help='Run only one sheet by name')
    args = parser.parse_args()

    print(f'Loading workbook: {EXCEL_PATH}')
    wb = openpyxl.load_workbook(EXCEL_PATH, read_only=True, data_only=True)

    conn = None if args.dry_run else psycopg2.connect(DB_URL)
    if conn:
        conn.autocommit = False

    stats = {}   # sheet → {inserted, updated, skipped, errors}
    t0 = _time.monotonic()

    sheets_to_run = [args.sheet] if args.sheet else SHEETS
    for name in sheets_to_run:
        if name not in wb.sheetnames:
            print(f'  [WARN] Sheet "{name}" not found — skipping')
            continue
        ws = wb[name]
        rows = list(ws.iter_rows(values_only=True))
        print(f'\nProcessing sheet: {name} ({len(rows)} rows)')

        fn = LOADERS.get(name)
        if fn is None:
            print(f'  [WARN] No loader for sheet "{name}"')
            continue

        s = fn(rows, conn, args.dry_run)
        stats[name] = s
        print(f'  → inserted={s["inserted"]} updated={s["updated"]} skipped={s["skipped"]} errors={s["errors"]}')

    elapsed_ms = int((_time.monotonic() - t0) * 1000)

    if conn:
        conn.commit()
        log_etl_run(conn, stats, elapsed_ms)
        conn.commit()
        conn.close()

    print(f'\nDone in {elapsed_ms}ms.')
    if args.dry_run:
        print('DRY RUN — no data written.')


def empty_stats():
    return {'inserted': 0, 'updated': 0, 'skipped': 0, 'errors': 0}


def upsert(cur, sql, values, stats):
    """Execute one upsert row. Increments inserted/updated/errors in stats."""
    try:
        cur.execute(sql, values)
        if cur.rowcount == 0:
            stats['skipped'] += 1
        else:
            # heuristic: if the row already existed, rowcount=1 but it's an update
            stats['inserted'] += 1
    except Exception as e:
        stats['errors'] += 1
        print(f'    [ERROR] {e} | values={values}')


def log_etl_run(conn, stats, elapsed_ms):
    total_err = sum(s['errors'] for s in stats.values())
    status = 'ok' if total_err == 0 else 'error'
    detail = str({k: v for k, v in stats.items()})
    with conn.cursor() as cur:
        cur.execute(
            "INSERT INTO cr_etl_log (status, detail, duration_ms, source) VALUES (%s, %s, %s, %s)",
            (status, detail, elapsed_ms, 'may-data-etl'),
        )


# Loaders registry — populated as each loader function is defined below
LOADERS = {}
```

- [ ] **Step 3: Verify the skeleton parses without error**

```bash
python3 scripts/ingest_may_data.py --help
```

Expected: usage message listing `--dry-run` and `--sheet`.

- [ ] **Step 4: Commit skeleton**

```bash
git add scripts/ingest_may_data.py
git commit -m "feat: ETL script skeleton with CLI and DB wiring"
```

---

### Task 2: Dealers + Competition Period Upsert

**Files:**
- Modify: `scripts/ingest_may_data.py`

- [ ] **Step 1: Add ensure_dealers_and_period() and TARGETS loader**

Append to `scripts/ingest_may_data.py` (before the `if __name__ == '__main__':` at end, or at bottom of file):

```python
def ensure_dealers(rows, conn, dry_run):
    """Upsert dealers from TARGETS sheet rows (skipping title + header rows)."""
    s = empty_stats()
    if dry_run:
        for row in rows[2:]:  # row 0 = 'TARGETS', row 1 = headers
            if row[0] and row[1]:
                print(f'    [DRY] dealer cc={int(row[1])} name={row[0]}')
                s['inserted'] += 1
        return s

    sql = """
        INSERT INTO cr_dealers (cc_code, ro_name)
        VALUES (%s, %s)
        ON CONFLICT (cc_code) DO UPDATE SET ro_name = EXCLUDED.ro_name, updated_at = NOW()
    """
    with conn.cursor() as cur:
        for row in rows[2:]:
            if not row[0] or not row[1]:
                s['skipped'] += 1
                continue
            upsert(cur, sql, (str(int(row[1])), str(row[0]).strip()), s)
    return s


def ensure_competition_period(conn, dry_run):
    """Create the May 2026 competition period if it doesn't already exist."""
    if dry_run:
        print(f'  [DRY] ensure competition period {MONTH_YEAR}')
        return
    with conn.cursor() as cur:
        cur.execute("""
            INSERT INTO cr_competition_periods (name, month_year, total_slots, is_active)
            VALUES (%s, %s, 40, FALSE)
            ON CONFLICT (month_year) DO NOTHING
        """, ('May 2026', MONTH_YEAR))
```

- [ ] **Step 2: Call ensure_dealers and ensure_competition_period in main()**

In the `main()` function, before the `for name in sheets_to_run:` loop, add:

```python
    # Ensure dealers and period exist first (FK prerequisites)
    if not args.sheet or args.sheet == 'TARGETS':
        targets_ws = wb['TARGETS']
        targets_rows = list(targets_ws.iter_rows(values_only=True))
        print('\nUpserting dealers from TARGETS...')
        ds = ensure_dealers(targets_rows, conn, args.dry_run)
        stats['_dealers'] = ds

    if not args.dry_run:
        ensure_competition_period(conn, args.dry_run)
        if conn:
            conn.commit()
```

- [ ] **Step 3: Dry-run test**

```bash
python3 scripts/ingest_may_data.py --dry-run --sheet TARGETS 2>&1 | head -30
```

Expected: lines like `[DRY] dealer cc=112385 name=ANAND SUPER SERVICE STN.`

- [ ] **Step 4: Commit**

```bash
git add scripts/ingest_may_data.py
git commit -m "feat: ETL dealers + competition period upsert"
```

---

### Task 3: TARGETS Sheet Loader

**Files:**
- Modify: `scripts/ingest_may_data.py`

- [ ] **Step 1: Add load_targets() and register it**

Append to `scripts/ingest_may_data.py`:

```python
def _parse_bool(val):
    if isinstance(val, bool):
        return val
    if isinstance(val, str):
        return val.strip().upper() == 'YES'
    return False


def load_targets(rows, conn, dry_run):
    """TARGETS sheet → cr_monthly_targets. rows[0]=title, rows[1]=headers, rows[2:]=data."""
    s = empty_stats()
    sql = """
        INSERT INTO cr_monthly_targets
          (cc_code, month_year, ufill_target, qoc_target, speed_kl, ms_kl, hsd_kl,
           ms_ly, hsd_ly, dsw_available, nitrogen, mak_ge_target, darpan_target,
           coolant_lube_kl, remarks)
        VALUES (%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s,%s)
        ON CONFLICT (cc_code, month_year) DO UPDATE SET
          ufill_target=EXCLUDED.ufill_target, qoc_target=EXCLUDED.qoc_target,
          speed_kl=EXCLUDED.speed_kl, ms_kl=EXCLUDED.ms_kl, hsd_kl=EXCLUDED.hsd_kl,
          ms_ly=EXCLUDED.ms_ly, hsd_ly=EXCLUDED.hsd_ly,
          dsw_available=EXCLUDED.dsw_available, nitrogen=EXCLUDED.nitrogen,
          mak_ge_target=EXCLUDED.mak_ge_target, darpan_target=EXCLUDED.darpan_target,
          coolant_lube_kl=EXCLUDED.coolant_lube_kl, remarks=EXCLUDED.remarks
    """
    # Columns: [0]=RO Name, [1]=CC, [2]=UFILL, [3]=OilChanges, [4]=SPEED,
    #          [5]=MS, [6]=HSD, [7]=MS_LY, [8]=HSD_LY, [9]=DSW, [10]=Nitrogen,
    #          [11]=Remarks, [12]=MakGE, [13]=Darpan, [14]=Coolant
    for row in rows[2:]:
        if not row[0] or not row[1]:
            s['skipped'] += 1
            continue
        try:
            cc = str(int(row[1]))
        except (TypeError, ValueError):
            s['skipped'] += 1
            continue

        vals = (
            cc, MONTH_YEAR,
            int(row[2]) if row[2] is not None else None,   # ufill_target
            None,                                           # qoc_target (not in sheet)
            float(row[4]) if row[4] is not None else None, # speed_kl
            float(row[5]) if row[5] is not None else None, # ms_kl
            float(row[6]) if row[6] is not None else None, # hsd_kl
            float(row[7]) if row[7] is not None else None, # ms_ly
            float(row[8]) if row[8] is not None else None, # hsd_ly
            _parse_bool(row[9]),                            # dsw_available
            _parse_bool(row[10]),                           # nitrogen
            float(row[12]) if row[12] is not None else None, # mak_ge_target
            int(row[13]) if row[13] is not None else None,   # darpan_target
            float(row[14]) if row[14] is not None else None, # coolant_lube_kl
            str(row[11]).strip() if row[11] else None,       # remarks
        )
        if dry_run:
            print(f'    [DRY] target cc={cc} ms_kl={vals[5]}')
            s['inserted'] += 1
            continue
        with conn.cursor() as cur:
            upsert(cur, sql, vals, s)
    return s


LOADERS['TARGETS'] = load_targets
```

- [ ] **Step 2: Dry-run TARGETS**

```bash
python3 scripts/ingest_may_data.py --dry-run --sheet TARGETS 2>&1 | tail -10
```

Expected: lines like `[DRY] target cc=112385 ms_kl=456.0`, then summary.

- [ ] **Step 3: Commit**

```bash
git add scripts/ingest_may_data.py
git commit -m "feat: ETL TARGETS loader"
```

---

### Task 4: Daily Volume Loaders — UFILL, QOC, MS, HSD, SPEED

**Files:**
- Modify: `scripts/ingest_may_data.py`

- [ ] **Step 1: Add generic daily loader helper**

Append to `scripts/ingest_may_data.py`:

```python
def _load_daily(rows, conn, dry_run, table, value_col, value_type=float):
    """
    Generic loader for sheets with layout:
      row[0] = title, row[1] = headers: [RO Name, CC Code, date1, date2, ..., MTD]
      rows[2:] = data

    table: target table name (cr_daily_ufill, cr_daily_ms, etc.)
    value_col: 'count' for integer tables, 'kl' for volume tables
    value_type: int or float
    """
    s = empty_stats()
    # Parse date columns from row[1]: indices 2..N-1, last col is MTD (skip)
    headers = rows[1]
    date_cols = []  # list of (col_index, date)
    for i, h in enumerate(headers):
        if i < 2:
            continue
        if isinstance(h, datetime.datetime):
            date_cols.append((i, h.date()))
        # skip MTD and None columns

    cast = int if value_type == int else float
    sql = f"""
        INSERT INTO {table} (cc_code, txn_date, {value_col})
        VALUES (%s, %s, %s)
        ON CONFLICT (cc_code, txn_date) DO UPDATE SET {value_col} = EXCLUDED.{value_col}
    """

    for row in rows[2:]:
        if not row[1]:
            s['skipped'] += 1
            continue
        try:
            cc = str(int(row[1]))
        except (TypeError, ValueError):
            s['skipped'] += 1
            continue

        for col_idx, txn_date in date_cols:
            val = row[col_idx] if col_idx < len(row) else None
            if val is None:
                continue
            try:
                cast_val = cast(val)
            except (TypeError, ValueError):
                s['errors'] += 1
                continue
            if cast_val < 0:
                s['skipped'] += 1
                continue

            if dry_run:
                print(f'    [DRY] {table} cc={cc} date={txn_date} {value_col}={cast_val}')
                s['inserted'] += 1
                continue

            with conn.cursor() as cur:
                upsert(cur, sql, (cc, txn_date, cast_val), s)
    return s
```

- [ ] **Step 2: Add UFILL and QOC loaders**

Append to `scripts/ingest_may_data.py`:

```python
def load_ufill(rows, conn, dry_run):
    return _load_daily(rows, conn, dry_run, 'cr_daily_ufill', 'count', int)

def load_qoc(rows, conn, dry_run):
    return _load_daily(rows, conn, dry_run, 'cr_daily_qoc', 'count', int)

LOADERS['UFILL'] = load_ufill
LOADERS['QOC'] = load_qoc
```

- [ ] **Step 3: Add MS and HSD loaders**

Append to `scripts/ingest_may_data.py`:

```python
def load_ms(rows, conn, dry_run):
    return _load_daily(rows, conn, dry_run, 'cr_daily_ms', 'kl', float)

def load_hsd(rows, conn, dry_run):
    return _load_daily(rows, conn, dry_run, 'cr_daily_hsd', 'kl', float)

LOADERS['MS'] = load_ms
LOADERS['HSD'] = load_hsd
```

- [ ] **Step 4: Add SPEED loader**

The SPEED sheet has 3 header columns: `[Retail Outlet, RO ID, Product, date1, ...]`. CC code is at index 1, daily data starts at index 3.

Append to `scripts/ingest_may_data.py`:

```python
def load_speed(rows, conn, dry_run):
    """SPEED sheet: col[0]=RO Name, col[1]=CC, col[2]=Product, col[3..N-1]=dates, last=MTD."""
    s = empty_stats()
    headers = rows[1]
    date_cols = []
    for i, h in enumerate(headers):
        if i < 3:
            continue
        if isinstance(h, datetime.datetime):
            date_cols.append((i, h.date()))

    sql = """
        INSERT INTO cr_daily_speed (cc_code, txn_date, kl)
        VALUES (%s, %s, %s)
        ON CONFLICT (cc_code, txn_date) DO UPDATE SET kl = EXCLUDED.kl
    """

    for row in rows[2:]:
        if not row[1]:
            s['skipped'] += 1
            continue
        try:
            cc = str(int(row[1]))
        except (TypeError, ValueError):
            s['skipped'] += 1
            continue

        for col_idx, txn_date in date_cols:
            val = row[col_idx] if col_idx < len(row) else None
            if val is None:
                continue
            try:
                kl = float(val)
            except (TypeError, ValueError):
                s['errors'] += 1
                continue
            if kl < 0:
                s['skipped'] += 1
                continue

            if dry_run:
                print(f'    [DRY] cr_daily_speed cc={cc} date={txn_date} kl={kl}')
                s['inserted'] += 1
                continue

            with conn.cursor() as cur:
                upsert(cur, sql, (cc, txn_date, kl), s)
    return s

LOADERS['SPEED'] = load_speed
```

- [ ] **Step 5: Dry-run all daily sheets**

```bash
python3 scripts/ingest_may_data.py --dry-run --sheet UFILL 2>&1 | tail -5
python3 scripts/ingest_may_data.py --dry-run --sheet MS 2>&1 | tail -5
python3 scripts/ingest_may_data.py --dry-run --sheet SPEED 2>&1 | tail -5
```

Expected: summary lines showing inserted counts > 0.

- [ ] **Step 6: Commit**

```bash
git add scripts/ingest_may_data.py
git commit -m "feat: ETL daily loaders (UFILL, QOC, MS, HSD, SPEED)"
```

---

### Task 5: MAK GE, SANGAM, GOOGLE RATINGS Loaders

**Files:**
- Modify: `scripts/ingest_may_data.py`

- [ ] **Step 1: Add MAK GE Reading loader**

The sheet layout: row[0]=headers (RO Name, CC Code, date1..date5, _, TOTAL, _).
Date headers are strings like `'MAK GE Reading as on 1st May'` — parse dates from col indices 2..6.

Append to `scripts/ingest_may_data.py`:

```python
_MAK_GE_DATES = [
    (2, datetime.date(2026, 5, 1)),
    (3, datetime.date(2026, 5, 8)),
    (4, datetime.date(2026, 5, 15)),
    (5, datetime.date(2026, 5, 22)),
    (6, datetime.date(2026, 5, 31)),
]

def load_mak_ge(rows, conn, dry_run):
    """MAK GE Reading → cr_mak_ge_readings. row[0]=headers, rows[1:]=data."""
    s = empty_stats()
    sql = """
        INSERT INTO cr_mak_ge_readings (cc_code, reading_date, meter_reading)
        VALUES (%s, %s, %s)
        ON CONFLICT (cc_code, reading_date) DO UPDATE SET meter_reading = EXCLUDED.meter_reading
    """
    for row in rows[1:]:
        if not row[1]:
            s['skipped'] += 1
            continue
        try:
            cc = str(int(row[1]))
        except (TypeError, ValueError):
            s['skipped'] += 1
            continue

        for col_idx, reading_date in _MAK_GE_DATES:
            val = row[col_idx] if col_idx < len(row) else None
            if val is None:
                continue
            try:
                reading = float(val)
            except (TypeError, ValueError):
                s['errors'] += 1
                continue

            if dry_run:
                print(f'    [DRY] cr_mak_ge_readings cc={cc} date={reading_date} reading={reading}')
                s['inserted'] += 1
                continue

            with conn.cursor() as cur:
                upsert(cur, sql, (cc, reading_date, reading), s)
    return s

LOADERS['MAK GE Reading'] = load_mak_ge
```

- [ ] **Step 2: Add SANGAM loader**

Append to `scripts/ingest_may_data.py`:

```python
def load_sangam(rows, conn, dry_run):
    """SANGAM sheet → cr_sangam_data. row[0]=title, row[1]=headers, rows[2:]=data.
    Columns: [0]=Sr No (skip), [1]=RO Name, [2]=CC Code... wait — let's check actual layout.
    Actual: row[1] = ('RO NAME', 'CC Code', 'Certificates', 'Status', 'REMARKS')
    """
    s = empty_stats()
    sql = """
        INSERT INTO cr_sangam_data (cc_code, month_year, cert_count, status, remarks)
        VALUES (%s, %s, %s, %s, %s)
        ON CONFLICT (cc_code, month_year) DO UPDATE SET
          cert_count=EXCLUDED.cert_count, status=EXCLUDED.status,
          remarks=EXCLUDED.remarks, updated_at=NOW()
    """
    # rows[0]='SANGAM' title, rows[1]=headers, rows[2:]=data
    # col[0]=RO Name, col[1]=CC Code, col[2]=Certificates, col[3]=Status, col[4]=Remarks
    for row in rows[2:]:
        if not row[1]:
            s['skipped'] += 1
            continue
        try:
            cc = str(int(row[1]))
        except (TypeError, ValueError):
            s['skipped'] += 1
            continue

        cert = int(row[2]) if row[2] is not None else 0
        status = str(row[3]).strip() if row[3] else None
        remarks = str(row[4]).strip() if row[4] else None

        if dry_run:
            print(f'    [DRY] cr_sangam_data cc={cc} status={status}')
            s['inserted'] += 1
            continue

        with conn.cursor() as cur:
            upsert(cur, sql, (cc, MONTH_YEAR, cert, status, remarks), s)
    return s

LOADERS['SANGAM'] = load_sangam
```

- [ ] **Step 3: Add GOOGLE RATINGS loader**

Sheet layout: row[0]=title, row[1]=date headers (3 snapshot dates), row[2]=col labels (CC Code, RO Name, Ratings, Reviews, None, Ratings, Reviews, ...).

Three snapshots per row at column offsets:
- Snapshot 1 (2026-05-01): cols 2,3
- Snapshot 2 (2026-05-15): cols 5,6
- Snapshot 3 (2026-05-25): cols 8,9

Append to `scripts/ingest_may_data.py`:

```python
_RATING_SNAPSHOTS = [
    (2, 3, datetime.date(2026, 5, 1)),
    (5, 6, datetime.date(2026, 5, 15)),
    (8, 9, datetime.date(2026, 5, 25)),
]

def load_google_ratings(rows, conn, dry_run):
    """GOOGLE RATINGS → cr_google_ratings.
    rows[0]=title, rows[1]=date row, rows[2]=col headers, rows[3:]=data.
    col[0]=CC Code, col[1]=RO Name, then triplets (Rating, Reviews, None) per snapshot.
    """
    s = empty_stats()
    sql = """
        INSERT INTO cr_google_ratings (cc_code, snapshot_date, rating, review_count)
        VALUES (%s, %s, %s, %s)
        ON CONFLICT (cc_code, snapshot_date) DO UPDATE SET
          rating=EXCLUDED.rating, review_count=EXCLUDED.review_count
    """
    for row in rows[3:]:
        if not row[0]:
            s['skipped'] += 1
            continue
        try:
            cc = str(int(row[0]))
        except (TypeError, ValueError):
            s['skipped'] += 1
            continue

        for rating_col, reviews_col, snap_date in _RATING_SNAPSHOTS:
            r_val = row[rating_col] if rating_col < len(row) else None
            rv_val = row[reviews_col] if reviews_col < len(row) else None
            if r_val is None:
                continue
            try:
                rating = float(r_val)
                reviews = int(rv_val) if rv_val is not None else 0
            except (TypeError, ValueError):
                s['errors'] += 1
                continue
            if not (1.0 <= rating <= 5.0):
                s['skipped'] += 1
                continue

            if dry_run:
                print(f'    [DRY] cr_google_ratings cc={cc} date={snap_date} rating={rating} reviews={reviews}')
                s['inserted'] += 1
                continue

            with conn.cursor() as cur:
                upsert(cur, sql, (cc, snap_date, rating, reviews), s)
    return s

LOADERS['GOOGLE RATINGS'] = load_google_ratings
```

- [ ] **Step 4: Dry-run all three sheets**

```bash
python3 scripts/ingest_may_data.py --dry-run --sheet 'MAK GE Reading' 2>&1 | tail -5
python3 scripts/ingest_may_data.py --dry-run --sheet SANGAM 2>&1 | tail -5
python3 scripts/ingest_may_data.py --dry-run --sheet 'GOOGLE RATINGS' 2>&1 | tail -5
```

Expected: summary lines with non-zero counts (or zeroes if all readings are None — that's valid for MAK GE).

- [ ] **Step 5: Commit**

```bash
git add scripts/ingest_may_data.py
git commit -m "feat: ETL MAK GE, SANGAM, GOOGLE RATINGS loaders"
```

---

### Task 6: Full Run + Verify

**Files:**
- None (verification only)

- [ ] **Step 1: Full dry run — all sheets**

```bash
python3 scripts/ingest_may_data.py --dry-run 2>&1
```

Expected: all 9 sheets processed with no Python exceptions. Errors/skips are OK.

- [ ] **Step 2: Ensure DB is running**

```bash
docker compose up -d db
sleep 2
psql postgres://bpcl:bpcl@localhost:5433/bpcl_portal -c '\dt cr_*'
```

Expected: lists `cr_dealers`, `cr_monthly_targets`, `cr_daily_ufill`, etc.

- [ ] **Step 3: Run migrations if needed**

```bash
bash scripts/migrate.sh up
```

- [ ] **Step 4: Run the real ingest**

```bash
python3 scripts/ingest_may_data.py
```

Expected: each sheet shows `inserted > 0`, `errors = 0` (or low), final `Done in Xms`.

- [ ] **Step 5: Spot-check key tables**

```bash
psql postgres://bpcl:bpcl@localhost:5433/bpcl_portal -c \
  "SELECT cc_code, ro_name FROM cr_dealers LIMIT 5;"

psql postgres://bpcl:bpcl@localhost:5433/bpcl_portal -c \
  "SELECT cc_code, ms_kl, hsd_kl FROM cr_monthly_targets WHERE month_year='2026-05-01' LIMIT 5;"

psql postgres://bpcl:bpcl@localhost:5433/bpcl_portal -c \
  "SELECT cc_code, txn_date, count FROM cr_daily_ufill ORDER BY txn_date LIMIT 10;"

psql postgres://bpcl:bpcl@localhost:5433/bpcl_portal -c \
  "SELECT status, detail FROM cr_etl_log ORDER BY run_at DESC LIMIT 1;"
```

Expected: rows present, `cr_etl_log` shows `status='ok'`.

- [ ] **Step 6: Idempotency check — re-run should be clean**

```bash
python3 scripts/ingest_may_data.py
```

Expected: all `errors=0`. Row counts may differ (upserts don't add duplicates).

- [ ] **Step 7: Commit final state**

```bash
git add scripts/ingest_may_data.py
git commit -m "feat: complete MAY DATA ETL pipeline"
```

---

## Done

Run the full ETL any time with:
```bash
python3 scripts/ingest_may_data.py
```

Re-runs are safe (all upserts are idempotent). For a single sheet:
```bash
python3 scripts/ingest_may_data.py --sheet UFILL
```
