CREATE TABLE IF NOT EXISTS metric_snapshots (
    id SERIAL PRIMARY KEY,
    engineer_id INTEGER NOT NULL REFERENCES engineers(id) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    prs_raised INTEGER NOT NULL DEFAULT 0,
    prs_reviewed INTEGER NOT NULL DEFAULT 0,
    tickets_closed INTEGER NOT NULL DEFAULT 0,
    complexity_score NUMERIC(6,2) NOT NULL DEFAULT 0,
    synced_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE (engineer_id, period_start, period_end)
);

CREATE INDEX idx_metric_snapshots_engineer_id ON metric_snapshots(engineer_id);
