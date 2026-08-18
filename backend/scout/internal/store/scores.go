package store

import (
	"database/sql"

	"scout/internal/models"
)

type ScoreStore struct {
	db *sql.DB
}

func NewScoreStore(db *sql.DB) *ScoreStore {
	return &ScoreStore{db: db}
}

// MainAttributeScores computes each main attribute's score for an engineer
// in a cycle as the average of that main attribute's sub-attribute scores
// (F8). Computed on read from sub_attribute_rankings — never cached (NF2).
// A main attribute with no sub-attribute rankings for this cycle+engineer
// is simply absent from the result (not returned with a score of 0), and
// this list is NOT gated by the F8 cutover rule — that rule only applies
// to OverallScore below.
func (s *ScoreStore) MainAttributeScores(engineerID, cycleID int) ([]models.MainAttributeScore, error) {
	rows, err := s.db.Query(
		`SELECT ma.id, ma.key, ma.name, AVG(sar.score)
		 FROM sub_attribute_rankings sar
		 JOIN sub_attributes sa ON sa.id = sar.sub_attribute_id
		 JOIN main_attributes ma ON ma.id = sa.main_attribute_id
		 WHERE sar.cycle_id = $1 AND sar.engineer_id = $2
		 GROUP BY ma.id, ma.key, ma.name
		 ORDER BY ma.id`,
		cycleID, engineerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	scores := []models.MainAttributeScore{}
	for rows.Next() {
		var m models.MainAttributeScore
		if err := rows.Scan(&m.MainAttributeID, &m.Key, &m.Name, &m.Score); err != nil {
			return nil, err
		}
		scores = append(scores, m)
	}
	return scores, rows.Err()
}

// OverallScore computes the average of the main-attribute scores that
// existed as of this cycle (F8): a main attribute counts toward Overall
// only if it was created at or before the cycle was opened, so adding a
// main attribute later never retroactively changes a past cycle's Overall.
// Returns nil if the engineer has no rankings for the cycle at all (or
// none from a main attribute that existed as of that cycle).
func (s *ScoreStore) OverallScore(engineerID, cycleID int) (*float64, error) {
	var overall sql.NullFloat64
	err := s.db.QueryRow(
		`SELECT AVG(ma_scores.score) FROM (
			SELECT ma.id, AVG(sar.score) AS score
			FROM sub_attribute_rankings sar
			JOIN sub_attributes sa ON sa.id = sar.sub_attribute_id
			JOIN main_attributes ma ON ma.id = sa.main_attribute_id
			JOIN rating_cycles rc ON rc.id = sar.cycle_id
			WHERE sar.cycle_id = $1 AND sar.engineer_id = $2 AND ma.created_at <= rc.created_at
			GROUP BY ma.id
		 ) ma_scores`,
		cycleID, engineerID,
	).Scan(&overall)
	if err != nil {
		return nil, err
	}
	if !overall.Valid {
		return nil, nil
	}
	return &overall.Float64, nil
}
