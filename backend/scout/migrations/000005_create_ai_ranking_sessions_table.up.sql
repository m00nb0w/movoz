CREATE TABLE IF NOT EXISTS ai_ranking_sessions (
    id SERIAL PRIMARY KEY,
    cycle_id INTEGER NOT NULL REFERENCES rating_cycles(id) ON DELETE CASCADE,
    sub_attribute_id INTEGER NOT NULL REFERENCES sub_attributes(id) ON DELETE CASCADE,
    transcript JSONB NOT NULL DEFAULT '[]',
    proposed_ranking JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ai_ranking_sessions_cycle_id ON ai_ranking_sessions(cycle_id);
