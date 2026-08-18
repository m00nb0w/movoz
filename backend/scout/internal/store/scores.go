package store

import (
	"database/sql"
	"time"

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
//
// created_at on main_attributes/rating_cycles is nullable at the schema
// level (DEFAULT NOW(), no NOT NULL constraint). SQL's NULL <= x evaluates
// to NULL/unknown (excluded from WHERE), so a NULL created_at would already
// fail safe by being silently excluded — but that reliance on implicit
// three-valued-logic semantics is fragile for future readers/writers, so
// the NULL check is made explicit here rather than left implicit.
func (s *ScoreStore) OverallScore(engineerID, cycleID int) (*float64, error) {
	var overall sql.NullFloat64
	err := s.db.QueryRow(
		`SELECT AVG(ma_scores.score) FROM (
			SELECT ma.id, AVG(sar.score) AS score
			FROM sub_attribute_rankings sar
			JOIN sub_attributes sa ON sa.id = sar.sub_attribute_id
			JOIN main_attributes ma ON ma.id = sa.main_attribute_id
			JOIN rating_cycles rc ON rc.id = sar.cycle_id
			WHERE sar.cycle_id = $1 AND sar.engineer_id = $2
			  AND ma.created_at IS NOT NULL AND rc.created_at IS NOT NULL
			  AND ma.created_at <= rc.created_at
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

// EngineerCard returns the engineer's Overall + main-attribute scores for
// one cycle (F10).
func (s *ScoreStore) EngineerCard(engineerStore *EngineerStore, engineerID, cycleID int) (*models.EngineerCard, error) {
	engineer, err := engineerStore.GetByID(engineerID)
	if err != nil {
		return nil, err
	}
	if engineer == nil {
		return nil, nil
	}

	mainScores, err := s.MainAttributeScores(engineerID, cycleID)
	if err != nil {
		return nil, err
	}
	overall, err := s.OverallScore(engineerID, cycleID)
	if err != nil {
		return nil, err
	}

	return &models.EngineerCard{
		Engineer:       *engineer,
		CycleID:        cycleID,
		Overall:        overall,
		MainAttributes: mainScores,
	}, nil
}

// EngineerTrend returns the engineer's scores across every past cycle they
// have at least one ranking in, oldest first (F10).
func (s *ScoreStore) EngineerTrend(engineerStore *EngineerStore, engineerID int) ([]models.TrendPoint, error) {
	rows, err := s.db.Query(
		`SELECT DISTINCT rc.id, rc.period_start, rc.period_end
		 FROM rating_cycles rc
		 JOIN sub_attribute_rankings sar ON sar.cycle_id = rc.id
		 WHERE sar.engineer_id = $1
		 ORDER BY rc.period_start`,
		engineerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type cycleRow struct {
		id          int
		periodStart time.Time
		periodEnd   time.Time
	}
	cycles := []cycleRow{}
	for rows.Next() {
		var cr cycleRow
		if err := rows.Scan(&cr.id, &cr.periodStart, &cr.periodEnd); err != nil {
			return nil, err
		}
		cycles = append(cycles, cr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	trend := make([]models.TrendPoint, 0, len(cycles))
	for _, cr := range cycles {
		mainScores, err := s.MainAttributeScores(engineerID, cr.id)
		if err != nil {
			return nil, err
		}
		overall, err := s.OverallScore(engineerID, cr.id)
		if err != nil {
			return nil, err
		}
		trend = append(trend, models.TrendPoint{
			CycleID:        cr.id,
			PeriodStart:    cr.periodStart,
			PeriodEnd:      cr.periodEnd,
			Overall:        overall,
			MainAttributes: mainScores,
		})
	}
	return trend, nil
}

// CycleScores returns every active engineer with their Overall + main-attribute
// scores as of the given cycle (F15) — so the team can be compared at a single
// point in time. Active engineers with no rankings for this cycle appear with
// nil Overall and empty MainAttributes. Deactivated engineers are excluded even
// if they have rankings in the cycle.
func (s *ScoreStore) CycleScores(engineerStore *EngineerStore, cycleID int) ([]models.EngineerCycleScore, error) {
	activeIDs, err := engineerStore.ListActiveIDs()
	if err != nil {
		return nil, err
	}

	results := make([]models.EngineerCycleScore, 0, len(activeIDs))
	for _, id := range activeIDs {
		engineer, err := engineerStore.GetByID(id)
		if err != nil {
			return nil, err
		}
		mainScores, err := s.MainAttributeScores(id, cycleID)
		if err != nil {
			return nil, err
		}
		overall, err := s.OverallScore(id, cycleID)
		if err != nil {
			return nil, err
		}
		results = append(results, models.EngineerCycleScore{
			Engineer:       *engineer,
			Overall:        overall,
			MainAttributes: mainScores,
		})
	}
	return results, nil
}
