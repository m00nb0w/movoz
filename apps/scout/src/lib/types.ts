export interface Engineer {
  id: number;
  name: string;
  role: string | null;
  github_username: string | null;
  jira_account_id: string | null;
  started_at: string;
  is_active: boolean;
  created_at: string;
}

export interface MainAttribute {
  id: number;
  key: string;
  name: string;
  created_at: string;
}

export interface SubAttribute {
  id: number;
  main_attribute_id: number;
  name: string;
  description: string | null;
  is_active: boolean;
  created_at: string;
}

export interface RatingCycle {
  id: number;
  period_start: string;
  period_end: string;
  created_at: string;
}

export interface SubAttributeRanking {
  id: number;
  cycle_id: number;
  sub_attribute_id: number;
  engineer_id: number;
  rank: number;
  score: number;
}

export interface MainAttributeScore {
  main_attribute_id: number;
  key: string;
  name: string;
  score: number;
}

export interface EngineerCard {
  engineer: Engineer;
  cycle_id: number;
  overall: number | null;
  main_attributes: MainAttributeScore[];
}

export interface TrendPoint {
  cycle_id: number;
  period_start: string;
  period_end: string;
  overall: number | null;
  main_attributes: MainAttributeScore[];
}

export interface EngineerCycleScore {
  engineer: Engineer;
  overall: number | null;
  main_attributes: MainAttributeScore[];
}

export interface RosterEntry {
  engineer: Engineer;
  latest_overall: number | null;
  last_cycle_date: string | null;
}

export interface MetricSnapshot {
  id: number;
  engineer_id: number;
  period_start: string;
  period_end: string;
  prs_raised: number;
  prs_reviewed: number;
  tickets_closed: number;
  complexity_score: number;
  synced_at: string;
}

export interface HighlightEntry {
  id: number;
  engineer_id: number;
  kind: "highlight" | "lowlight";
  body: string;
  created_at: string;
}
