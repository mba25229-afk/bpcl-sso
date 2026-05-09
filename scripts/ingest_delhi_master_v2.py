#!/usr/bin/env python3
"""
Delhi Master Excel ingestion — v2.

Parses docs/crystal/Delhi Master.xlsx, Sheet: Data.
Columns J through BE (0-indexed 9–56) contain MS and HSD volumes.

Column naming convention: MS-MMYY or HSD-MMYY (dash optional).
  MM = month (01–12), YY = 2-digit year (24=2024, 25=2025, 26=2026).

Each month has exactly 2 columns (TY and LY):
  e.g. MS-0326 (MS Mar 2026) and MS-0325 (MS Mar 2025) sit side by side.

This script:
  1. Reads all columns J–BE, decodes their month from the header.
  2. For each row (outlet), upserts ONE market_share_data row per month
     using the TY column value.
  3. Resolves trading_area_id via fuzzy name match → creates missing areas.
  4. Recomputes trading_area_totals for every affected period.

Usage:
  python scripts/ingest_delhi_master_v2.py
  python scripts/ingest_delhi_master_v2.py --dry-run
  python scripts/ingest_delhi_master_v2.py --period 2026-03  # single month
"""

import sys
import os
import re
import argparse
import datetime
from collections import defaultdict

import openpyxl
import psycopg2
import psycopg2.extras

EXCEL_PATH = '/Users/meynesh/Documents/bpcl-sso/docs/crystal /Delhi Master.xlsx'
DB_URL = os.environ.get('BPCL_DB_URL',
    'postgres://bpcl:bpcl@localhost:5433/bpcl_portal')

OMC_MAP = {'BPC': 'BPCL', 'HPC': 'HPCL', 'IOC': 'IOCL'}

# ── column parsing ─────────────────────────────────────────────────────────────

def parse_header(h):
    """Return (metric, year, month) from a header like 'MS-0326' or 'HSD0424'.
    Returns None for unrecognised headers."""
    if not h:
        return None
    h = str(h).strip()
    m = re.match(r'^(MS|HSD)-?(\d{2})(\d{2})$', h, re.IGNORECASE)
    if not m:
        return None
    metric = m.group(1).upper()
    mm = m.group(2)
    yy = m.group(3)
    year = int('20' + yy)
    month = int(mm)
    return metric, year, month


def build_column_map(ws):
    """Scan row 1 cols J–BE (1-based 10–57) → dict of 0-based-idx -> (metric,year,month)."""
    col_map = {}
    for col1 in range(10, 58):   # col 10 = J, col 57 = BE
        idx0 = col1 - 1          # 0-based index for row tuple access
        header = ws.cell(row=1, column=col1).value
        parsed = parse_header(header)
        if parsed:
            col_map[idx0] = parsed
    return col_map


def select_ty_columns(col_map):
    """Return ALL columns — each (metric, year, month) is a distinct period.
    Every column in J–BE represents a real month's volume worth ingesting."""
    # Each header uniquely identifies (metric, year, month).
    # Ingest all of them — the DB UNIQUE constraint prevents duplicates.
    return {(metric, year, month): idx0
            for idx0, (metric, year, month) in col_map.items()}


# ── trading area helpers ───────────────────────────────────────────────────────

def normalise_ta(name):
    """Normalise trading area name for fuzzy matching."""
    return re.sub(r'[^a-z0-9]', '', str(name).lower())


def load_trading_areas(cur):
    cur.execute("SELECT id, name FROM trading_areas")
    rows = cur.fetchall()
    by_id = {r[0]: r[1] for r in rows}
    by_norm = {normalise_ta(r[1]): r[0] for r in rows}
    return by_id, by_norm


def resolve_or_create_ta(cur, raw_name, district, by_norm, by_id, dry_run):
    norm = normalise_ta(raw_name)
    if norm in by_norm:
        return by_norm[norm]
    # Create new — district required NOT NULL; default 'Delhi' if missing
    display = raw_name.replace('_', ' ').title()
    dist = (district or 'Delhi').strip()
    if dry_run:
        fake_id = -(len(by_id) + 1)
        by_norm[norm] = fake_id
        by_id[fake_id] = display
        return fake_id
    cur.execute(
        "INSERT INTO trading_areas (name, district) VALUES (%s, %s) RETURNING id",
        (display, dist))
    new_id = cur.fetchone()[0]
    by_norm[norm] = new_id
    by_id[new_id] = display
    print(f"  Created trading_area: id={new_id} name={display!r} district={dist!r}")
    return new_id


# ── main ──────────────────────────────────────────────────────────────────────

def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('--dry-run', action='store_true')
    ap.add_argument('--period', help='Only ingest this month, e.g. 2026-03')
    ap.add_argument('--verbose', action='store_true')
    args = ap.parse_args()

    only_period = None
    if args.period:
        y, m = args.period.split('-')
        only_period = (int(y), int(m))

    print(f"Loading: {EXCEL_PATH}")
    wb = openpyxl.load_workbook(EXCEL_PATH, data_only=True)
    ws = wb['Data']
    print(f"Sheet 'Data': {ws.max_row} rows x {ws.max_column} cols")

    col_map = build_column_map(ws)
    ty_cols = select_ty_columns(col_map)
    print(f"Found {len(ty_cols)} TY columns covering "
          f"{len(set((y,m) for _,y,m in ty_cols))} months")
    for (metric, year, month), idx0 in sorted(ty_cols.items()):
        period_str = f"{year}-{month:02d}-01"
        if args.verbose:
            print(f"  {metric:3s} {period_str} <- col idx {idx0} "
                  f"(header={ws.cell(row=1,column=idx0+1).value})")

    conn = psycopg2.connect(DB_URL)
    conn.autocommit = False
    cur = conn.cursor()

    by_id, by_norm = load_trading_areas(cur)
    print(f"Loaded {len(by_id)} trading areas from DB")

    # Collect all rows to upsert
    upsert_rows = []   # (outlet_name, cc_number, omc, ta_id, period, ms_vol, hsd_vol)
    skipped = 0
    periods_seen = set()

    for ri, row in enumerate(ws.iter_rows(min_row=2, values_only=True), start=2):
        if not row or row[1] is None:
            continue
        omc_raw = str(row[1]).strip().upper()
        if omc_raw not in OMC_MAP:
            continue
        omc = OMC_MAP[omc_raw]

        cc_raw = row[0]
        cc_number = str(int(cc_raw)) if cc_raw is not None else None
        if omc != 'BPCL':
            cc_number = None  # only BPCL outlets have cc_numbers in our system

        outlet_name = str(row[2]).strip() if row[2] else f"Row-{ri}"
        ta_raw = row[5] or row[6] or 'Unknown'
        ta_name = str(ta_raw).strip()

        district = str(row[4]).strip() if row[4] else 'Delhi'
        ta_id = resolve_or_create_ta(cur, ta_name, district, by_norm, by_id, args.dry_run)

        # Group TY columns by (year, month) — build MS+HSD pair per month
        month_data = defaultdict(lambda: {'ms': None, 'hsd': None})
        for (metric, year, month), idx0 in ty_cols.items():
            if only_period and (year, month) != only_period:
                continue
            if idx0 >= len(row):
                continue
            val = row[idx0]
            if val is None:
                continue
            if isinstance(val, str):
                if '#DIV' in val.upper() or '#N/A' in val.upper():
                    continue
                try:
                    val = float(val)
                except ValueError:
                    continue
            val = float(val)
            if val < 0:
                val = 0.0
            month_key = (year, month)
            if metric == 'MS':
                month_data[month_key]['ms'] = val
            else:
                month_data[month_key]['hsd'] = val

        for (year, month), vols in month_data.items():
            ms_vol = vols['ms'] or 0.0
            hsd_vol = vols['hsd'] or 0.0
            if ms_vol == 0 and hsd_vol == 0:
                skipped += 1
                continue
            period_str = f"{year}-{month:02d}-01"
            periods_seen.add(period_str)
            upsert_rows.append((
                outlet_name, cc_number, omc, ta_id, period_str, ms_vol, hsd_vol
            ))
            if args.verbose:
                print(f"  Row {ri}: {outlet_name[:20]:20s} {omc} {period_str} "
                      f"MS={ms_vol:.1f} HSD={hsd_vol:.1f}")

    print(f"\nParsed {len(upsert_rows)} outlet-month rows "
          f"({skipped} zero/null skipped)")
    print(f"Periods: {sorted(periods_seen)}")

    if args.dry_run:
        print("\n[DRY RUN] No DB writes.")
        conn.close()
        return

    # Upsert market_share_data
    print("\nUpserting market_share_data...")
    psycopg2.extras.execute_batch(cur, """
        INSERT INTO market_share_data
            (outlet_name, cc_number, omc, trading_area_id, period, ms_vol_kl, hsd_vol_kl, source)
        VALUES (%s, %s, %s, %s, %s, %s, %s, 'delhi_master')
        ON CONFLICT (outlet_name, omc, trading_area_id, period) DO UPDATE SET
            ms_vol_kl  = EXCLUDED.ms_vol_kl,
            hsd_vol_kl = EXCLUDED.hsd_vol_kl,
            source     = 'delhi_master'
    """, upsert_rows, page_size=500)
    print(f"  Upserted {len(upsert_rows)} rows")

    # Recompute trading_area_totals for all affected periods
    print("\nRecomputing trading_area_totals...")
    for period in sorted(periods_seen):
        cur.execute("""
            INSERT INTO trading_area_totals
                (trading_area_id, period,
                 total_ms_kl, total_hsd_kl,
                 bpcl_ms_kl, hpcl_ms_kl, iocl_ms_kl,
                 bpcl_hsd_kl, hpcl_hsd_kl, iocl_hsd_kl,
                 computed_at)
            SELECT
                trading_area_id,
                period,
                SUM(ms_vol_kl)                                           AS total_ms_kl,
                SUM(hsd_vol_kl)                                          AS total_hsd_kl,
                SUM(ms_vol_kl)  FILTER (WHERE omc = 'BPCL')             AS bpcl_ms_kl,
                SUM(ms_vol_kl)  FILTER (WHERE omc = 'HPCL')             AS hpcl_ms_kl,
                SUM(ms_vol_kl)  FILTER (WHERE omc = 'IOCL')             AS iocl_ms_kl,
                SUM(hsd_vol_kl) FILTER (WHERE omc = 'BPCL')             AS bpcl_hsd_kl,
                SUM(hsd_vol_kl) FILTER (WHERE omc = 'HPCL')             AS hpcl_hsd_kl,
                SUM(hsd_vol_kl) FILTER (WHERE omc = 'IOCL')             AS iocl_hsd_kl,
                NOW()
            FROM market_share_data
            WHERE period = %s
            GROUP BY trading_area_id, period
            ON CONFLICT (trading_area_id, period) DO UPDATE SET
                total_ms_kl  = EXCLUDED.total_ms_kl,
                total_hsd_kl = EXCLUDED.total_hsd_kl,
                bpcl_ms_kl   = EXCLUDED.bpcl_ms_kl,
                hpcl_ms_kl   = EXCLUDED.hpcl_ms_kl,
                iocl_ms_kl   = EXCLUDED.iocl_ms_kl,
                bpcl_hsd_kl  = EXCLUDED.bpcl_hsd_kl,
                hpcl_hsd_kl  = EXCLUDED.hpcl_hsd_kl,
                iocl_hsd_kl  = EXCLUDED.iocl_hsd_kl,
                computed_at  = NOW()
        """, (period,))

    print(f"  Recomputed totals for {len(periods_seen)} periods")

    conn.commit()
    cur.close()
    conn.close()
    print("\n✓ Done.")

    # Quick sanity check
    conn2 = psycopg2.connect(DB_URL)
    cur2 = conn2.cursor()
    cur2.execute("""
        SELECT period, omc, COUNT(*), ROUND(SUM(ms_vol_kl)::numeric,0) AS ms_total
        FROM market_share_data
        WHERE period IN %s
        GROUP BY period, omc
        ORDER BY period, omc
    """, (tuple(sorted(periods_seen)),))
    print("\nSanity check — market_share_data rows per period/OMC:")
    for row in cur2.fetchall():
        print(f"  {row[0]}  {row[1]:4s}  {row[2]:4d} outlets  MS={row[3]} KL")

    cur2.execute("""
        SELECT period, ROUND(total_ms_kl::numeric,0), ROUND(bpcl_ms_kl::numeric,0),
               ROUND(bpcl_ms_kl/NULLIF(total_ms_kl,0)*100::numeric,1) AS bpcl_ms_pct
        FROM trading_area_totals
        WHERE period IN %s
        GROUP BY period, total_ms_kl, bpcl_ms_kl
        ORDER BY period LIMIT 10
    """, (tuple(sorted(periods_seen)),))
    print("\nSample trading_area_totals (first 10):")
    for row in cur2.fetchall():
        print(f"  {row[0]}  total_ms={row[1]} bpcl_ms={row[2]} BPCL%={row[3]}%")
    cur2.close()
    conn2.close()


if __name__ == '__main__':
    main()
