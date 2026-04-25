CREATE TABLE missing_rating_reports (
    id           BIGSERIAL PRIMARY KEY,
    sticker_id   TEXT        NOT NULL,
    reported_by  UUID        REFERENCES users(id) ON DELETE SET NULL,
    reported_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (sticker_id, reported_by)
);

CREATE INDEX idx_missing_rating_reports_sticker_id ON missing_rating_reports(sticker_id);
