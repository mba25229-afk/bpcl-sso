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

    stats = {}   # sheet -> {inserted, updated, skipped, errors}
    t0 = _time.monotonic()

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
        print(f'  -> inserted={s["inserted"]} updated={s["updated"]} skipped={s["skipped"]} errors={s["errors"]}')

    elapsed_ms = int((_time.monotonic() - t0) * 1000)

    if conn:
        conn.commit()
        log_etl_run(conn, stats, elapsed_ms)
        conn.commit()
        conn.close()

    print(f'\nDone in {elapsed_ms}ms.')
    if args.dry_run:
        print('DRY RUN -- no data written.')


def empty_stats():
    return {'inserted': 0, 'updated': 0, 'skipped': 0, 'errors': 0}


def upsert(cur, sql, values, stats):
    """Execute one upsert row. Increments inserted/updated/errors in stats."""
    try:
        cur.execute(sql, values)
        if cur.rowcount == 0:
            stats['skipped'] += 1
        else:
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


def _parse_bool(val):
    if isinstance(val, bool):
        return val
    if isinstance(val, str):
        return val.strip().upper() == 'YES'
    return False


def load_targets(rows, conn, dry_run):
    """TARGETS sheet -> cr_monthly_targets. rows[0]=title, rows[1]=headers, rows[2:]=data."""
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
    for row in rows[2:]:
        if not row[0] or not row[1]:
            s['skipped'] += 1
            continue
        try:
            cc = str(int(row[1]))
        except (TypeError, ValueError):
            s['skipped'] += 1
            continue

        def _col(r, i, cast=None):
            """Safely get column i from row r, return None if out of range or None value."""
            if i >= len(r) or r[i] is None:
                return None
            try:
                return cast(r[i]) if cast else r[i]
            except (TypeError, ValueError):
                return None

        vals = (
            cc, MONTH_YEAR,
            _col(row, 2, int),      # ufill_target
            None,                    # qoc_target (not in sheet)
            _col(row, 4, float),    # speed_kl
            _col(row, 5, float),    # ms_kl
            _col(row, 6, float),    # hsd_kl
            _col(row, 7, float),    # ms_ly
            _col(row, 8, float),    # hsd_ly
            _parse_bool(row[9]) if 9 < len(row) else False,   # dsw_available
            _parse_bool(row[10]) if 10 < len(row) else False,  # nitrogen
            _col(row, 12, float),   # mak_ge_target
            _col(row, 13, int),     # darpan_target
            _col(row, 14, float),   # coolant_lube_kl
            str(row[11]).strip() if 11 < len(row) and row[11] else None,  # remarks
        )
        if dry_run:
            print(f'    [DRY] target cc={cc} ms_kl={vals[5]}')
            s['inserted'] += 1
            continue
        with conn.cursor() as cur:
            upsert(cur, sql, vals, s)
    return s


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
    headers = rows[1]
    date_cols = []  # list of (col_index, date)
    for i, h in enumerate(headers):
        if i < 2:
            continue
        if isinstance(h, datetime.datetime):
            date_cols.append((i, h.date()))

    cast = int if value_type == int else float
    sql = f"""
        INSERT INTO {table} (cc_code, txn_date, {value_col})
        VALUES (%s, %s, %s)
        ON CONFLICT (cc_code, txn_date) DO UPDATE SET {value_col} = EXCLUDED.{value_col}
    """

    for row in rows[2:]:
        if len(row) < 2 or not row[1]:
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


def load_ufill(rows, conn, dry_run):
    return _load_daily(rows, conn, dry_run, 'cr_daily_ufill', 'count', int)

def load_qoc(rows, conn, dry_run):
    return _load_daily(rows, conn, dry_run, 'cr_daily_qoc', 'count', int)

def load_ms(rows, conn, dry_run):
    return _load_daily(rows, conn, dry_run, 'cr_daily_ms', 'kl', float)

def load_hsd(rows, conn, dry_run):
    return _load_daily(rows, conn, dry_run, 'cr_daily_hsd', 'kl', float)

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
        if len(row) < 2 or not row[1]:
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


_MAK_GE_DATES = [
    (2, datetime.date(2026, 5, 1)),
    (3, datetime.date(2026, 5, 8)),
    (4, datetime.date(2026, 5, 15)),
    (5, datetime.date(2026, 5, 22)),
    (6, datetime.date(2026, 5, 31)),
]

def load_mak_ge(rows, conn, dry_run):
    """MAK GE Reading -> cr_mak_ge_readings. row[0]=headers (NO title row), rows[1:]=data."""
    s = empty_stats()
    sql = """
        INSERT INTO cr_mak_ge_readings (cc_code, reading_date, meter_reading)
        VALUES (%s, %s, %s)
        ON CONFLICT (cc_code, reading_date) DO UPDATE SET meter_reading = EXCLUDED.meter_reading
    """
    for row in rows[1:]:
        if len(row) < 2 or not row[1]:
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


def load_sangam(rows, conn, dry_run):
    """SANGAM -> cr_sangam_data. row[0]=title, row[1]=headers, rows[2:]=data.
    col[0]=RO Name, col[1]=CC Code, col[2]=Certificates, col[3]=Status, col[4]=Remarks.
    """
    s = empty_stats()
    sql = """
        INSERT INTO cr_sangam_data (cc_code, month_year, cert_count, status, remarks)
        VALUES (%s, %s, %s, %s, %s)
        ON CONFLICT (cc_code, month_year) DO UPDATE SET
          cert_count=EXCLUDED.cert_count, status=EXCLUDED.status,
          remarks=EXCLUDED.remarks, updated_at=NOW()
    """
    for row in rows[2:]:
        if len(row) < 2 or not row[1]:
            s['skipped'] += 1
            continue
        try:
            cc = str(int(row[1]))
        except (TypeError, ValueError):
            s['skipped'] += 1
            continue

        cert = int(row[2]) if len(row) > 2 and row[2] is not None else 0
        status = str(row[3]).strip() if len(row) > 3 and row[3] else None
        remarks = str(row[4]).strip() if len(row) > 4 and row[4] else None

        if dry_run:
            print(f'    [DRY] cr_sangam_data cc={cc} status={status}')
            s['inserted'] += 1
            continue

        with conn.cursor() as cur:
            upsert(cur, sql, (cc, MONTH_YEAR, cert, status, remarks), s)
    return s


_RATING_SNAPSHOTS = [
    (2, 3, datetime.date(2026, 5, 1)),
    (5, 6, datetime.date(2026, 5, 15)),
    (8, 9, datetime.date(2026, 5, 25)),
]

def load_google_ratings(rows, conn, dry_run):
    """GOOGLE RATINGS -> cr_google_ratings.
    rows[0]=title, rows[1]=date row, rows[2]=col headers, rows[3:]=data.
    col[0]=CC Code (float), col[1]=RO Name, then triplets (Rating, Reviews, None) per snapshot.
    """
    s = empty_stats()
    sql = """
        INSERT INTO cr_google_ratings (cc_code, snapshot_date, rating, review_count)
        VALUES (%s, %s, %s, %s)
        ON CONFLICT (cc_code, snapshot_date) DO UPDATE SET
          rating=EXCLUDED.rating, review_count=EXCLUDED.review_count
    """
    for row in rows[3:]:
        if not row or not row[0]:
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


# Loaders registry
LOADERS = {
    'TARGETS': load_targets,
    'UFILL': load_ufill,
    'QOC': load_qoc,
    'SPEED': load_speed,
    'MS': load_ms,
    'HSD': load_hsd,
    'MAK GE Reading': load_mak_ge,
    'SANGAM': load_sangam,
    'GOOGLE RATINGS': load_google_ratings,
}


if __name__ == '__main__':
    main()
