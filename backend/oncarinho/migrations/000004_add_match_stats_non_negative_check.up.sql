ALTER TABLE match_stats
    ADD CONSTRAINT match_stats_non_negative CHECK (
        goals >= 0 AND assists >= 0 AND yellow_cards >= 0 AND red_cards >= 0
    );
