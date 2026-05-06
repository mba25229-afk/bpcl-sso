-- =============================================================================
-- BPCL SSO Portal — Development Seed Data
-- Period coverage: April 2024 – March 2026
-- 39 BPCL Delhi dealers + supporting OMC data
-- =============================================================================

-- -----------------------------------------------------------------------------
-- 1. USERS
-- -----------------------------------------------------------------------------
-- Password for all test users: "Bpcl@2026" (bcrypt hash below)
-- Generate fresh: htpasswd -bnBC 10 "" "Bpcl@2026" | tr -d ':\n'

INSERT INTO users (id, employee_id, email, name, password_hash, role, territory_code) VALUES
('a1b2c3d4-0001-0001-0001-000000000001', 'EMP10001', 'admin@bpcl.in',
 'System Admin', '$2a$10$q1ri/Y5FIDAdsOc7eQncIuLzDBrpi9QCvtvbHUsuu2qFElVMH3fUS', 'admin', NULL),
('a1b2c3d4-0001-0001-0001-000000000002', 'EMP10002', 'tm.delhi.west@bpcl.in',
 'Rajesh Kumar', '$2a$10$q1ri/Y5FIDAdsOc7eQncIuLzDBrpi9QCvtvbHUsuu2qFElVMH3fUS', 'territory_manager', 'DELHI-W'),
('a1b2c3d4-0001-0001-0001-000000000003', 'EMP10003', 'tm.delhi.east@bpcl.in',
 'Priya Sharma', '$2a$10$q1ri/Y5FIDAdsOc7eQncIuLzDBrpi9QCvtvbHUsuu2qFElVMH3fUS', 'territory_manager', 'DELHI-E'),
('a1b2c3d4-0001-0001-0001-000000000004', 'EMP10004', 'tm.delhi.south@bpcl.in',
 'Amit Verma', '$2a$10$q1ri/Y5FIDAdsOc7eQncIuLzDBrpi9QCvtvbHUsuu2qFElVMH3fUS', 'territory_manager', 'DELHI-S'),
('a1b2c3d4-0001-0001-0001-000000000005', 'EMP10005', 'ro.112847@bpcl.in',
 'Suresh Sethi', '$2a$10$q1ri/Y5FIDAdsOc7eQncIuLzDBrpi9QCvtvbHUsuu2qFElVMH3fUS', 'ro_manager', 'DELHI-W'),
('a1b2c3d4-0001-0001-0001-000000000006', 'EMP10006', 'ro.108923@bpcl.in',
 'Vikram Singh', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LPTIuiRXlWS', 'ro_manager', 'DELHI-S'),
('a1b2c3d4-0001-0001-0001-000000000007', 'EMP10007', 'ro.115634@bpcl.in',
 'Deepak Gupta', '$2a$10$q1ri/Y5FIDAdsOc7eQncIuLzDBrpi9QCvtvbHUsuu2qFElVMH3fUS', 'ro_manager', 'DELHI-C'),
('a1b2c3d4-0001-0001-0001-000000000008', 'EMP10008', 'ro.109251@bpcl.in',
 'Harish Garg', '$2a$10$q1ri/Y5FIDAdsOc7eQncIuLzDBrpi9QCvtvbHUsuu2qFElVMH3fUS', 'ro_manager', 'DELHI-NW'),
('a1b2c3d4-0001-0001-0001-000000000009', 'EMP10009', 'ro.115012@bpcl.in',
 'Mohan Mahadev', '$2a$10$q1ri/Y5FIDAdsOc7eQncIuLzDBrpi9QCvtvbHUsuu2qFElVMH3fUS', 'ro_manager', 'DELHI-C');

-- -----------------------------------------------------------------------------
-- 2. TRADING AREAS
-- -----------------------------------------------------------------------------
INSERT INTO trading_areas (id, name, district, state) VALUES
(1,  'Kirti Nagar',      'West Delhi',      'Delhi'),
(2,  'Rajouri Garden',   'West Delhi',      'Delhi'),
(3,  'Tilak Nagar',      'West Delhi',      'Delhi'),
(4,  'Janakpuri',        'West Delhi',      'Delhi'),
(5,  'Saket',            'South Delhi',     'Delhi'),
(6,  'Lajpat Nagar',     'South Delhi',     'Delhi'),
(7,  'Hauz Khas',        'South Delhi',     'Delhi'),
(8,  'Greater Kailash',  'South Delhi',     'Delhi'),
(9,  'Rohini',           'North West Delhi','Delhi'),
(10, 'Pitampura',        'North West Delhi','Delhi'),
(11, 'Shalimar Bagh',    'North West Delhi','Delhi'),
(12, 'Preet Vihar',      'East Delhi',      'Delhi'),
(13, 'Mayur Vihar',      'East Delhi',      'Delhi'),
(14, 'Laxmi Nagar',      'East Delhi',      'Delhi'),
(15, 'Dwarka',           'South West Delhi','Delhi'),
(16, 'Uttam Nagar',      'South West Delhi','Delhi'),
(17, 'Palam',            'South West Delhi','Delhi'),
(18, 'Connaught Place',  'Central Delhi',   'Delhi'),
(19, 'Karol Bagh',       'Central Delhi',   'Delhi'),
(20, 'Paharganj',        'Central Delhi',   'Delhi'),
(21, 'Dilshad Garden',   'North East Delhi','Delhi'),
(22, 'Seelampur',        'North East Delhi','Delhi'),
(23, 'Chanakyapuri',     'New Delhi',       'Delhi'),
(24, 'RK Puram',         'New Delhi',       'Delhi'),
(25, 'Shahdara North',   'Shahdara',        'Delhi'),
(26, 'Shahdara South',   'Shahdara',        'Delhi'),
(27, 'Dwarka Sector 10', 'South West Delhi','Delhi');

-- -----------------------------------------------------------------------------
-- 3. RETAIL OUTLETS (39 BPCL Delhi dealers)
-- -----------------------------------------------------------------------------
INSERT INTO retail_outlets (cc_number, name, location, district, territory_code, trading_area_id, outlet_type, rank, ro_manager_id) VALUES
('112847', 'M.L. SETHI SERVICE STATION',    'Kirti Nagar Marg, West Delhi 110015',         'West Delhi',      'DELHI-W',  1,  'regular', 'A+', 'a1b2c3d4-0001-0001-0001-000000000005'),
('108923', 'SAI FILLING STATION',            '14 Press Road, South Delhi 110006',           'South Delhi',     'DELHI-S',  5,  'regular', 'A+', 'a1b2c3d4-0001-0001-0001-000000000006'),
('115634', 'AUTO CARE',                      'Connaught Circus, New Delhi 110001',          'New Delhi',       'DELHI-C',  18, 'regular', 'A',  'a1b2c3d4-0001-0001-0001-000000000007'),
('109251', 'GARG ROAD LINES',                'Rohini Sector 9, North West Delhi 110085',   'North West Delhi','DELHI-NW', 9,  'regular', 'A',  'a1b2c3d4-0001-0001-0001-000000000008'),
('113782', 'SAHAS FILLING STATION',          'Dwarka Sector 6, South West Delhi 110075',   'South West Delhi','DELHI-SW', 15, 'regular', 'A',  NULL),
('107634', 'VAIBHAV FILLING STATION',        'Mayur Vihar Phase 1, East Delhi 110091',     'East Delhi',      'DELHI-E',  13, 'regular', 'A',  NULL),
('114521', 'SHANKAR FILLING STATION',        'Karol Bagh, Central Delhi 110005',           'Central Delhi',   'DELHI-C',  19, 'regular', 'A',  NULL),
('116890', 'SAKSHAM MOTORS',                 'Dilshad Garden, North East Delhi 110095',    'North East Delhi','DELHI-NE', 21, 'regular', 'A',  NULL),
('111234', 'SANJEEV FILLING STATION',        'Rajouri Garden, West Delhi 110027',          'West Delhi',      'DELHI-W',  2,  'regular', 'A',  NULL),
('108765', 'LINK ROAD PETROL F.STN.',        'Pitampura, North West Delhi 110034',         'North West Delhi','DELHI-NW', 10, 'regular', 'B',  NULL),
('112345', 'GUPTA SERVICE STATION',          'Lajpat Nagar II, South Delhi 110024',        'South Delhi',     'DELHI-S',  6,  'regular', 'B',  NULL),
('119876', 'NATIONAL FILLING STATION',       'Preet Vihar, East Delhi 110092',             'East Delhi',      'DELHI-E',  12, 'regular', 'B',  NULL),
('113456', 'DELHI PETROLEUM',                'Shahdara Main Road, Shahdara 110032',        'Shahdara',        'DELHI-E',  25, 'regular', 'B',  NULL),
('117234', 'SUNRISE FUEL STATION',           'RK Puram Sector 1, New Delhi 110022',        'New Delhi',       'DELHI-C',  24, 'regular', 'B',  NULL),
('115789', 'KRISHNA PETROLEUM',              'Paharganj, Central Delhi 110055',            'Central Delhi',   'DELHI-C',  20, 'regular', 'B',  NULL),
('118901', 'METRO FILLING STATION',          'Shalimar Bagh, North West Delhi 110088',     'North West Delhi','DELHI-NW', 11, 'regular', 'B',  NULL),
('110234', 'SHARMA PETROL PUMP',             'Tilak Nagar, West Delhi 110018',             'West Delhi',      'DELHI-W',  3,  'regular', 'B',  NULL),
('116123', 'BALAJI SERVICE STATION',         'Uttam Nagar, South West Delhi 110059',       'South West Delhi','DELHI-SW', 16, 'regular', 'B',  NULL),
('119012', 'CAPITAL FUELS',                  'Chanakyapuri, New Delhi 110021',             'New Delhi',       'DELHI-C',  23, 'regular', 'B',  NULL),
('113234', 'ANAND FILLING STATION',          'Laxmi Nagar, East Delhi 110092',             'East Delhi',      'DELHI-E',  14, 'regular', 'B',  NULL),
('117890', 'PUNJAB PETROLEUM',               'Rohini Sector 3, North West Delhi 110085',   'North West Delhi','DELHI-NW', 9,  'regular', 'B',  NULL),
('114678', 'RAJDHANI FUELS',                 'Karol Bagh Extension, Central Delhi 110005', 'Central Delhi',   'DELHI-C',  19, 'regular', 'B',  NULL),
('112789', 'SHIV SHAKTI FILLING STN.',       'Hauz Khas, South Delhi 110016',              'South Delhi',     'DELHI-S',  7,  'regular', 'C',  NULL),
('118345', 'OLYMPIC FILLING STATION',        'Shahdara South, Shahdara 110032',            'Shahdara',        'DELHI-E',  26, 'regular', 'C',  NULL),
('110567', 'LOTUS PETROLEUM',                'Seelampur, North East Delhi 110053',         'North East Delhi','DELHI-NE', 22, 'regular', 'C',  NULL),
('115901', 'HARI OM FUELS',                  'Janakpuri C Block, West Delhi 110058',       'West Delhi',      'DELHI-W',  4,  'regular', 'C',  NULL),
('119456', 'STAR PETROL PUMP',               'Palam Colony, South West Delhi 110045',      'South West Delhi','DELHI-SW', 17, 'regular', 'C',  NULL),
('113890', 'NEW INDIA FILLING STN.',          'Mayur Vihar Phase 3, East Delhi 110096',    'East Delhi',      'DELHI-E',  13, 'regular', 'C',  NULL),
('116456', 'BHATIA PETROLEUM',               'Chanakyapuri Inner Ring, New Delhi 110021',  'New Delhi',       'DELHI-C',  23, 'regular', 'C',  NULL),
('118012', 'DIAMOND FILLING STATION',        'Pitampura Outer Ring, North West Delhi',     'North West Delhi','DELHI-NW', 10, 'regular', 'C',  NULL),
('111789', 'RISHI PETROLEUM',                'Paharganj Extension, Central Delhi',         'Central Delhi',   'DELHI-C',  20, 'regular', 'C',  NULL),
('114234', 'SUNRISE AUTO CENTRE',            'Saket District Centre, South Delhi 110017',  'South Delhi',     'DELHI-S',  5,  'regular', 'C',  NULL),
('117567', 'JUPITER FILLING STATION',        'Tilak Nagar Ext., West Delhi 110018',        'West Delhi',      'DELHI-W',  3,  'regular', 'C',  NULL),
('116789', 'BHAGWATI FILLING STATION',       'Dwarka Sector 10, South West Delhi 110075',  'South West Delhi','DELHI-SW', 27, 'regular', 'C',  NULL),
('112012', 'MODERN FUELS',                   'Shahdara North Ring Rd, Shahdara 110032',    'Shahdara',        'DELHI-E',  25, 'regular', 'C',  NULL),
('118567', 'BP-GOLDEN PARK',                 'Shalimar Bagh Outer, North West Delhi',      'North West Delhi','DELHI-NW', 11, 'regular', 'C',  NULL),
('111456', 'FAST TRACK FILLING STATION',     'Preet Vihar Outer, East Delhi 110092',       'East Delhi',      'DELHI-E',  12, 'regular', 'C',  NULL),
('109678', 'KRITI NANAK FILLING STN.',       'Kirti Nagar Industrial, West Delhi 110015',  'West Delhi',      'DELHI-W',  1,  'regular', 'C',  NULL),
('259713', 'SAKSHAM MOTORS ADHOC',           'Dilshad Garden (Temporary), NE Delhi',       'North East Delhi','DELHI-NE', 21, 'adhoc',   NULL, NULL),
('115012', 'MAHADEV FILLING STATION',        'Karol Bagh Outer Ring, Central Delhi',       'Central Delhi',   'DELHI-C',  19, 'regular', 'C',  'a1b2c3d4-0001-0001-0001-000000000009');

-- -----------------------------------------------------------------------------
-- 4. PERFORMANCE RECORDS — April 2024 through March 2026 (24 months)
-- Top performers get realistic high volumes; bottom performers get low/declining
-- Covers all 8 products for all 39 outlets
-- Using a PL/pgSQL block for maintainability
-- -----------------------------------------------------------------------------

DO $$
DECLARE
    v_period DATE;
    v_month  INT;
BEGIN
    -- Generate 24 months of data
    FOR v_month IN 0..23 LOOP
        v_period := DATE_TRUNC('month', '2024-04-01'::DATE + (v_month || ' months')::INTERVAL);

        -- Top performer: M.L. SETHI SERVICE STATION (112847)
        INSERT INTO performance_records (cc_number, product_id, period, achieved, last_year, volume_kl, source)
        VALUES
        ('112847', 1, v_period, ROUND((195 + v_month*1.2 + RANDOM()*20 - 10)::NUMERIC, 2),
         ROUND((180 + v_month*1.0 + RANDOM()*15)::NUMERIC, 2),
         ROUND((195 + v_month*1.2 + RANDOM()*20 - 10)::NUMERIC, 3), 'manual'),
        ('112847', 2, v_period, ROUND((420 + v_month*2.0 + RANDOM()*40 - 20)::NUMERIC, 2),
         ROUND((390 + v_month*1.8 + RANDOM()*30)::NUMERIC, 2),
         ROUND((420 + v_month*2.0 + RANDOM()*40 - 20)::NUMERIC, 3), 'manual'),
        ('112847', 3, v_period, ROUND((52 + v_month*0.5 + RANDOM()*8 - 4)::NUMERIC, 2),
         ROUND((48 + v_month*0.4 + RANDOM()*6)::NUMERIC, 2),
         ROUND((52 + v_month*0.5 + RANDOM()*8 - 4)::NUMERIC, 3), 'manual'),
        ('112847', 4, v_period, ROUND((85 + v_month*0.8 + RANDOM()*10)::NUMERIC, 0), ROUND((78 + v_month*0.6)::NUMERIC,0), NULL, 'manual'),
        ('112847', 5, v_period, ROUND((185000 + v_month*2000 + RANDOM()*20000)::NUMERIC, 2), ROUND((170000 + v_month*1800)::NUMERIC,2), NULL, 'manual'),
        ('112847', 6, v_period, ROUND((380 + v_month*4 + RANDOM()*40)::NUMERIC, 0), ROUND((350 + v_month*3)::NUMERIC,0), NULL, 'manual'),
        ('112847', 7, v_period, ROUND((210 + v_month*2 + RANDOM()*20)::NUMERIC, 0), ROUND((195 + v_month*1.5)::NUMERIC,0), NULL, 'manual'),
        ('112847', 8, v_period, ROUND((68000 + v_month*800 + RANDOM()*8000)::NUMERIC, 2), ROUND((62000 + v_month*700)::NUMERIC,2), NULL, 'manual')
        ON CONFLICT (cc_number, product_id, period) DO NOTHING;

        -- #2 SAI FILLING STATION (108923) — strong performer
        INSERT INTO performance_records (cc_number, product_id, period, achieved, last_year, volume_kl, source)
        VALUES
        ('108923', 1, v_period, ROUND((185+v_month*1.1+RANDOM()*18-9)::NUMERIC,2), ROUND((170+v_month*0.9+RANDOM()*12)::NUMERIC,2), ROUND((185+v_month*1.1+RANDOM()*18-9)::NUMERIC,3),'manual'),
        ('108923', 2, v_period, ROUND((400+v_month*1.8+RANDOM()*35-17)::NUMERIC,2), ROUND((375+v_month*1.6+RANDOM()*25)::NUMERIC,2), ROUND((400+v_month*1.8+RANDOM()*35-17)::NUMERIC,3),'manual'),
        ('108923', 3, v_period, ROUND((48+v_month*0.4+RANDOM()*7-3)::NUMERIC,2), ROUND((44+v_month*0.3+RANDOM()*5)::NUMERIC,2), ROUND((48+v_month*0.4+RANDOM()*7-3)::NUMERIC,3),'manual'),
        ('108923', 4, v_period, ROUND((78+v_month*0.7+RANDOM()*8)::NUMERIC,0), ROUND((72+v_month*0.5)::NUMERIC,0), NULL,'manual'),
        ('108923', 5, v_period, ROUND((175000+v_month*1800+RANDOM()*18000)::NUMERIC,2), ROUND((162000+v_month*1600)::NUMERIC,2), NULL,'manual'),
        ('108923', 6, v_period, ROUND((360+v_month*3.5+RANDOM()*35)::NUMERIC,0), ROUND((330+v_month*3)::NUMERIC,0), NULL,'manual'),
        ('108923', 7, v_period, ROUND((200+v_month*1.8+RANDOM()*18)::NUMERIC,0), ROUND((185+v_month*1.4)::NUMERIC,0), NULL,'manual'),
        ('108923', 8, v_period, ROUND((64000+v_month*750+RANDOM()*7000)::NUMERIC,2), ROUND((59000+v_month*650)::NUMERIC,2), NULL,'manual')
        ON CONFLICT (cc_number, product_id, period) DO NOTHING;

        -- Mid-range performers (outlets 3-10): moderate volumes with positive trends
        -- AUTO CARE (115634)
        INSERT INTO performance_records (cc_number, product_id, period, achieved, last_year, volume_kl, source)
        VALUES
        ('115634', 1, v_period, ROUND((145+v_month*0.9+RANDOM()*15-7)::NUMERIC,2), ROUND((133+v_month*0.7+RANDOM()*10)::NUMERIC,2), ROUND((145+v_month*0.9+RANDOM()*15-7)::NUMERIC,3),'manual'),
        ('115634', 2, v_period, ROUND((310+v_month*1.4+RANDOM()*28-14)::NUMERIC,2), ROUND((288+v_month*1.2+RANDOM()*20)::NUMERIC,2), ROUND((310+v_month*1.4+RANDOM()*28-14)::NUMERIC,3),'manual'),
        ('115634', 3, v_period, ROUND((38+v_month*0.3+RANDOM()*5-2)::NUMERIC,2), ROUND((35+v_month*0.25+RANDOM()*4)::NUMERIC,2), ROUND((38+v_month*0.3+RANDOM()*5-2)::NUMERIC,3),'manual'),
        ('115634', 4, v_period, ROUND((62+v_month*0.5+RANDOM()*6)::NUMERIC,0), ROUND((57+v_month*0.4)::NUMERIC,0), NULL,'manual'),
        ('115634', 5, v_period, ROUND((138000+v_month*1400+RANDOM()*14000)::NUMERIC,2), ROUND((127000+v_month*1200)::NUMERIC,2), NULL,'manual'),
        ('115634', 6, v_period, ROUND((280+v_month*2.5+RANDOM()*25)::NUMERIC,0), ROUND((258+v_month*2)::NUMERIC,0), NULL,'manual'),
        ('115634', 7, v_period, ROUND((155+v_month*1.2+RANDOM()*14)::NUMERIC,0), ROUND((143+v_month*1)::NUMERIC,0), NULL,'manual'),
        ('115634', 8, v_period, ROUND((50000+v_month*580+RANDOM()*5500)::NUMERIC,2), ROUND((46000+v_month*500)::NUMERIC,2), NULL,'manual')
        ON CONFLICT (cc_number, product_id, period) DO NOTHING;

        -- Bottom performers: MAHADEV (115012) — extreme decline
        INSERT INTO performance_records (cc_number, product_id, period, achieved, last_year, volume_kl, source)
        VALUES
        ('115012', 1, v_period, GREATEST(0, ROUND((95 - v_month*3.5 + RANDOM()*10-5)::NUMERIC,2)), ROUND((140+v_month*0.5+RANDOM()*8)::NUMERIC,2), GREATEST(0, ROUND((95-v_month*3.5+RANDOM()*10-5)::NUMERIC,3)),'manual'),
        ('115012', 2, v_period, GREATEST(0, ROUND((180 - v_month*6 + RANDOM()*15-7)::NUMERIC,2)), ROUND((260+v_month*0.8+RANDOM()*12)::NUMERIC,2), GREATEST(0, ROUND((180-v_month*6+RANDOM()*15-7)::NUMERIC,3)),'manual'),
        ('115012', 3, v_period, GREATEST(0, ROUND((18 - v_month*0.5 + RANDOM()*3-1)::NUMERIC,2)), ROUND((24+v_month*0.2+RANDOM()*2)::NUMERIC,2), GREATEST(0, ROUND((18-v_month*0.5+RANDOM()*3-1)::NUMERIC,3)),'manual'),
        ('115012', 4, v_period, 0, ROUND((28+v_month*0.2)::NUMERIC,0), NULL,'manual'),
        ('115012', 5, v_period, 0, ROUND((62000+v_month*500)::NUMERIC,2), NULL,'manual'),
        ('115012', 6, v_period, 0, ROUND((125+v_month*1)::NUMERIC,0), NULL,'manual'),
        ('115012', 7, v_period, 0, ROUND((68+v_month*0.5)::NUMERIC,0), NULL,'manual'),
        ('115012', 8, v_period, 0, ROUND((22000+v_month*200)::NUMERIC,2), NULL,'manual')
        ON CONFLICT (cc_number, product_id, period) DO NOTHING;

        -- BHAGWATI (116789) — -45% MS decline
        INSERT INTO performance_records (cc_number, product_id, period, achieved, last_year, volume_kl, source)
        VALUES
        ('116789', 1, v_period, GREATEST(0,ROUND((88-v_month*2.8+RANDOM()*8-4)::NUMERIC,2)), ROUND((120+v_month*0.4+RANDOM()*6)::NUMERIC,2), GREATEST(0,ROUND((88-v_month*2.8+RANDOM()*8-4)::NUMERIC,3)),'manual'),
        ('116789', 2, v_period, ROUND((155+v_month*0.6+RANDOM()*15-7)::NUMERIC,2), ROUND((165+v_month*0.5+RANDOM()*10)::NUMERIC,2), ROUND((155+v_month*0.6+RANDOM()*15-7)::NUMERIC,3),'manual'),
        ('116789', 3, v_period, ROUND((22+v_month*0.1+RANDOM()*3-1)::NUMERIC,2), ROUND((20+v_month*0.1+RANDOM()*2)::NUMERIC,2), ROUND((22+v_month*0.1+RANDOM()*3-1)::NUMERIC,3),'manual'),
        ('116789', 4, v_period, ROUND((28+v_month*0.2+RANDOM()*3)::NUMERIC,0), ROUND((32+v_month*0.2)::NUMERIC,0), NULL,'manual'),
        ('116789', 5, v_period, ROUND((62000+v_month*400+RANDOM()*5000)::NUMERIC,2), ROUND((68000+v_month*350)::NUMERIC,2), NULL,'manual'),
        ('116789', 6, v_period, ROUND((115+v_month*0.8+RANDOM()*10)::NUMERIC,0), ROUND((125+v_month*0.7)::NUMERIC,0), NULL,'manual'),
        ('116789', 7, v_period, ROUND((58+v_month*0.4+RANDOM()*5)::NUMERIC,0), ROUND((62+v_month*0.3)::NUMERIC,0), NULL,'manual'),
        ('116789', 8, v_period, ROUND((21000+v_month*150+RANDOM()*2000)::NUMERIC,2), ROUND((23000+v_month*120)::NUMERIC,2), NULL,'manual')
        ON CONFLICT (cc_number, product_id, period) DO NOTHING;

        -- Generate mid-range data for remaining outlets (104 outlets × 8 products)
        -- Outlets 10-37 (mid-performers with varied trends)
        INSERT INTO performance_records (cc_number, product_id, period, achieved, last_year, volume_kl, source)
        SELECT
            o.cc_number,
            p.id,
            v_period,
            CASE
                WHEN p.category = 'fuel' AND p.code = 'MS'   THEN GREATEST(0, ROUND((120 + (HASHTEXT(o.cc_number) % 80) + v_month*(0.5 + (HASHTEXT(o.cc_number||'t') % 20)/20.0) + RANDOM()*12-6)::NUMERIC,2))
                WHEN p.category = 'fuel' AND p.code = 'HSD'  THEN GREATEST(0, ROUND((250 + (HASHTEXT(o.cc_number) % 150) + v_month*(0.8 + (HASHTEXT(o.cc_number||'h') % 30)/20.0) + RANDOM()*20-10)::NUMERIC,2))
                WHEN p.category = 'fuel' AND p.code = 'SPEED' THEN GREATEST(0, ROUND((25 + (HASHTEXT(o.cc_number) % 30) + v_month*0.3 + RANDOM()*5-2)::NUMERIC,2))
                WHEN p.code = 'QOC'         THEN GREATEST(0, ROUND((40 + (HASHTEXT(o.cc_number) % 40) + v_month*0.4 + RANDOM()*5)::NUMERIC,0))
                WHEN p.code = 'Lubricants'  THEN GREATEST(0, ROUND((95000 + (HASHTEXT(o.cc_number) % 80000) + v_month*900 + RANDOM()*9000)::NUMERIC,2))
                WHEN p.code = 'UFill'       THEN GREATEST(0, ROUND((200 + (HASHTEXT(o.cc_number) % 200) + v_month*2 + RANDOM()*20)::NUMERIC,0))
                WHEN p.code = 'SBI'         THEN GREATEST(0, ROUND((110 + (HASHTEXT(o.cc_number) % 100) + v_month*1 + RANDOM()*10)::NUMERIC,0))
                WHEN p.code = 'BeCafe'      THEN GREATEST(0, ROUND((32000 + (HASHTEXT(o.cc_number) % 30000) + v_month*350 + RANDOM()*3000)::NUMERIC,2))
            END as achieved,
            CASE
                WHEN p.category = 'fuel' AND p.code = 'MS'   THEN GREATEST(0, ROUND((110 + (HASHTEXT(o.cc_number) % 80) + v_month*0.4 + RANDOM()*10)::NUMERIC,2))
                WHEN p.category = 'fuel' AND p.code = 'HSD'  THEN GREATEST(0, ROUND((230 + (HASHTEXT(o.cc_number) % 150) + v_month*0.7 + RANDOM()*18)::NUMERIC,2))
                WHEN p.category = 'fuel' AND p.code = 'SPEED' THEN GREATEST(0, ROUND((23 + (HASHTEXT(o.cc_number) % 30) + v_month*0.25 + RANDOM()*4)::NUMERIC,2))
                WHEN p.code = 'QOC'         THEN GREATEST(0, ROUND((37 + (HASHTEXT(o.cc_number) % 40) + v_month*0.3 + RANDOM()*4)::NUMERIC,0))
                WHEN p.code = 'Lubricants'  THEN GREATEST(0, ROUND((88000 + (HASHTEXT(o.cc_number) % 80000) + v_month*800 + RANDOM()*8000)::NUMERIC,2))
                WHEN p.code = 'UFill'       THEN GREATEST(0, ROUND((185 + (HASHTEXT(o.cc_number) % 200) + v_month*1.5 + RANDOM()*18)::NUMERIC,0))
                WHEN p.code = 'SBI'         THEN GREATEST(0, ROUND((100 + (HASHTEXT(o.cc_number) % 100) + v_month*0.8 + RANDOM()*9)::NUMERIC,0))
                WHEN p.code = 'BeCafe'      THEN GREATEST(0, ROUND((29000 + (HASHTEXT(o.cc_number) % 30000) + v_month*300 + RANDOM()*2500)::NUMERIC,2))
            END as last_year,
            CASE WHEN p.category = 'fuel'
                 THEN GREATEST(0, ROUND((CASE p.code
                     WHEN 'MS'    THEN 120 + (HASHTEXT(o.cc_number) % 80) + v_month*(0.5 + (HASHTEXT(o.cc_number||'t') % 20)/20.0) + RANDOM()*12-6
                     WHEN 'HSD'   THEN 250 + (HASHTEXT(o.cc_number) % 150) + v_month*(0.8 + (HASHTEXT(o.cc_number||'h') % 30)/20.0) + RANDOM()*20-10
                     WHEN 'SPEED' THEN 25 + (HASHTEXT(o.cc_number) % 30) + v_month*0.3 + RANDOM()*5-2
                 END)::NUMERIC,3))
                 ELSE NULL
            END as volume_kl,
            'manual'
        FROM retail_outlets o
        CROSS JOIN products p
        WHERE o.cc_number NOT IN ('112847','108923','115634','115012','116789')
        ON CONFLICT (cc_number, product_id, period) DO NOTHING;

    END LOOP;
END $$;

-- -----------------------------------------------------------------------------
-- 5. TARGETS — March 2026 period for all outlets
-- -----------------------------------------------------------------------------
INSERT INTO targets (cc_number, product_id, period, target_value, set_by)
SELECT
    o.cc_number,
    p.id,
    '2026-03-01'::DATE,
    CASE
        WHEN p.code = 'MS'         THEN ROUND((pr.achieved * 1.08)::NUMERIC, 0)
        WHEN p.code = 'HSD'        THEN ROUND((pr.achieved * 1.06)::NUMERIC, 0)
        WHEN p.code = 'SPEED'      THEN ROUND((pr.achieved * 1.10)::NUMERIC, 0)
        WHEN p.code = 'QOC'        THEN ROUND((pr.achieved * 1.12)::NUMERIC, 0)
        WHEN p.code = 'Lubricants' THEN ROUND((pr.achieved * 1.10)::NUMERIC, 0)
        WHEN p.code = 'UFill'      THEN ROUND((pr.achieved * 1.15)::NUMERIC, 0)
        WHEN p.code = 'SBI'        THEN ROUND((pr.achieved * 1.10)::NUMERIC, 0)
        WHEN p.code = 'BeCafe'     THEN ROUND((pr.achieved * 1.08)::NUMERIC, 0)
    END,
    (SELECT id FROM users WHERE role = 'territory_manager' LIMIT 1)
FROM retail_outlets o
CROSS JOIN products p
JOIN performance_records pr ON pr.cc_number = o.cc_number
    AND pr.product_id = p.id
    AND pr.period = '2026-02-01'::DATE
WHERE pr.achieved IS NOT NULL AND pr.achieved > 0
ON CONFLICT (cc_number, product_id, period) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 6. COMPETITION PERIOD — Boost and Win March 2026
-- -----------------------------------------------------------------------------
INSERT INTO competition_periods (id, name, period, territory_code, status, published_at, created_by)
VALUES (
    'c0000000-0001-0001-0001-000000000001',
    'Boost and Win — March 2026',
    '2026-03-01',
    'DELHI-ALL',
    'published',
    '2026-04-10 10:00:00+05:30',
    (SELECT id FROM users WHERE role = 'admin' LIMIT 1)
);

-- -----------------------------------------------------------------------------
-- 7. COMPETITION SCORES — March 2026 (known results from project data)
-- -----------------------------------------------------------------------------
INSERT INTO competition_scores
    (competition_id, cc_number, total_score, rank,
     score_ms_vol, score_ms_growth, score_ms_ta_gain,
     score_hsd_vol, score_hsd_growth, score_hsd_ta_gain,
     score_qoc, score_lubricants, score_ufill, score_speed,
     score_cleanliness, score_sangam, score_google, score_ips, score_bonus, computed_at)
VALUES
('c0000000-0001-0001-0001-000000000001', '112847', 56.37, 1,  9.8, 3.2, NULL, 9.5, 3.1, NULL, 8.9, 8.7, 4.2, 4.5, 4.0, 3.8, 4.2, 3.7, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '108923', 55.35, 2,  9.5, 3.0, NULL, 9.2, 2.9, NULL, 8.6, 8.5, 4.0, 4.3, 4.0, 3.7, 4.1, 3.5, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '115634', 48.17, 3,  8.2, 2.5, NULL, 8.0, 2.4, NULL, 7.5, 7.2, 3.5, 3.8, 3.5, 3.2, 3.8, 3.0, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '109251', 45.85, 4,  7.8, 2.3, NULL, 7.6, 2.2, NULL, 7.2, 6.9, 3.3, 3.6, 3.5, 3.0, 3.5, 2.9, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '113782', 45.73, 5,  7.7, 2.2, NULL, 7.5, 2.1, NULL, 7.1, 6.8, 3.2, 3.5, 3.5, 3.0, 3.5, 2.8, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '107634', 43.68, 6,  7.4, 2.0, NULL, 7.2, 1.9, NULL, 6.9, 6.5, 3.0, 3.3, 3.0, 2.8, 3.3, 2.7, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '114521', 41.50, 7,  7.0, 1.8, NULL, 6.8, 1.7, NULL, 6.5, 6.2, 2.8, 3.1, 3.0, 2.6, 3.1, 2.5, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '116890', 37.02, 8,  6.2, 1.5, NULL, 6.0, 1.4, NULL, 5.8, 5.5, 2.5, 2.8, 2.5, 2.3, 2.8, 2.2, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '111234', 35.60, 9,  5.9, 1.3, NULL, 5.7, 1.2, NULL, 5.5, 5.2, 2.3, 2.6, 2.5, 2.1, 2.6, 2.0, 10, NOW()),
('c0000000-0001-0001-0001-000000000001', '108765', 34.20, 10, 5.6, 1.2, NULL, 5.5, 1.1, NULL, 5.2, 4.9, 2.1, 2.4, 2.5, 2.0, 2.4, 1.9, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '116789', 14.85, 34, 2.1,-3.8, NULL, 3.5, 1.0, NULL, 2.8, 2.5, 1.2, 1.5, 1.5, 0.8, 1.5, 1.2, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '118567', 13.57, 36, 2.0,-1.5, NULL, 2.8,-0.8, NULL, 2.2, 2.0, 0.9, 1.2, 1.0, 0.6, 1.2, 0.8, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '111456', 11.50, 37, 1.8,-2.5, NULL, 2.2,-1.8, NULL, 1.8, 1.5, 0.7, 0.8, 1.0, 0.5, 1.0, 0.7, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '109678', 10.74, 38, 1.5, 0.5, NULL, 1.8, 0.4, NULL, 0.0, 0.0, 0.9, 1.0, 2.5, 0.7, 1.0, 0.9, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '259713', 10.10, 39, 1.6, 0.0, NULL, 1.5, 0.0, NULL, 0.0, 0.0, 0.0, 0.8, 0.0, 0.0, 0.0, 0.0, 0,  NOW()),
('c0000000-0001-0001-0001-000000000001', '115012', -5.65, 40, 1.0,-4.5, NULL, 1.2,-5.0, NULL, 0.0, 0.0, 0.0, 0.5, 0.0, 0.0, 0.5, 0.0, 0,  NOW());

-- Fill mid-range scores for remaining outlets (ranks 11-33)
INSERT INTO competition_scores
    (competition_id, cc_number, total_score, rank, score_ms_vol, score_ms_growth,
     score_hsd_vol, score_hsd_growth, score_qoc, score_lubricants, score_ufill,
     score_speed, score_cleanliness, score_sangam, score_google, score_ips, score_bonus, computed_at)
SELECT
    'c0000000-0001-0001-0001-000000000001',
    o.cc_number,
    ROUND((15 + (HASHTEXT(o.cc_number) % 170) / 10.0)::NUMERIC, 2),
    11 + (ROW_NUMBER() OVER (ORDER BY HASHTEXT(o.cc_number)) - 1)::INT,
    ROUND((3.0 + (HASHTEXT(o.cc_number||'ms') % 40)/10.0)::NUMERIC,2),
    ROUND((-1.0 + (HASHTEXT(o.cc_number||'gr') % 30)/10.0)::NUMERIC,2),
    ROUND((3.5 + (HASHTEXT(o.cc_number||'hd') % 40)/10.0)::NUMERIC,2),
    ROUND((-0.5 + (HASHTEXT(o.cc_number||'hg') % 25)/10.0)::NUMERIC,2),
    ROUND((2.5 + (HASHTEXT(o.cc_number||'qc') % 50)/10.0)::NUMERIC,2),
    ROUND((2.0 + (HASHTEXT(o.cc_number||'lb') % 55)/10.0)::NUMERIC,2),
    ROUND((1.0 + (HASHTEXT(o.cc_number||'uf') % 35)/10.0)::NUMERIC,2),
    ROUND((1.0 + (HASHTEXT(o.cc_number||'sp') % 35)/10.0)::NUMERIC,2),
    ROUND((1.5 + (HASHTEXT(o.cc_number||'cl') % 25)/10.0)::NUMERIC,2),
    ROUND((1.0 + (HASHTEXT(o.cc_number||'sg') % 30)/10.0)::NUMERIC,2),
    ROUND((1.5 + (HASHTEXT(o.cc_number||'gg') % 25)/10.0)::NUMERIC,2),
    ROUND((1.0 + (HASHTEXT(o.cc_number||'ip') % 30)/10.0)::NUMERIC,2),
    0,
    NOW()
FROM retail_outlets o
WHERE o.cc_number NOT IN (
    '112847','108923','115634','109251','113782','107634','114521',
    '116890','111234','108765','116789','118567','111456','109678','259713','115012'
)
ON CONFLICT (competition_id, cc_number) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 8. BONUS MARKS — Sanjeev (rank 9) with documented reason (Bug #4 fix)
-- -----------------------------------------------------------------------------
INSERT INTO competition_bonus (competition_id, cc_number, bonus_marks, remarks, awarded_by)
VALUES (
    'c0000000-0001-0001-0001-000000000001',
    '111234',
    10,
    'Exceptional performance during CNG pilot launch in March 2026. Dealer coordinated with 3 fleet accounts and achieved 100% UFill adoption ahead of territory target.',
    (SELECT id FROM users WHERE role = 'territory_manager' AND territory_code = 'DELHI-W' LIMIT 1)
);

-- -----------------------------------------------------------------------------
-- 9. DEALER AUDIT SCORES — March 2026
-- -----------------------------------------------------------------------------
INSERT INTO dealer_audit_scores (cc_number, period, cleanliness_grade, cleanliness_status, sangam_count, google_rating, ips_pct, audit_status, audited_at)
VALUES
('112847', '2026-03-01', 'Excellent',     'audited', 42, 4.6, 92.0, 'audited', '2026-03-28 10:00:00+05:30'),
('108923', '2026-03-01', 'Excellent',     'audited', 38, 4.5, 88.0, 'audited', '2026-03-28 11:00:00+05:30'),
('115634', '2026-03-01', 'Good',          'audited', 32, 4.3, 82.0, 'audited', '2026-03-29 09:00:00+05:30'),
('109251', '2026-03-01', 'Good',          'audited', 30, 4.2, 79.0, 'audited', '2026-03-29 10:00:00+05:30'),
('113782', '2026-03-01', 'Good',          'audited', 29, 4.2, 78.0, 'audited', '2026-03-29 11:00:00+05:30'),
('107634', '2026-03-01', 'Good',          'audited', 27, 4.0, 74.0, 'audited', '2026-03-30 09:00:00+05:30'),
('114521', '2026-03-01', 'Good',          'audited', 25, 4.0, 70.0, 'audited', '2026-03-30 10:00:00+05:30'),
('116890', '2026-03-01', 'Average',       'audited', 22, 3.8, 65.0, 'audited', '2026-03-30 11:00:00+05:30'),
('111234', '2026-03-01', 'Average',       'audited', 20, 3.7, 62.0, 'audited', '2026-03-31 09:00:00+05:30'),
('108765', '2026-03-01', 'Average',       'audited', 19, 3.6, 60.0, 'audited', '2026-03-31 10:00:00+05:30'),
('116789', '2026-03-01', 'Below Average', 'audited', 8,  3.2, 38.0, 'audited', '2026-03-25 09:00:00+05:30'),
('118567', '2026-03-01', 'Below Average', 'audited', 6,  3.0, 32.0, 'audited', '2026-03-25 10:00:00+05:30'),
('111456', '2026-03-01', 'Below Average', 'audited', 5,  3.1, 30.0, 'audited', '2026-03-25 11:00:00+05:30'),
('109678', '2026-03-01', 'Average',       'audited', 18, 3.5, 52.0, 'audited', '2026-03-26 09:00:00+05:30'),
('115012', '2026-03-01', 'Poor',          'audited', 0,  2.8, 18.0, 'audited', '2026-03-24 09:00:00+05:30'),
('259713', '2026-03-01', NULL,            'not_audited', NULL, NULL, NULL, 'not_audited', NULL);

-- Add partial audit data for remaining outlets
INSERT INTO dealer_audit_scores (cc_number, period, cleanliness_grade, cleanliness_status, sangam_count, google_rating, ips_pct, audit_status, audited_at)
SELECT
    o.cc_number,
    '2026-03-01'::DATE,
    CASE (HASHTEXT(o.cc_number) % 5)
        WHEN 0 THEN 'Excellent' WHEN 1 THEN 'Good' WHEN 2 THEN 'Average'
        WHEN 3 THEN 'Below Average' ELSE 'Good'
    END,
    'audited',
    10 + (HASHTEXT(o.cc_number||'sg') % 25),
    ROUND((3.2 + (HASHTEXT(o.cc_number||'gr') % 18)/10.0)::NUMERIC,1),
    ROUND((45 + (HASHTEXT(o.cc_number||'ip') % 40))::NUMERIC,1),
    'audited',
    ('2026-03-26'::TIMESTAMPTZ + ((HASHTEXT(o.cc_number) % 5) || ' days')::INTERVAL)
FROM retail_outlets o
WHERE o.cc_number NOT IN (
    '112847','108923','115634','109251','113782','107634','114521',
    '116890','111234','108765','116789','118567','111456','109678','115012','259713'
)
ON CONFLICT (cc_number, period) DO NOTHING;

-- -----------------------------------------------------------------------------
-- Verification counts
-- -----------------------------------------------------------------------------
SELECT 'users' AS tbl, COUNT(*) FROM users
UNION ALL SELECT 'trading_areas', COUNT(*) FROM trading_areas
UNION ALL SELECT 'retail_outlets', COUNT(*) FROM retail_outlets
UNION ALL SELECT 'products', COUNT(*) FROM products
UNION ALL SELECT 'performance_records', COUNT(*) FROM performance_records
UNION ALL SELECT 'targets', COUNT(*) FROM targets
UNION ALL SELECT 'competition_periods', COUNT(*) FROM competition_periods
UNION ALL SELECT 'competition_scores', COUNT(*) FROM competition_scores
UNION ALL SELECT 'dealer_audit_scores', COUNT(*) FROM dealer_audit_scores
UNION ALL SELECT 'competition_bonus', COUNT(*) FROM competition_bonus;
