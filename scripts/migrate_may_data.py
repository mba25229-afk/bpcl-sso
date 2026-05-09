#!/usr/bin/env python3
"""
Migrate MAY DATA.xlsx into the Crystal DB tables.

Sheets handled:
  TARGETS        → cr_dealers (upsert) + cr_monthly_targets
  MS             → cr_daily_ms
  HSD            → cr_daily_hsd
  UFILL          → cr_daily_ufill
  QOC            → cr_daily_qoc
  SPEED          → cr_daily_speed  (Speed + Speed100 combined per dealer)
  MAK GE Reading → cr_mak_ge_readings  (weekly snapshots)
  SANGAM         → cr_sangam_data
  GOOGLE RATINGS → cr_google_ratings  (three snapshots: May-01, May-15, May-25)
"""

import sys
import os
import re
import datetime
import openpyxl
import psycopg2
import psycopg2.extras

XLSX = "/Users/meynesh/Documents/bpcl-sso/docs/crystal /MAY DATA.xlsx"
DB_URL = "postgres://bpcl:bpcl@localhost:5433/bpcl_portal"
MONTH_YEAR = datetime.date(2026, 5, 1)


# ── helpers ───────────────────────────────────────────────────────────────────

def clean_cc(val):
    """Return integer cc_code string or None."""
    if val is None:
        return None
    s = str(val).replace(".0", "").strip()
    return s if s.isdigit() else None


def to_float(val):
    """Evaluate simple Excel formula strings like '=1.37+0.2', else cast."""
    if val is None:
        return None
    if isinstance(val, (int, float)):
        return float(val)
    s = str(val).strip()
    if not s:
        return None
    # handle simple addition formulas
    m = re.match(r'^=?([\d.]+)\+([\d.]+)$', s.replace('\n','').replace(' ',''))
    if m:
        return float(m.group(1)) + float(m.group(2))
    try:
        return float(s.lstrip('='))
    except ValueError:
        return None


def to_int(val):
    f = to_float(val)
    return int(f) if f is not None else None


def load_wb():
    return openpyxl.load_workbook(XLSX, data_only=True)


# ── sheet parsers ─────────────────────────────────────────────────────────────

def parse_targets(wb):
    ws = wb['TARGETS']
    rows = list(ws.iter_rows(values_only=True))
    # row 0 = title, row 1 = headers, rows 2+ = data
    dealers = []
    targets = []
    for row in rows[2:]:
        cc = clean_cc(row[1])
        if not cc:
            continue
        ro_name = str(row[0]).strip() if row[0] else ""
        dealers.append((cc, ro_name))
        targets.append({
            'cc_code':        cc,
            'ufill_target':   to_int(row[2]),
            'qoc_target':     to_int(row[3]),
            'speed_kl':       to_float(row[4]),
            'ms_kl':          to_float(row[5]),
            'hsd_kl':         to_float(row[6]),
            'ms_ly':          to_float(row[7]),
            'hsd_ly':         to_float(row[8]),
            'dsw_available':  str(row[9]).strip().upper() == 'YES' if row[9] else False,
            'nitrogen':       str(row[10]).strip().upper() == 'YES' if row[10] else False,
            'mak_ge_target':  to_float(row[12]),
            'darpan_target':  to_int(row[13]),
            'coolant_lube_kl':to_float(row[14]),
        })
    return dealers, targets


def parse_daily_standard(wb, sheet_name):
    """Sheets with layout: RO NAME | CC Code | May-01 ... May-31 | MTD"""
    ws = wb[sheet_name]
    rows = list(ws.iter_rows(values_only=True))
    # row 1 = headers; cols 2..32 = dates May-01..May-31
    header = rows[1]
    dates = []
    for i, h in enumerate(header[2:], start=2):
        if isinstance(h, datetime.datetime):
            dates.append((i, h.date()))
        elif isinstance(h, datetime.date):
            dates.append((i, h))

    records = []  # (cc_code, date, value)
    for row in rows[2:]:
        cc = clean_cc(row[1])
        if not cc:
            continue
        for col_idx, d in dates:
            v = to_float(row[col_idx])
            if v is not None and v > 0:
                records.append((cc, d, v))
    return records


def parse_speed(wb):
    """SPEED sheet: RO | CC | Product | May-01..May-31 | MTD
       Has only one row per dealer (Speed; no separate Speed100 row here).
       Sum by cc_code across dates."""
    ws = wb['SPEED']
    rows = list(ws.iter_rows(values_only=True))
    header = rows[1]
    dates = []
    for i, h in enumerate(header[3:], start=3):
        if isinstance(h, (datetime.datetime, datetime.date)):
            d = h.date() if isinstance(h, datetime.datetime) else h
            dates.append((i, d))

    from collections import defaultdict
    daily = defaultdict(float)  # (cc, date) -> kl

    for row in rows[2:]:
        cc = clean_cc(row[1])
        if not cc:
            continue
        for col_idx, d in dates:
            v = to_float(row[col_idx])
            if v is not None:
                daily[(cc, d)] += v

    return [(cc, d, v) for (cc, d), v in daily.items() if v > 0]


def parse_mak_ge(wb):
    """MAK GE Reading: RO | CC | reading1 | reading2 | reading3 | reading4 | reading5
       Dates: May-01, May-08, May-15, May-22, May-31"""
    ws = wb['MAK GE Reading']
    rows = list(ws.iter_rows(values_only=True))
    snapshot_dates = [
        datetime.date(2026, 5, 1),
        datetime.date(2026, 5, 8),
        datetime.date(2026, 5, 15),
        datetime.date(2026, 5, 22),
        datetime.date(2026, 5, 31),
    ]
    records = []
    for row in rows[1:]:
        cc = clean_cc(row[1])
        if not cc:
            continue
        for i, d in enumerate(snapshot_dates):
            v = to_float(row[2 + i])
            if v is not None and v > 0:
                records.append((cc, d, v))
    return records


def parse_sangam(wb):
    """SANGAM: RO | CC | Certificates"""
    ws = wb['SANGAM']
    rows = list(ws.iter_rows(values_only=True))
    records = []
    for row in rows[2:]:
        cc = clean_cc(row[1])
        if not cc:
            continue
        cert = to_int(row[2])
        records.append((cc, cert or 0))
    return records


def parse_google(wb):
    """GOOGLE RATINGS: CC | RO | May-01 rating | reviews | _ | May-15 rating | reviews | _ | May-25 rating | reviews"""
    ws = wb['GOOGLE RATINGS']
    rows = list(ws.iter_rows(values_only=True))
    snapshot_dates = [
        datetime.date(2026, 5, 1),
        datetime.date(2026, 5, 15),
        datetime.date(2026, 5, 25),
    ]
    # cols per snapshot: (rating_col, review_col)
    snapshot_cols = [(2, 3), (5, 6), (8, 9)]

    records = []
    for row in rows[3:]:
        cc = clean_cc(row[0])
        if not cc:
            continue
        for (rc, vc), d in zip(snapshot_cols, snapshot_dates):
            rating = to_float(row[rc])
            reviews = to_int(row[vc])
            if rating is not None:
                records.append((cc, d, rating, reviews or 0))
    return records


# ── DB writers ────────────────────────────────────────────────────────────────

def upsert_dealers(cur, dealers):
    psycopg2.extras.execute_batch(cur, """
        INSERT INTO cr_dealers (cc_code, ro_name, area, is_active)
        VALUES (%s, %s, 'Central Delhi', true)
        ON CONFLICT (cc_code) DO UPDATE SET
            ro_name    = EXCLUDED.ro_name,
            updated_at = NOW()
    """, dealers, page_size=100)
    print(f"  cr_dealers: upserted {len(dealers)} rows")


def upsert_targets(cur, targets):
    psycopg2.extras.execute_batch(cur, """
        INSERT INTO cr_monthly_targets (
            cc_code, month_year,
            ufill_target, qoc_target, speed_kl, ms_kl, hsd_kl,
            ms_ly, hsd_ly, dsw_available, nitrogen,
            mak_ge_target, darpan_target, coolant_lube_kl
        ) VALUES (
            %(cc_code)s, '2026-05-01',
            %(ufill_target)s, %(qoc_target)s, %(speed_kl)s, %(ms_kl)s, %(hsd_kl)s,
            %(ms_ly)s, %(hsd_ly)s, %(dsw_available)s, %(nitrogen)s,
            %(mak_ge_target)s, %(darpan_target)s, %(coolant_lube_kl)s
        )
        ON CONFLICT (cc_code, month_year) DO UPDATE SET
            ufill_target    = EXCLUDED.ufill_target,
            qoc_target      = EXCLUDED.qoc_target,
            speed_kl        = EXCLUDED.speed_kl,
            ms_kl           = EXCLUDED.ms_kl,
            hsd_kl          = EXCLUDED.hsd_kl,
            ms_ly           = EXCLUDED.ms_ly,
            hsd_ly          = EXCLUDED.hsd_ly,
            dsw_available   = EXCLUDED.dsw_available,
            nitrogen        = EXCLUDED.nitrogen,
            mak_ge_target   = EXCLUDED.mak_ge_target,
            darpan_target   = EXCLUDED.darpan_target,
            coolant_lube_kl = EXCLUDED.coolant_lube_kl
    """, targets, page_size=100)
    print(f"  cr_monthly_targets: upserted {len(targets)} rows")


def upsert_daily(cur, table, col, records):
    psycopg2.extras.execute_batch(cur, f"""
        INSERT INTO {table} (cc_code, txn_date, {col})
        VALUES (%s, %s, %s)
        ON CONFLICT (cc_code, txn_date) DO UPDATE SET
            {col} = EXCLUDED.{col}
    """, records, page_size=500)
    print(f"  {table}: upserted {len(records)} rows")


def upsert_mak_ge(cur, records):
    psycopg2.extras.execute_batch(cur, """
        INSERT INTO cr_mak_ge_readings (cc_code, reading_date, meter_reading)
        VALUES (%s, %s, %s)
        ON CONFLICT (cc_code, reading_date) DO UPDATE SET
            meter_reading = EXCLUDED.meter_reading
    """, records, page_size=200)
    print(f"  cr_mak_ge_readings: upserted {len(records)} rows")


def upsert_sangam(cur, records):
    psycopg2.extras.execute_batch(cur, """
        INSERT INTO cr_sangam_data (cc_code, month_year, cert_count)
        VALUES (%s, '2026-05-01', %s)
        ON CONFLICT (cc_code, month_year) DO UPDATE SET
            cert_count = EXCLUDED.cert_count
    """, records, page_size=100)
    print(f"  cr_sangam_data: upserted {len(records)} rows")


def upsert_google(cur, records):
    psycopg2.extras.execute_batch(cur, """
        INSERT INTO cr_google_ratings (cc_code, snapshot_date, rating, review_count)
        VALUES (%s, %s, %s, %s)
        ON CONFLICT (cc_code, snapshot_date) DO UPDATE SET
            rating       = EXCLUDED.rating,
            review_count = EXCLUDED.review_count
    """, records, page_size=200)
    print(f"  cr_google_ratings: upserted {len(records)} rows")


def ensure_competition_period(cur):
    cur.execute("""
        INSERT INTO cr_competition_periods (name, month_year, total_slots, is_active)
        VALUES ('Boost and Win May 2026', '2026-05-01', 40, true)
        ON CONFLICT (month_year) DO UPDATE SET
            is_active = true,
            name = EXCLUDED.name
    """)
    print("  cr_competition_periods: May 2026 period active")


# ── main ──────────────────────────────────────────────────────────────────────

def main():
    print("Loading workbook…")
    wb = load_wb()

    print("\nParsing sheets…")
    dealers, targets    = parse_targets(wb)
    ms_records          = parse_daily_standard(wb, 'MS')
    hsd_records         = parse_daily_standard(wb, 'HSD')
    ufill_records       = parse_daily_standard(wb, 'UFILL')
    qoc_records         = parse_daily_standard(wb, 'QOC')
    speed_records       = parse_speed(wb)
    mak_ge_records      = parse_mak_ge(wb)
    sangam_records      = parse_sangam(wb)
    google_records      = parse_google(wb)

    print(f"  dealers={len(dealers)}, targets={len(targets)}")
    print(f"  MS days={len(ms_records)}, HSD={len(hsd_records)}, UFill={len(ufill_records)}")
    print(f"  QOC={len(qoc_records)}, Speed={len(speed_records)}")
    print(f"  MAK GE={len(mak_ge_records)}, Sangam={len(sangam_records)}, Google={len(google_records)}")

    print("\nConnecting to DB…")
    conn = psycopg2.connect(DB_URL)
    conn.autocommit = False
    cur = conn.cursor()

    try:
        print("\nWriting to DB…")
        ensure_competition_period(cur)
        upsert_dealers(cur, dealers)
        upsert_targets(cur, targets)
        upsert_daily(cur, 'cr_daily_ms',    'kl',    ms_records)
        upsert_daily(cur, 'cr_daily_hsd',   'kl',    hsd_records)
        upsert_daily(cur, 'cr_daily_ufill', 'count', ufill_records)
        upsert_daily(cur, 'cr_daily_qoc',   'count', qoc_records)
        upsert_daily(cur, 'cr_daily_speed', 'kl',    speed_records)
        upsert_mak_ge(cur, mak_ge_records)
        upsert_sangam(cur, sangam_records)
        # Only insert google ratings for dealers already in cr_dealers
        known_cc = {cc for cc, _ in dealers}
        google_filtered = [r for r in google_records if r[0] in known_cc]
        skipped_google = len(google_records) - len(google_filtered)
        if skipped_google:
            print(f"  (skipped {skipped_google} google rows — cc_codes not in TARGETS)")
        upsert_google(cur, google_filtered)

        conn.commit()
        print("\n✓ All data committed successfully.")
    except Exception as e:
        conn.rollback()
        print(f"\n✗ Error — rolled back: {e}")
        raise
    finally:
        cur.close()
        conn.close()


if __name__ == '__main__':
    main()
