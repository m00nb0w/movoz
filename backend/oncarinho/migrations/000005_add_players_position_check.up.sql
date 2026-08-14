ALTER TABLE players
    ADD CONSTRAINT players_position_valid CHECK (
        position IS NULL OR position IN ('goalkeeper', 'defender', 'midfielder', 'forward')
    );
