CREATE TABLE IF NOT EXISTS match_stats (
    id SERIAL PRIMARY KEY,
    matchday_id INTEGER NOT NULL REFERENCES matchdays(id) ON DELETE CASCADE,
    player_id INTEGER NOT NULL REFERENCES players(id) ON DELETE CASCADE,
    goals INTEGER NOT NULL DEFAULT 0,
    assists INTEGER NOT NULL DEFAULT 0,
    yellow_cards INTEGER NOT NULL DEFAULT 0,
    red_cards INTEGER NOT NULL DEFAULT 0,
    UNIQUE (matchday_id, player_id)
);

CREATE INDEX idx_match_stats_player_id ON match_stats(player_id);
CREATE INDEX idx_match_stats_matchday_id ON match_stats(matchday_id);
