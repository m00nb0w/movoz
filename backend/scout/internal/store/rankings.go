package store

import (
	"database/sql"
	"errors"
	"fmt"

	"scout/internal/models"
	"scout/internal/scoring"
)

type RankingStore struct {
	db            *sql.DB
	engineerStore *EngineerStore
}

func NewRankingStore(db *sql.DB, engineerStore *EngineerStore) *RankingStore {
	return &RankingStore{db: db, engineerStore: engineerStore}
}

// ErrInvalidRanking wraps a scoring.ValidatePermutation rejection — a client
// error (bad request body) as opposed to any other error SubmitRanking can
// return (failure to reach the DB, a broken transaction, etc.), which are
// infra/server errors. Callers should use errors.Is(err, ErrInvalidRanking)
// to tell the two apart, mirroring the 400-vs-500 split already used in
// cycles.go/sub_attributes.go for pre-store-vs-store failures.
var ErrInvalidRanking = errors.New("invalid ranking submission")

// SubmitRanking validates entries as a strict 1..N permutation of the active
// roster (F6), converts each rank to a score via linear interpolation (F7),
// and replaces all rows for this cycle+sub-attribute in a single transaction
// (delete-then-insert), so re-submission fully overwrites the prior ranking
// rather than merging with it — past cycles are editable via re-submission.
func (s *RankingStore) SubmitRanking(cycleID, subAttributeID int, entries []scoring.RankEntry) ([]models.SubAttributeRanking, error) {
	activeIDs, err := s.engineerStore.ListActiveIDs()
	if err != nil {
		return nil, err
	}
	if err := scoring.ValidatePermutation(entries, activeIDs); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRanking, err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		"DELETE FROM sub_attribute_rankings WHERE cycle_id = $1 AND sub_attribute_id = $2",
		cycleID, subAttributeID,
	); err != nil {
		return nil, err
	}

	n := len(entries)
	stmt, err := tx.Prepare(
		`INSERT INTO sub_attribute_rankings (cycle_id, sub_attribute_id, engineer_id, rank, score)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, cycle_id, sub_attribute_id, engineer_id, rank, score`,
	)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	results := make([]models.SubAttributeRanking, 0, n)
	for _, e := range entries {
		var r models.SubAttributeRanking
		score := scoring.RankToScore(e.Rank, n)
		if err := stmt.QueryRow(cycleID, subAttributeID, e.EngineerID, e.Rank, score).
			Scan(&r.ID, &r.CycleID, &r.SubAttributeID, &r.EngineerID, &r.Rank, &r.Score); err != nil {
			return nil, err
		}
		results = append(results, r)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return results, nil
}

// GetByCycleAndSubAttribute returns whatever has been persisted so far for
// this cycle+sub-attribute (an empty slice if nothing has been submitted).
func (s *RankingStore) GetByCycleAndSubAttribute(cycleID, subAttributeID int) ([]models.SubAttributeRanking, error) {
	rows, err := s.db.Query(
		`SELECT id, cycle_id, sub_attribute_id, engineer_id, rank, score
		 FROM sub_attribute_rankings WHERE cycle_id = $1 AND sub_attribute_id = $2 ORDER BY rank`,
		cycleID, subAttributeID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rankings := []models.SubAttributeRanking{}
	for rows.Next() {
		var r models.SubAttributeRanking
		if err := rows.Scan(&r.ID, &r.CycleID, &r.SubAttributeID, &r.EngineerID, &r.Rank, &r.Score); err != nil {
			return nil, err
		}
		rankings = append(rankings, r)
	}
	return rankings, rows.Err()
}
