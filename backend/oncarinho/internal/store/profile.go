package store

import (
	"database/sql"

	"oncarinho/internal/models"
)

type ProfileStore struct {
	db          *sql.DB
	playerStore *PlayerStore
}

func NewProfileStore(db *sql.DB, playerStore *PlayerStore) *ProfileStore {
	return &ProfileStore{db: db, playerStore: playerStore}
}

func (s *ProfileStore) Profile(playerID int) (*models.PlayerProfile, error) {
	player, err := s.playerStore.GetByID(playerID)
	if err != nil {
		return nil, err
	}
	if player == nil {
		return nil, nil
	}

	var allTime models.StatTotals
	err = s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(goals), 0), COALESCE(SUM(assists), 0),
		       COALESCE(SUM(yellow_cards), 0), COALESCE(SUM(red_cards), 0)
		FROM match_stats WHERE player_id = $1
	`, playerID).Scan(&allTime.MatchesPlayed, &allTime.Goals, &allTime.Assists, &allTime.YellowCards, &allTime.RedCards)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(`
		SELECT EXTRACT(YEAR FROM m.played_on)::int AS year,
		       COUNT(*), COALESCE(SUM(ms.goals), 0), COALESCE(SUM(ms.assists), 0),
		       COALESCE(SUM(ms.yellow_cards), 0), COALESCE(SUM(ms.red_cards), 0)
		FROM match_stats ms
		JOIN matchdays m ON m.id = ms.matchday_id
		WHERE ms.player_id = $1
		GROUP BY year
		ORDER BY year DESC
	`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byYear := []models.YearStats{}
	for rows.Next() {
		var y models.YearStats
		if err := rows.Scan(&y.Year, &y.MatchesPlayed, &y.Goals, &y.Assists, &y.YellowCards, &y.RedCards); err != nil {
			return nil, err
		}
		byYear = append(byYear, y)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &models.PlayerProfile{Player: *player, AllTime: allTime, ByYear: byYear}, nil
}
