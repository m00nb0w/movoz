package store

import (
	"database/sql"

	"oncarinho/internal/models"
)

type StatStore struct {
	db *sql.DB
}

func NewStatStore(db *sql.DB) *StatStore {
	return &StatStore{db: db}
}

func (s *StatStore) UpsertBulk(matchdayID int, entries []models.StatInput) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO match_stats (matchday_id, player_id, goals, assists, yellow_cards, red_cards)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (matchday_id, player_id)
		DO UPDATE SET goals = $3, assists = $4, yellow_cards = $5, red_cards = $6
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, e := range entries {
		if _, err := stmt.Exec(matchdayID, e.PlayerID, e.Goals, e.Assists, e.YellowCards, e.RedCards); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *StatStore) ListByMatchday(matchdayID int) ([]models.MatchStat, error) {
	rows, err := s.db.Query(
		`SELECT id, matchday_id, player_id, goals, assists, yellow_cards, red_cards
		 FROM match_stats WHERE matchday_id = $1 ORDER BY player_id`,
		matchdayID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := []models.MatchStat{}
	for rows.Next() {
		var m models.MatchStat
		if err := rows.Scan(&m.ID, &m.MatchdayID, &m.PlayerID, &m.Goals, &m.Assists, &m.YellowCards, &m.RedCards); err != nil {
			return nil, err
		}
		stats = append(stats, m)
	}
	return stats, rows.Err()
}

func (s *StatStore) Delete(matchdayID, playerID int) (bool, error) {
	res, err := s.db.Exec(
		"DELETE FROM match_stats WHERE matchday_id = $1 AND player_id = $2",
		matchdayID, playerID,
	)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}
