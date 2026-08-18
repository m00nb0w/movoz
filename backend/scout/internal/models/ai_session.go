package models

import (
	"encoding/json"
	"time"
)

type AIRankingSession struct {
	ID              int             `json:"id"`
	CycleID         int             `json:"cycle_id"`
	SubAttributeID  int             `json:"sub_attribute_id"`
	Transcript      json.RawMessage `json:"transcript"`
	ProposedRanking json.RawMessage `json:"proposed_ranking"`
	CreatedAt       time.Time       `json:"created_at"`
}
