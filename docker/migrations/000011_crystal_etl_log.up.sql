-- ETL health log: every cron run POSTs here (success or failure)
CREATE TABLE cr_etl_log (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  run_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  status       TEXT NOT NULL CHECK (status IN ('ok', 'error')),
  detail       TEXT,
  duration_ms  INT,
  source       TEXT DEFAULT 'cron'
);

CREATE INDEX idx_cr_etl_log_run_at ON cr_etl_log (run_at DESC);
