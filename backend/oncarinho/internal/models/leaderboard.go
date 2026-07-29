package models

type LeaderboardEntry struct {
	PlayerID   int    `json:"player_id"`
	PlayerName string `json:"player_name"`
	Value      int    `json:"value"`
}
