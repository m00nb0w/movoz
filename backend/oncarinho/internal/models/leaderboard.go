package models

type LeaderboardEntry struct {
	PlayerID   int     `json:"player_id"`
	PlayerName string  `json:"player_name"`
	Position   *string `json:"position"`
	IsActive   bool    `json:"is_active"`
	Value      int     `json:"value"`
}
