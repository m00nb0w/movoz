package store

import (
	"database/sql"
	"encoding/json"

	"scout/internal/models"
)

type AISessionStore struct {
	db *sql.DB
}

func NewAISessionStore(db *sql.DB) *AISessionStore {
	return &AISessionStore{db: db}
}

func (s *AISessionStore) Create(cycleID, subAttributeID int) (*models.AIRankingSession, error) {
	var session models.AIRankingSession
	var proposedRankingStr sql.NullString
	err := s.db.QueryRow(
		`INSERT INTO ai_ranking_sessions (cycle_id, sub_attribute_id, transcript)
		 VALUES ($1, $2, '[]')
		 RETURNING id, cycle_id, sub_attribute_id, transcript, proposed_ranking, created_at`,
		cycleID, subAttributeID,
	).Scan(&session.ID, &session.CycleID, &session.SubAttributeID, &session.Transcript, &proposedRankingStr, &session.CreatedAt)
	if err != nil {
		return nil, err
	}
	if proposedRankingStr.Valid {
		session.ProposedRanking = json.RawMessage(proposedRankingStr.String)
	}
	return &session, nil
}

func (s *AISessionStore) GetByID(id int) (*models.AIRankingSession, error) {
	var session models.AIRankingSession
	var proposedRankingStr sql.NullString
	err := s.db.QueryRow(
		`SELECT id, cycle_id, sub_attribute_id, transcript, proposed_ranking, created_at
		 FROM ai_ranking_sessions WHERE id = $1`, id,
	).Scan(&session.ID, &session.CycleID, &session.SubAttributeID, &session.Transcript, &proposedRankingStr, &session.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if proposedRankingStr.Valid {
		session.ProposedRanking = json.RawMessage(proposedRankingStr.String)
	}
	return &session, nil
}

// UpdateTranscript replaces the session's full transcript (the chat history
// accumulated so far) and, when the assistant's latest reply included one,
// the proposed_ranking payload (F9). Persisting the ranking into
// sub_attribute_rankings only happens later, via the explicit accept
// endpoint (NF3) — this method never writes to sub_attribute_rankings.
func (s *AISessionStore) UpdateTranscript(id int, transcript json.RawMessage, proposedRanking json.RawMessage) error {
	_, err := s.db.Exec(
		"UPDATE ai_ranking_sessions SET transcript = $1, proposed_ranking = $2 WHERE id = $3",
		transcript, proposedRanking, id,
	)
	return err
}
