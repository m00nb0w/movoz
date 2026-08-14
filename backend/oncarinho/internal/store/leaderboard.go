package store

import (
	"database/sql"
	"fmt"

	"oncarinho/internal/models"
)

var leaderboardStatColumns = map[string]string{
	"goals":   "ms.goals",
	"assists": "ms.assists",
	"cards":   "(ms.yellow_cards + ms.red_cards)",
}

type LeaderboardStore struct {
	db *sql.DB
}

func NewLeaderboardStore(db *sql.DB) *LeaderboardStore {
	return &LeaderboardStore{db: db}
}

func (s *LeaderboardStore) Leaderboard(year *int, stat string) ([]models.LeaderboardEntry, error) {
	column, ok := leaderboardStatColumns[stat]
	if !ok {
		return nil, fmt.Errorf("unknown stat: %s", stat)
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.name, p.position, p.is_active, COALESCE(SUM(%s), 0) AS value
		FROM players p
		JOIN match_stats ms ON ms.player_id = p.id
		JOIN matchdays m ON m.id = ms.matchday_id
	`, column)

	args := []interface{}{}
	if year != nil {
		query += " WHERE EXTRACT(YEAR FROM m.played_on) = $1"
		args = append(args, *year)
	}
	query += fmt.Sprintf(`
		GROUP BY p.id, p.name, p.position, p.is_active
		HAVING COALESCE(SUM(%s), 0) > 0
		ORDER BY value DESC, p.name ASC
	`, column)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []models.LeaderboardEntry{}
	for rows.Next() {
		var e models.LeaderboardEntry
		if err := rows.Scan(&e.PlayerID, &e.PlayerName, &e.Position, &e.IsActive, &e.Value); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
