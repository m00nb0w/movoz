CREATE TABLE IF NOT EXISTS highlight_entries (
    id SERIAL PRIMARY KEY,
    engineer_id INTEGER NOT NULL REFERENCES engineers(id) ON DELETE CASCADE,
    kind VARCHAR(10) NOT NULL CHECK (kind IN ('highlight', 'lowlight')),
    body TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_highlight_entries_engineer_id ON highlight_entries(engineer_id);
