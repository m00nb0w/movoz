package models

type MatchStat struct {
	ID          int `json:"id"`
	MatchdayID  int `json:"matchday_id"`
	PlayerID    int `json:"player_id"`
	Goals       int `json:"goals"`
	Assists     int `json:"assists"`
	YellowCards int `json:"yellow_cards"`
	RedCards    int `json:"red_cards"`
}

type StatInput struct {
	PlayerID    int `json:"player_id" binding:"required"`
	Goals       int `json:"goals" binding:"min=0"`
	Assists     int `json:"assists" binding:"min=0"`
	YellowCards int `json:"yellow_cards" binding:"min=0"`
	RedCards    int `json:"red_cards" binding:"min=0"`
}
