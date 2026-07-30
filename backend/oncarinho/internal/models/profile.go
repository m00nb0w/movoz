package models

type StatTotals struct {
	MatchesPlayed int `json:"matches_played"`
	Goals         int `json:"goals"`
	Assists       int `json:"assists"`
	YellowCards   int `json:"yellow_cards"`
	RedCards      int `json:"red_cards"`
}

type YearStats struct {
	Year int `json:"year"`
	StatTotals
}

type PlayerProfile struct {
	Player  Player      `json:"player"`
	AllTime StatTotals  `json:"all_time"`
	ByYear  []YearStats `json:"by_year"`
}
