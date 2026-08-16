export type Position = "goalkeeper" | "defender" | "midfielder" | "forward";

export interface Player {
  id: number;
  name: string;
  position: Position | null;
  is_active: boolean;
  created_at: string;
}

export interface Matchday {
  id: number;
  played_on: string; // "YYYY-MM-DD"
  created_at: string;
}

export interface MatchStat {
  id: number;
  matchday_id: number;
  player_id: number;
  goals: number;
  assists: number;
  yellow_cards: number;
  red_cards: number;
}

export interface StatInput {
  player_id: number;
  goals: number;
  assists: number;
  yellow_cards: number;
  red_cards: number;
}

export interface LeaderboardEntry {
  player_id: number;
  player_name: string;
  position: Position | null;
  is_active: boolean;
  value: number;
}

export interface StatTotals {
  matches_played: number;
  goals: number;
  assists: number;
  yellow_cards: number;
  red_cards: number;
}

export interface YearStats extends StatTotals {
  year: number;
}

export interface PlayerProfile {
  player: Player;
  all_time: StatTotals;
  by_year: YearStats[];
}

export interface Summary {
  year: number;
  matches_played: number;
  goals_scored: number;
  roster_size: number;
}
