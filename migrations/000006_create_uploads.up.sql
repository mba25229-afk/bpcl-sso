CREATE TABLE uploaded_files (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id),
    original_name   TEXT NOT NULL,
    stored_path     TEXT NOT NULL,
    size_bytes      BIGINT,
    mime_type       TEXT,
    upload_type     TEXT NOT NULL DEFAULT 'performance'
                    CHECK (upload_type IN ('performance','delhi_master')),
    status          TEXT NOT NULL DEFAULT 'pending'
                    CHECK (status IN ('pending','processing','done','failed','deleted')),
    error_message   TEXT,
    row_count       INT,
    cc_number       VARCHAR(20),
    period          DATE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_uploads_user   ON uploaded_files(user_id);
CREATE INDEX idx_uploads_status ON uploaded_files(status);

-- Now add the FK from performance_records
ALTER TABLE performance_records
    ADD CONSTRAINT fk_perf_upload
    FOREIGN KEY (uploaded_file_id) REFERENCES uploaded_files(id);
