package models

import "time"

type MainAttributeScore struct {
	MainAttributeID int     `json:"main_attribute_id"`
	Key             string  `json:"key"`
	Name            string  `json:"name"`
	Score           float64 `json:"score"`
}

type EngineerCard struct {
	Engineer       Engineer             `json:"engineer"`
	CycleID        int                  `json:"cycle_id"`
	Overall        *float64             `json:"overall"`
	MainAttributes []MainAttributeScore `json:"main_attributes"`
}

type TrendPoint struct {
	CycleID        int                  `json:"cycle_id"`
	PeriodStart    time.Time            `json:"period_start"`
	PeriodEnd      time.Time            `json:"period_end"`
	Overall        *float64             `json:"overall"`
	MainAttributes []MainAttributeScore `json:"main_attributes"`
}

type EngineerCycleScore struct {
	Engineer       Engineer             `json:"engineer"`
	Overall        *float64             `json:"overall"`
	MainAttributes []MainAttributeScore `json:"main_attributes"`
}

type RosterEntry struct {
	Engineer      Engineer   `json:"engineer"`
	LatestOverall *float64   `json:"latest_overall"`
	LastCycleDate *time.Time `json:"last_cycle_date"`
}
