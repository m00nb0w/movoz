package models

type Summary struct {
	Year          int `json:"year"`
	MatchesPlayed int `json:"matches_played"`
	GoalsScored   int `json:"goals_scored"`
	RosterSize    int `json:"roster_size"`
}
