-- =============================================================================
-- BPCL SSO Portal — Development Seed Data
-- Run automatically on first PostgreSQL startup via docker-entrypoint-initdb.d
-- =============================================================================

-- Password for all test users: "password123"
-- Bcrypt hash: $2b$10$omswmnQb2Xd7JqKoWJ1VVOJHBGAcz8rkXHDF3prjOIgVBjj.NODRG

-- -----------------------------------------------------------------------------
-- 1. Test Users
-- -----------------------------------------------------------------------------
INSERT INTO users (id, employee_id, email, name, password_hash, role, territory_code) VALUES
('a1b2c3d4-0001-0001-0001-000000000001', 'EMP10001', 'admin@bpcl.in',
 'System Admin', '$2b$10$omswmnQb2Xd7JqKoWJ1VVOJHBGAcz8rkXHDF3prjOIgVBjj.NODRG', 'admin', NULL),
('a1b2c3d4-0001-0001-0001-000000000002', 'EMP10002', 'tm.delhi.west@bpcl.in',
 'Rajesh Kumar', '$2b$10$omswmnQb2Xd7JqKoWJ1VVOJHBGAcz8rkXHDF3prjOIgVBjj.NODRG', 'territory_manager', 'DELHI-W'),
('a1b2c3d4-0001-0001-0001-000000000005', 'EMP10005', 'ro.112847@bpcl.in',
 'Suresh Sethi', '$2b$10$omswmnQb2Xd7JqKoWJ1VVOJHBGAcz8rkXHDF3prjOIgVBjj.NODRG', 'ro_manager', 'DELHI-W');

-- -----------------------------------------------------------------------------
-- 2. Trading Areas
-- -----------------------------------------------------------------------------
INSERT INTO trading_areas (id, name, district, state) VALUES
(1, 'Kirti Nagar',      'West Delhi',      'Delhi'),
(2, 'Rajouri Garden',   'West Delhi',      'Delhi'),
(3, 'Tilak Nagar',      'West Delhi',      'Delhi'),
(4, 'Janakpuri',        'West Delhi',      'Delhi'),
(5, 'Saket',            'South Delhi',     'Delhi'),
(6, 'Lajpat Nagar',     'South Delhi',     'Delhi'),
(7, 'Hauz Khas',        'South Delhi',     'Delhi'),
(8, 'Greater Kailash',  'South Delhi',     'Delhi'),
(9, 'Rohini',           'North West Delhi','Delhi'),
(10, 'Pitampura',        'North West Delhi','Delhi');

-- -----------------------------------------------------------------------------
-- 3. Retail Outlets (sample of 10)
-- -----------------------------------------------------------------------------
INSERT INTO retail_outlets (id, cc_number, name, address, trading_area_id, territory_code, outlet_type) VALUES
(1, '112847', 'BPCL Kirti Nagar', 'Kirti Nagar, West Delhi', 1, 'DELHI-W', 'regular'),
(2, '108923', 'BPCL Rajouri Garden', 'Rajouri Garden, West Delhi', 2, 'DELHI-W', 'regular'),
(3, '115634', 'BPCL Tilak Nagar', 'Tilak Nagar, West Delhi', 3, 'DELHI-W', 'regular'),
(4, '109251', 'BPCL Janakpuri', 'Janakpuri, West Delhi', 4, 'DELHI-W', 'regular'),
(5, '115012', 'BPCL Saket', 'Saket, South Delhi', 5, 'DELHI-S', 'regular'),
(6, '110234', 'BPCL Lajpat Nagar', 'Lajpat Nagar, South Delhi', 6, 'DELHI-S', 'regular'),
(7, '111567', 'BPCL Hauz Khas', 'Hauz Khas, South Delhi', 7, 'DELHI-S', 'regular'),
(8, '112890', 'BPCL Greater Kailash', 'Greater Kailash, South Delhi', 8, 'DELHI-S', 'regular'),
(9, '113456', 'BPCL Rohini', 'Rohini, North West Delhi', 9, 'DELHI-NW', 'regular'),
(10, '114789', 'BPCL Pitampura', 'Pitampura, North West Delhi', 10, 'DELHI-NW', 'regular');

-- -----------------------------------------------------------------------------
-- 4. Products (if not already seeded by migration)
-- -----------------------------------------------------------------------------
INSERT INTO products (id, code, name, category, unit) VALUES
(1, 'MS', 'MS', 'fuel', 'kL'),
(2, 'HSD', 'HSD', 'fuel', 'kL'),
(3, 'SPEED', 'SPEED', 'fuel', 'kL'),
(4, 'QOC', 'QOC', 'non_fuel', 'Nos'),
(5, 'LUBRICANTS', 'Lubricants', 'non_fuel', '₹'),
(6, 'UFILL', 'UFill', 'non_fuel', 'Nos'),
(7, 'SBI', 'SBI', 'non_fuel', 'Nos'),
(8, 'BECAFE', 'BeCafe', 'non_fuel', '₹')
ON CONFLICT (id) DO NOTHING;

-- -----------------------------------------------------------------------------
-- 5. Competition Period
-- -----------------------------------------------------------------------------
INSERT INTO competition_periods (id, month_year, is_active) VALUES
(1, '2025-05-01', true)
ON CONFLICT (id) DO NOTHING;
