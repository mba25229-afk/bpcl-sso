-- Add products needed for ETL targets import
INSERT INTO products (id, code, name, category, unit) VALUES 
(9, 'DSW', 'DSW', 'non_fuel', 'Nos'),
(10, 'NITROGEN', 'Nitrogen', 'non_fuel', 'Nos'),
(11, 'MAKGE', 'MAK GE', 'non_fuel', '₹')
ON CONFLICT (id) DO NOTHING;

INSERT INTO products (id, code, name, category, unit) VALUES 
(9, 'DSW', 'DSW', 'non_fuel', 'Nos'),
(10, 'NITROGEN', 'Nitrogen', 'non_fuel', 'Nos'),
(11, 'MAKGE', 'MAK GE', 'non_fuel', '₹')
ON CONFLICT (code) DO NOTHING;