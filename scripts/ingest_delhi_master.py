#!/usr/bin/env python3
"""
Delhi Master Excel Ingestion Script

Parses Delhi_Master.xlsx and ingests into bpcl_portal database.
Supports market share data, trading areas, retail outlets, and synthetic audit scores.

Usage:
    python scripts/ingest_delhi_master.py                    # full run
    python scripts/ingest_delhi_master.py --dry-run          # no DB writes
    python scripts/ingest_delhi_master.py --verbose           # row by row
    python scripts/ingest_delhi_master.py --skip-synthetic   # no audit gen
    python scripts/ingest_delhi_master.py --bpc-only         # skip HPC/IOC
"""

import argparse
import os
import sys
import random
from datetime import date, timedelta, datetime
from pathlib import Path

import openpyxl
import psycopg2
from psycopg2.extras import execute_batch
import requests
from dotenv import load_dotenv

load_dotenv()

random.seed(42)

EXCEL_PATH = os.environ.get('DELHI_MASTER_PATH', '/Users/meynesh/Documents/bpcl-sso/Delhi_Master.xlsx')

MONTHLY_COLUMN_PAIRS = [
    ((9, 'MS0424'), (10, 'MS-0425'), (11, 'HSD-0424'), (12, 'HSD-0425'), '2024-04'),
    ((13, 'MS-0524'), (14, 'MS-0525'), (15, 'HSD-0524'), (16, 'HSD-0525'), '2024-05'),
    ((17, 'MS-0625'), (18, 'MS-0624'), (19, 'HSD-0625'), (20, 'HSD-0624'), '2024-06'),
    ((21, 'MS-0725'), (22, 'MS-0724'), (23, 'HSD-0725'), (24, 'HSD-0724'), '2024-07'),
    ((25, 'MS-0825'), (26, 'MS-0824'), (27, 'HSD-0825'), (28, 'HSD-0824'), '2024-08'),
    ((29, 'MS-0925'), (30, 'MS-0924'), (31, 'HSD-0925'), (32, 'HSD-0924'), '2024-09'),
    ((33, 'MS-1025'), (34, 'MS-1024'), (35, 'HSD-1025'), (36, 'HSD-1024'), '2024-10'),
    ((37, 'MS-1125'), (38, 'MS-1124'), (39, 'HSD-1125'), (40, 'HSD-1124'), '2024-11'),
    ((41, 'MS-1225'), (42, 'MS-1224'), (43, 'HSD-1225'), (44, 'HSD-1224'), '2024-12'),
    ((45, 'MS-0126'), (46, 'MS-0125'), (47, 'HSD-0126'), (48, 'HSD-0125'), '2026-01'),
    ((49, 'MS-0226'), (50, 'MS-0225'), (51, 'HSD-0226'), (52, 'HSD-0225'), '2026-02'),
    ((53, 'MS-0326'), (54, 'MS-0325'), (55, 'HSD-0326'), (56, 'HSD-0325'), '2026-03'),
    ((57, 'MS-0426'), (58, 'MS-0425'), (59, 'HSD-0426'), (60, 'HSD-0425'), '2026-04'),
    ((61, 'MS-0526'), (62, 'MS-0525'), (63, 'HSD-0526'), (64, 'HSD-0525'), '2026-05'),
    ((65, 'MS-0626'), (66, 'MS-0625'), (67, 'HSD-0626'), (68, 'HSD-0625'), '2026-06'),
    ((69, 'MS-0726'), (70, 'MS-0725'), (71, 'HSD-0726'), (72, 'HSD-0725'), '2026-07'),
    ((73, 'MS-0826'), (74, 'MS-0825'), (75, 'HSD-0826'), (76, 'HSD-0825'), '2026-08'),
    ((77, 'MS-0926'), (78, 'MS-0925'), (79, 'HSD-0926'), (80, 'HSD-0925'), '2026-09'),
    ((81, 'MS-1026'), (82, 'MS-1025'), (83, 'HSD-1026'), (84, 'HSD-1025'), '2026-10'),
    ((85, 'MS-1126'), (86, 'MS-1125'), (87, 'HSD-1126'), (88, 'HSD-1125'), '2026-11'),
    ((89, 'MS-1226'), (90, 'MS-1225'), (91, 'HSD-1226'), (92, 'HSD-1225'), '2026-12'),
    ((93, 'MS-0127'), (94, 'MS-0126'), (95, 'HSD-0127'), (96, 'HSD-0126'), '2027-01'),
    ((97, 'MS-0227'), (98, 'MS-0226'), (99, 'HSD-0227'), (100, 'HSD-0226'), '2027-02'),
]

DISTRICT_TERRITORY_MAP = {
    'WEST DELHI': 'DELHI-W',
    'EAST DELHI': 'DELHI-E',
    'SOUTH DELHI': 'DELHI-S',
    'SOUTH EAST DELHI': 'DELHI-S',
    'CENTRAL DELHI': 'DELHI-C',
    'NORTH WEST DELHI': 'DELHI-NW',
    'NORTH EAST DELHI': 'DELHI-NE',
    'SOUTH WEST DELHI': 'DELHI-SW',
    'NEW DELHI': 'DELHI-C',
}

DEFAULT_TERRITORY = 'DELHI-ALL'


class RawRow:
    def __init__(self, outlet_name, cc_number, omc, trading_area_name, district, location, period, ms_vol_kl, hsd_vol_kl):
        self.outlet_name = outlet_name
        self.cc_number = cc_number
        self.omc = omc
        self.trading_area_name = trading_area_name
        self.district = district
        self.location = location
        self.period = period
        self.ms_vol_kl = ms_vol_kl
        self.hsd_vol_kl = hsd_vol_kl


def get_db_connection():
    db_url = os.environ.get('BPCL_DB_URL')
    if not db_url:
        print("ERROR: BPCL_DB_URL not set in environment")
        sys.exit(1)
    return psycopg2.connect(db_url)


def parse_excel(dry_run=False, verbose=False, bpc_only=False):
    """Parse the Delhi Master Excel file."""
    print(f"Parsing Excel: {EXCEL_PATH}")
    
    if not EXCEL_PATH.exists():
        print(f"ERROR: Excel file not found at {EXCEL_PATH}")
        sys.exit(1)
    
    wb = openpyxl.load_workbook(EXCEL_PATH, data_only=True)
    ws = wb['Data']
    
    rows = []
    bpc_count = 0
    hpc_count = 0
    ioc_count = 0
    null_skipped = 0
    div_zero_skipped = 0
    
    for idx, row in enumerate(ws.iter_rows(min_row=2, values_only=True), start=2):
        if not row or not row[1]:
            continue
        
        omc_val = str(row[1]).strip().upper() if row[1] else None
        if omc_val not in ('BPC', 'HPC', 'IOC'):
            continue
        
        if bpc_only and omc_val != 'BPC':
            continue
        
        if omc_val == 'BPC':
            bpc_count += 1
        elif omc_val == 'HPC':
            hpc_count += 1
        elif omc_val == 'IOC':
            ioc_count += 1
        
        cc_number = str(int(row[0])) if row[0] else None
        outlet_name = row[2] if row[2] else f"Unknown-{idx}"
        location = row[3] if row[3] else None
        district = row[4] if row[4] else None
        trading_area_excel = row[5] if row[5] else row[6]
        if not trading_area_excel:
            trading_area_excel = row[6] if row[6] else "Unknown"
        
        trading_area_name = str(trading_area_excel).strip().replace(' ', '_').upper()
        
        for (ly_ms_idx, ly_ms_name), (ty_ms_idx, ty_ms_name), (ly_hsd_idx, ly_hsd_name), (ty_hsd_idx, ty_hsd_name), period_base in MONTHLY_COLUMN_PAIRS:
            if ty_ms_idx >= len(row):
                continue
            
            try:
                ms_val = row[ty_ms_idx] if ty_ms_idx < len(row) else None
                hsd_val = row[ty_hsd_idx] if ty_hsd_idx < len(row) else None
                
                if isinstance(ms_val, str) and ('#DIV/0' in ms_val.upper() or '#DIV' in ms_val.upper()):
                    div_zero_skipped += 1
                    ms_val = None
                elif isinstance(ms_val, str):
                    ms_val = None
                
                if isinstance(hsd_val, str) and ('#DIV/0' in hsd_val.upper() or '#DIV' in hsd_val.upper()):
                    div_zero_skipped += 1
                    hsd_val = None
                elif isinstance(hsd_val, str):
                    hsd_val = None
                
                if ms_val is None and hsd_val is None:
                    null_skipped += 1
                    continue
                
                ms_vol = float(ms_val) if ms_val is not None else 0.0
                hsd_vol = float(hsd_val) if hsd_val is not None else 0.0
                
                if ms_vol == 0 and hsd_vol == 0:
                    null_skipped += 1
                    continue
                
                period_str = period_base + '-01'
                
                raw = RawRow(
                    outlet_name=outlet_name,
                    cc_number=cc_number if omc_val == 'BPC' else None,
                    omc=omc_val,
                    trading_area_name=trading_area_name,
                    district=district,
                    location=location,
                    period=period_str,
                    ms_vol_kl=ms_vol,
                    hsd_vol_kl=hsd_vol
                )
                rows.append(raw)
                
                if verbose:
                    print(f"  Row {idx}: {outlet_name} ({omc_val}) - {period_str}: MS={ms_vol}, HSD={hsd_vol}")
                    
            except (ValueError, TypeError) as e:
                if verbose:
                    print(f"  Skipping row {idx}, col {ty_ms_idx}: {e}")
                continue
    
    print(f"Parsed: {bpc_count} BPC, {hpc_count} HPC, {ioc_count} IOC")
    print(f"  Total rows: {len(rows)}, Null skipped: {null_skipped}, #DIV/0! skipped: {div_zero_skipped}")
    
    return rows, {'bpc': bpc_count, 'hpc': hpc_count, 'ioc': ioc_count, 
                 'null_skipped': null_skipped, 'div_skipped': div_zero_skipped}


def resolve_trading_areas(conn, rows, dry_run=False, verbose=False):
    """Resolve trading areas from Excel names to DB IDs."""
    cursor = conn.cursor()
    
    cursor.execute("SELECT id, name, district FROM trading_areas")
    db_areas = {row[1].upper(): {'id': row[0], 'district': row[2]} for row in cursor.fetchall()}
    
    cursor.execute("SELECT DISTINCT name FROM trading_areas")
    all_db_names = {row[0].upper() for row in cursor.fetchall()}
    
    ta_cache = {}
    exact_matches = 0
    fuzzy_matches = 0
    new_created = 0
    
    unique_ta = set()
    for r in rows:
        unique_ta.add((r.trading_area_name, r.district))
    
    for ta_name, district in unique_ta:
        if ta_name in ta_cache:
            continue
        
        ta_upper = ta_name.upper()
        
        if ta_upper in db_areas:
            ta_cache[ta_name] = db_areas[ta_upper]['id']
            exact_matches += 1
            if verbose:
                print(f"  Exact match: {ta_name} -> ID {ta_cache[ta_name]}")
            continue
        
        matched = False
        for db_name, db_info in db_areas.items():
            if ta_upper in db_name or db_name in ta_upper:
                ta_cache[ta_name] = db_info['id']
                fuzzy_matches += 1
                matched = True
                if verbose:
                    print(f"  Fuzzy match: {ta_name} -> {db_name} (ID {ta_cache[ta_name]})")
                break
        
        if not matched:
            state = 'Delhi'
            if district:
                district_clean = str(district).strip().title()
            else:
                district_clean = 'Delhi'
            
            if not dry_run:
                try:
                    cursor.execute("""
                        INSERT INTO trading_areas (name, district, state)
                        VALUES (%s, %s, %s)
                        RETURNING id
                    """, (ta_name.replace('_', ' ').title(), district_clean, state))
                    
                    new_id = cursor.fetchone()[0]
                    conn.commit()
                    ta_cache[ta_name] = new_id
                    new_created += 1
                    db_areas[ta_upper] = {'id': new_id, 'district': district_clean}
                    
                    if verbose:
                        print(f"  Created new: {ta_name} (district: {district_clean})")
                except Exception as e:
                    conn.rollback()
                    if verbose:
                        print(f"  Skipped new TA {ta_name}: {e}")
                    continue
            else:
                new_created += 1
    
    cursor.close()
    
    print(f"Trading areas: {exact_matches} exact, {fuzzy_matches} fuzzy, {new_created} new")
    
    return ta_cache


def upsert_outlets(conn, rows, ta_cache, dry_run=False, verbose=False):
    """Upsert missing BPC outlets into retail_outlets."""
    cursor = conn.cursor()
    
    bpc_rows = [r for r in rows if r.omc == 'BPC']
    bpc_ccs = set(r.cc_number for r in bpc_rows if r.cc_number)
    
    cursor.execute("SELECT cc_number FROM retail_outlets WHERE outlet_type = 'regular'")
    existing_ccs = {row[0] for row in cursor.fetchall()}
    
    new_outlets = 0
    updated_outlets = 0
    
    for r in bpc_rows:
        if not r.cc_number:
            continue
        
        if r.cc_number in existing_ccs:
            continue
        
        territory = DEFAULT_TERRITORY
        if r.district:
            dist_upper = str(r.district).strip().upper()
            territory = DISTRICT_TERRITORY_MAP.get(dist_upper, DEFAULT_TERRITORY)
        
        ta_id = ta_cache.get(r.trading_area_name)
        
        if dry_run:
            new_outlets += 1
            continue
        
        cursor.execute("""
            INSERT INTO retail_outlets (cc_number, name, location, district, state, territory_code, trading_area_id, outlet_type, is_active)
            VALUES (%s, %s, %s, %s, %s, %s, %s, 'regular', true)
            ON CONFLICT (cc_number) DO UPDATE SET
                name = EXCLUDED.name,
                trading_area_id = EXCLUDED.trading_area_id,
                updated_at = NOW()
            RETURNING cc_number
        """, (r.cc_number, r.outlet_name, r.location, r.district, 'Delhi', territory, ta_id))
        
        result = cursor.fetchone()
        if result:
            new_outlets += 1
    
    conn.commit()
    cursor.close()
    
    print(f"Outlets: {new_outlets} new")
    
    return new_outlets


def upsert_market_share_data(conn, rows, ta_cache, dry_run=False, verbose=False):
    """Upsert market share data for all outlets."""
    cursor = conn.cursor()
    
    data_rows = []
    for r in rows:
        ta_id = ta_cache.get(r.trading_area_name)
        if not ta_id:
            continue
        
        data_rows.append((
            r.outlet_name,
            r.cc_number,
            {'BPC': 'BPCL', 'HPC': 'HPCL', 'IOC': 'IOCL'}.get(r.omc, r.omc),
            ta_id,
            r.period,
            r.ms_vol_kl,
            r.hsd_vol_kl,
            'delhi_master'
        ))
    
    if dry_run:
        print(f"Market share: {len(data_rows)} rows would be inserted/updated")
        return 0, 0
    
    insert_count = 0
    update_count = 0
    
    batch_size = 500
    for i in range(0, len(data_rows), batch_size):
        batch = data_rows[i:i+batch_size]
        
        cursor.execute("SELECT COUNT(*) FROM market_share_data WHERE FALSE")
        
        for row in batch:
            cursor.execute("""
                INSERT INTO market_share_data (outlet_name, cc_number, omc, trading_area_id, period, ms_vol_kl, hsd_vol_kl, source)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (outlet_name, omc, trading_area_id, period) DO UPDATE SET
                    ms_vol_kl = EXCLUDED.ms_vol_kl,
                    hsd_vol_kl = EXCLUDED.hsd_vol_kl,
                    cc_number = EXCLUDED.cc_number
            """, row)
            
            if cursor.statusmessage and 'INSERT' in cursor.statusmessage:
                insert_count += 1
            else:
                update_count += 1
    
    conn.commit()
    cursor.close()
    
    print(f"Market share: {insert_count} inserted, {update_count} updated")
    
    return insert_count, update_count


def compute_trading_area_totals(conn, dry_run=False, verbose=False):
    """Compute and upsert trading_area_totals."""
    cursor = conn.cursor()
    
    cursor.execute("""
        SELECT DISTINCT period FROM market_share_data ORDER BY period
    """)
    periods = [row[0] for row in cursor.fetchall()]
    
    totals_count = 0
    
    for period in periods:
        cursor.execute("""
            SELECT 
                trading_area_id,
                %s as period,
                SUM(ms_vol_kl) as total_ms,
                SUM(hsd_vol_kl) as total_hsd,
                COALESCE(SUM(CASE WHEN omc = 'BPCL' THEN ms_vol_kl ELSE 0 END), 0) as bpcl_ms,
                COALESCE(SUM(CASE WHEN omc = 'HPCL' THEN ms_vol_kl ELSE 0 END), 0) as hpcl_ms,
                COALESCE(SUM(CASE WHEN omc = 'IOCL' THEN ms_vol_kl ELSE 0 END), 0) as iocl_ms,
                COALESCE(SUM(CASE WHEN omc = 'BPCL' THEN hsd_vol_kl ELSE 0 END), 0) as bpcl_hsd,
                COALESCE(SUM(CASE WHEN omc = 'HPCL' THEN hsd_vol_kl ELSE 0 END), 0) as hpcl_hsd,
                COALESCE(SUM(CASE WHEN omc = 'IOCL' THEN hsd_vol_kl ELSE 0 END), 0) as iocl_hsd
            FROM market_share_data
            WHERE period = %s
            GROUP BY trading_area_id
        """, (period, period))
        
        totals = cursor.fetchall()
        
        for t in totals:
            if dry_run:
                continue
            
            cursor.execute("""
                INSERT INTO trading_area_totals (trading_area_id, period, total_ms_kl, total_hsd_kl, 
                    bpcl_ms_kl, hpcl_ms_kl, iocl_ms_kl, bpcl_hsd_kl, hpcl_hsd_kl, iocl_hsd_kl, computed_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, NOW())
                ON CONFLICT (trading_area_id, period) DO UPDATE SET
                    total_ms_kl = EXCLUDED.total_ms_kl,
                    total_hsd_kl = EXCLUDED.total_hsd_kl,
                    bpcl_ms_kl = EXCLUDED.bpcl_ms_kl,
                    hpcl_ms_kl = EXCLUDED.hpcl_ms_kl,
                    iocl_ms_kl = EXCLUDED.iocl_ms_kl,
                    bpcl_hsd_kl = EXCLUDED.bpcl_hsd_kl,
                    hpcl_hsd_kl = EXCLUDED.hpcl_hsd_kl,
                    iocl_hsd_kl = EXCLUDED.iocl_hsd_kl,
                    computed_at = NOW()
            """, t)
        
        totals_count += len(totals)
    
    conn.commit()
    cursor.close()
    
    print(f"Trading area totals: {totals_count} records")
    return totals_count


def generate_synthetic_audit_scores(conn, rows, dry_run=False, verbose=False):
    """Generate synthetic audit scores for BPC outlets."""
    cursor = conn.cursor()
    
    bpc_rows = [r for r in rows if r.omc == 'BPC' and r.cc_number]
    
    cursor.execute("""
        SELECT cc_number, ms_vol_kl 
        FROM market_share_data 
        WHERE omc = 'BPCL' AND period = '2026-03-01'
    """)
    march_volumes = {row[0]: row[1] for row in cursor.fetchall()}
    
    if not march_volumes:
        print("Audit scores: No March 2026 data found, skipping synthetic generation")
        return 0, 0
    
    max_ms = max(v for v in march_volumes.values() if v) if march_volumes else 1
    
    audit_generated = 0
    audit_skipped = 0
    
    for cc, march_ms in march_volumes.items():
        ms_rank_pct = float(march_ms / max_ms) if max_ms > 0 else 0
        
        grade_rand = random.random() + (ms_rank_pct * 0.3)
        if grade_rand > 0.9:
            cleanliness_grade = 'Excellent'
        elif grade_rand > 0.7:
            cleanliness_grade = 'Good'
        elif grade_rand > 0.4:
            cleanliness_grade = 'Average'
        elif grade_rand > 0.2:
            cleanliness_grade = 'Below Average'
        else:
            cleanliness_grade = 'Poor'
        
        audit_status = 'not_audited' if march_ms < 100 else 'audited'
        
        rating = 3.0 + (ms_rank_pct * 1.8) + (random.random() * 0.4 - 0.2)
        rating = max(2.5, min(4.9, rating))
        google_rating = round(rating * 10) / 10
        
        ips = 40 + (ms_rank_pct * 50) + (random.random() * 10 - 5)
        ips = max(15, min(98, ips))
        
        sangam = int(march_ms / 30) + random.randint(0, 8)
        sangam = max(0, min(50, sangam))
        
        if dry_run:
            audit_generated += 1
            continue

        cursor.execute("""
                INSERT INTO dealer_audit_scores (cc_number, period, cleanliness_grade, cleanliness_status, audit_status, google_rating, ips_pct, sangam_count)
                VALUES (%s, '2026-03-01', %s, %s, %s, %s, %s, %s)
                ON CONFLICT (cc_number, period) DO NOTHING
            """, (cc, cleanliness_grade, 'audited', audit_status, google_rating, round(ips), sangam))
        
        if cursor.rowcount > 0:
            audit_generated += 1
        else:
            audit_skipped += 1
    
    conn.commit()
    cursor.close()
    
    print(f"Audit scores: {audit_generated} generated, {audit_skipped} skipped (exists)")
    return audit_generated, audit_skipped


def generate_synthetic_market_share(conn, dry_run=False, verbose=False):
    """Generate synthetic market share data for dealers missing from Delhi Master source."""
    cursor = conn.cursor()

    # Get current competition period
    cursor.execute("""
        SELECT period FROM competition_periods
        WHERE status IN ('active', 'published')
        ORDER BY period DESC LIMIT 1
    """)
    result = cursor.fetchone()
    current_period = result[0] if result else None

    if not current_period:
        print("Synthetic market share: No active competition found")
        return 0, 0

    if verbose:
        print(f"[DEBUG] Current period: {current_period}")

    # Get total dealers in competition
    cursor.execute("""
        SELECT COUNT(DISTINCT cc_number) FROM competition_scores
    """)
    total_dealers = cursor.fetchone()[0]
    if verbose:
        print(f"[DEBUG] Total dealers in competition: {total_dealers}")

    # Find dealers missing market_share_data for the current period
    cursor.execute("""
        SELECT DISTINCT cs.cc_number
        FROM competition_scores cs
        JOIN retail_outlets ro ON cs.cc_number = ro.cc_number
        LEFT JOIN market_share_data m
            ON cs.cc_number = m.cc_number
            AND m.period = %s
        WHERE ro.outlet_type = 'regular' AND m.id IS NULL
    """, (current_period,))
    missing_dealers = [row[0] for row in cursor.fetchall()]

    if verbose:
        print(f"[DEBUG] Found {len(missing_dealers)} missing dealers")
        # Debug: Check regular outlets count
        cursor.execute("""
            SELECT COUNT(DISTINCT cc_number) FROM retail_outlets WHERE outlet_type = 'regular'
        """)
        regular_count = cursor.fetchone()[0]
        print(f"[DEBUG] Total regular outlets: {regular_count}")

        # Debug: Check market_share_data count for period
        cursor.execute("""
            SELECT COUNT(DISTINCT cc_number) FROM market_share_data WHERE period = %s
        """, (current_period,))
        ms_count = cursor.fetchone()[0]
        print(f"[DEBUG] Dealers with market_share_data for {current_period}: {ms_count}")

        # Debug: Check competition dealers with matching retail outlets
        cursor.execute("""
            SELECT COUNT(DISTINCT cs.cc_number)
            FROM competition_scores cs
            JOIN retail_outlets ro ON cs.cc_number = ro.cc_number
            WHERE ro.outlet_type = 'regular'
        """)
        competition_regular = cursor.fetchone()[0]
        print(f"[DEBUG] Competition dealers with regular retail outlet: {competition_regular}")

    if not missing_dealers:
        print(f"Synthetic market share: No missing dealers found")
        return 0, 0

    print(f"Synthetic market share: Found {len(missing_dealers)} dealers missing from Delhi Master")

    # Calculate previous year period
    current_year = current_period.year
    current_month = current_period.month
    prev_year_date = datetime(current_year - 1, current_month, 1).date()

    synthetic_rows = []

    for cc_number in missing_dealers:
        # Get dealer's trading area
        cursor.execute("""
            SELECT ro.trading_area_id, ro.name, ta.district
            FROM retail_outlets ro
            LEFT JOIN trading_areas ta ON ro.trading_area_id = ta.id
            WHERE ro.cc_number = %s
        """, (cc_number,))
        dealer_result = cursor.fetchone()
        if not dealer_result or not dealer_result[0]:
            continue

        trading_area_id, outlet_name, district = dealer_result

        # Get dealer's performance volumes for current period (MS and HSD)
        cursor.execute("""
            SELECT SUM(CASE WHEN product_id = 1 THEN volume_kl ELSE 0 END),
                   SUM(CASE WHEN product_id = 2 THEN volume_kl ELSE 0 END)
            FROM performance_records
            WHERE cc_number = %s AND period = %s
        """, (cc_number, current_period))
        current_perf = cursor.fetchone()

        # Use performance volumes as proxy for market share volumes
        current_ms = float(current_perf[0]) if current_perf and current_perf[0] else 0
        current_hsd = float(current_perf[1]) if current_perf and current_perf[1] else 0

        if current_ms == 0 and current_hsd == 0:
            continue

        # Create synthetic entry
        synthetic_rows.append((
            outlet_name,
            cc_number,
            'BPCL',  # Assume BPCL for synthetic data
            trading_area_id,
            current_period,
            current_ms,
            current_hsd,
            'synthetic'
        ))

        if verbose:
            print(f"  {outlet_name} ({cc_number}): MS={current_ms}, HSD={current_hsd} (from performance)")

    if not synthetic_rows:
        print(f"Synthetic market share: No performance data found for missing dealers")
        return 0, 0

    if dry_run:
        print(f"Synthetic market share: {len(synthetic_rows)} rows would be inserted")
        return len(synthetic_rows), 0

    insert_count = 0
    for row in synthetic_rows:
        cursor.execute("""
            INSERT INTO market_share_data (outlet_name, cc_number, omc, trading_area_id, period, ms_vol_kl, hsd_vol_kl, source)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (outlet_name, omc, trading_area_id, period) DO UPDATE SET
                ms_vol_kl = EXCLUDED.ms_vol_kl,
                hsd_vol_kl = EXCLUDED.hsd_vol_kl,
                cc_number = EXCLUDED.cc_number,
                source = EXCLUDED.source
        """, row)
        if cursor.rowcount > 0:
            insert_count += 1

    conn.commit()

    # Recompute trading_area_totals for affected periods
    cursor.execute("""
        SELECT DISTINCT trading_area_id FROM market_share_data
        WHERE period = %s
    """, (current_period,))
    affected_tas = [row[0] for row in cursor.fetchall()]

    for ta_id in affected_tas:
        cursor.execute("""
            SELECT
                trading_area_id,
                SUM(ms_vol_kl) as total_ms,
                SUM(hsd_vol_kl) as total_hsd,
                COALESCE(SUM(CASE WHEN omc = 'BPCL' THEN ms_vol_kl ELSE 0 END), 0) as bpcl_ms,
                COALESCE(SUM(CASE WHEN omc = 'HPCL' THEN ms_vol_kl ELSE 0 END), 0) as hpcl_ms,
                COALESCE(SUM(CASE WHEN omc = 'IOCL' THEN ms_vol_kl ELSE 0 END), 0) as iocl_ms,
                COALESCE(SUM(CASE WHEN omc = 'BPCL' THEN hsd_vol_kl ELSE 0 END), 0) as bpcl_hsd,
                COALESCE(SUM(CASE WHEN omc = 'HPCL' THEN hsd_vol_kl ELSE 0 END), 0) as hpcl_hsd,
                COALESCE(SUM(CASE WHEN omc = 'IOCL' THEN hsd_vol_kl ELSE 0 END), 0) as iocl_hsd
            FROM market_share_data
            WHERE trading_area_id = %s AND period = %s
            GROUP BY trading_area_id
        """, (ta_id, current_period))

        total_result = cursor.fetchone()
        if total_result:
            cursor.execute("""
                INSERT INTO trading_area_totals (trading_area_id, period, total_ms_kl, total_hsd_kl,
                    bpcl_ms_kl, hpcl_ms_kl, iocl_ms_kl, bpcl_hsd_kl, hpcl_hsd_kl, iocl_hsd_kl, computed_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, NOW())
                ON CONFLICT (trading_area_id, period) DO UPDATE SET
                    total_ms_kl = EXCLUDED.total_ms_kl,
                    total_hsd_kl = EXCLUDED.total_hsd_kl,
                    bpcl_ms_kl = EXCLUDED.bpcl_ms_kl,
                    hpcl_ms_kl = EXCLUDED.hpcl_ms_kl,
                    iocl_ms_kl = EXCLUDED.iocl_ms_kl,
                    bpcl_hsd_kl = EXCLUDED.bpcl_hsd_kl,
                    hpcl_hsd_kl = EXCLUDED.hpcl_hsd_kl,
                    iocl_hsd_kl = EXCLUDED.iocl_hsd_kl,
                    computed_at = NOW()
            """, (ta_id, current_period, total_result[1], total_result[2], total_result[3], total_result[4], total_result[5], total_result[6], total_result[7], total_result[8], total_result[9]))

    conn.commit()
    cursor.close()

    print(f"Synthetic market share: {insert_count} records inserted, {len(affected_tas)} trading areas recomputed")
    return insert_count, len(affected_tas)


def trigger_recompute():
    """Trigger competition recompute via API."""
    try:
        login_resp = requests.post(
            'http://localhost:8080/api/v1/auth/login',
            json={'employee_id': 'EMP10001', 'password': 'Bpcl@2026'},
            timeout=10
        )
        
        if login_resp.status_code != 200:
            print(f"Recompute: Login failed ({login_resp.status_code})")
            return False
        
        token = login_resp.json().get('token')
        if not token:
            print("Recompute: No token in login response")
            return False
        
        competition_id = 'c0000000-0001-0001-0001-000000000001'
        
        recompute_resp = requests.post(
            f'http://localhost:8080/api/v1/competition/{competition_id}/recompute?confirm=true',
            headers={'Authorization': f'Bearer {token}'},
            timeout=30
        )
        
        if recompute_resp.status_code == 200:
            print("Recompute: ✅ triggered successfully")
            return True
        else:
            print(f"Recompute: Failed ({recompute_resp.status_code}) - {recompute_resp.text}")
            return False
            
    except requests.exceptions.RequestException as e:
        print(f"Recompute: API not reachable - {e}")
        return False


def main():
    parser = argparse.ArgumentParser(description='Ingest Delhi Master Excel data')
    parser.add_argument('--dry-run', action='store_true', help='No DB writes')
    parser.add_argument('--verbose', action='store_true', help='Row by row output')
    parser.add_argument('--skip-synthetic', action='store_true', help='Skip audit score generation')
    parser.add_argument('--bpc-only', action='store_true', help='Only process BPC outlets')
    parser.add_argument('--excel-path', type=str, default=None, help='Path to Delhi_Master.xlsx')
    parser.add_argument('--synthetic-only', action='store_true', help='Only generate synthetic market share data')

    args = parser.parse_args()
    
    if args.excel_path:
        global EXCEL_PATH
        EXCEL_PATH = args.excel_path
    EXCEL_PATH = Path(EXCEL_PATH)
    
    print("=" * 50)
    print("BPCL Delhi Master Ingestion")
    print("=" * 50)

    if args.dry_run:
        print("MODE: DRY RUN (no DB writes)")

    if args.synthetic_only:
        print("MODE: SYNTHETIC ONLY (no Excel parsing)")
        conn = get_db_connection()
        try:
            synthetic_count, affected_tas = generate_synthetic_market_share(conn, dry_run=args.dry_run, verbose=args.verbose)
        finally:
            conn.close()

        recompute_success = False
        if not args.dry_run:
            recompute_success = trigger_recompute()

        print("\n" + "=" * 50)
        print("Synthetic Generation Report")
        print("=" * 50)
        print(f"  Synthetic market share rows:   {synthetic_count}")
        print(f"  Trading areas recomputed:      {affected_tas}")
        print(f"  Recompute triggered:           {'✅' if recompute_success else '❌ (API not reachable)'}")
        print("=" * 50)
    else:
        rows, counts = parse_excel(dry_run=args.dry_run, verbose=args.verbose, bpc_only=args.bpc_only)

        conn = get_db_connection()

        try:
            ta_cache = resolve_trading_areas(conn, rows, dry_run=args.dry_run, verbose=args.verbose)

            new_outlets = upsert_outlets(conn, rows, ta_cache, dry_run=args.dry_run, verbose=args.verbose)

            insert_count, update_count = upsert_market_share_data(conn, rows, ta_cache, dry_run=args.dry_run, verbose=args.verbose)

            totals_count = compute_trading_area_totals(conn, dry_run=args.dry_run, verbose=args.verbose)

            audit_generated = 0
            audit_skipped = 0
            if not args.skip_synthetic:
                audit_generated, audit_skipped = generate_synthetic_audit_scores(conn, rows, dry_run=args.dry_run, verbose=args.verbose)

        finally:
            conn.close()

        recompute_success = False
        if not args.dry_run:
            recompute_success = trigger_recompute()

        print("\n" + "=" * 50)
        print("Ingestion Report")
        print("=" * 50)
        print(f"  BPC outlets processed:        {counts['bpc']}")
        print(f"  HPC outlets processed:         {counts['hpc']}")
        print(f"  IOC outlets processed:         {counts['ioc']}")
        print(f"  Total outlets:                 {counts['bpc'] + counts['hpc'] + counts['ioc']}")
        print()
        print(f"  New outlets added:             {new_outlets}")
        print()
        print(f"  Market share rows:             {insert_count + update_count}")
        print(f"  Trading area totals records:   {totals_count}")
        print()
        if not args.skip_synthetic:
            print(f"  Audit scores generated:        {audit_generated}")
            print(f"  Audit scores skipped:         {audit_skipped}")
        print()
        print(f"  Null volume cells skipped:     {counts['null_skipped']}")
        print(f"  #DIV/0! cells skipped:         {counts['div_skipped']}")
        print()
        print(f"  Recompute triggered:           {'✅' if recompute_success else '❌ (API not reachable)'}")
        print("=" * 50)


if __name__ == '__main__':
    main()