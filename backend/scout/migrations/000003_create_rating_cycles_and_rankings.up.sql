CREATE TABLE IF NOT EXISTS rating_cycles (
    id SERIAL PRIMARY KEY,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sub_attribute_rankings (
    id SERIAL PRIMARY KEY,
    cycle_id INTEGER NOT NULL REFERENCES rating_cycles(id) ON DELETE CASCADE,
    sub_attribute_id INTEGER NOT NULL REFERENCES sub_attributes(id) ON DELETE CASCADE,
    engineer_id INTEGER NOT NULL REFERENCES engineers(id) ON DELETE CASCADE,
    rank INTEGER NOT NULL,
    score NUMERIC(5,2) NOT NULL,
    UNIQUE (cycle_id, sub_attribute_id, engineer_id),
    UNIQUE (cycle_id, sub_attribute_id, rank)
);

CREATE INDEX idx_sub_attribute_rankings_cycle_id ON sub_attribute_rankings(cycle_id);
CREATE INDEX idx_sub_attribute_rankings_engineer_id ON sub_attribute_rankings(engineer_id);
