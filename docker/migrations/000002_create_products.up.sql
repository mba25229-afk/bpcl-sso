CREATE TABLE products (
    id              SMALLINT PRIMARY KEY,
    code            VARCHAR(20) UNIQUE NOT NULL,
    name            TEXT NOT NULL,
    category        TEXT NOT NULL CHECK (category IN ('fuel','non_fuel')),
    unit            TEXT NOT NULL,
    display_order   SMALLINT NOT NULL DEFAULT 0
);
