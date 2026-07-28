CREATE TABLE IF NOT EXISTS matchdays (
    id SERIAL PRIMARY KEY,
    played_on DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_matchdays_played_on ON matchdays(played_on);
