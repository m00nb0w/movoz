package jirasync

import (
	"testing"

	"el-storko/internal/models"
)

func TestJiraStatusToLocal(t *testing.T) {
	cases := []struct {
		jiraName string
		want     models.Status
	}{
		{"Done", models.StatusDone},
		{"Closed", models.StatusDone},
		{"Resolved", models.StatusDone},
		{"In Progress", models.StatusInProgress},
		{"Blocked", models.StatusBlocked},
		{"To Do", models.StatusBacklog},
		{"Some Unmapped Custom Status", models.StatusBacklog},
	}
	for _, c := range cases {
		if got := jiraStatusToLocal(c.jiraName); got != c.want {
			t.Errorf("jiraStatusToLocal(%q) = %q, want %q", c.jiraName, got, c.want)
		}
	}
}

func TestJiraStatusToLocal_NeverProducesPickedForToday(t *testing.T) {
	// picked_for_today is a personal daily-triage state with no Jira
	// equivalent (FR-005) — pull must never assign it, regardless of input.
	inputs := []string{"Done", "In Progress", "Blocked", "To Do", "Anything Else", ""}
	for _, in := range inputs {
		if got := jiraStatusToLocal(in); got == models.StatusPickedForToday {
			t.Errorf("jiraStatusToLocal(%q) produced picked_for_today, which pull must never do", in)
		}
	}
}

func TestLocalStatusToJiraTransitionName(t *testing.T) {
	cases := []struct {
		status models.Status
		want   string
	}{
		{models.StatusDone, "Done"},
		{models.StatusInProgress, "In Progress"},
		{models.StatusBlocked, "Blocked"},
		{models.StatusBacklog, "To Do"},
		{models.StatusPickedForToday, "In Progress"},
	}
	for _, c := range cases {
		if got := localStatusToJiraTransitionName(c.status); got != c.want {
			t.Errorf("localStatusToJiraTransitionName(%q) = %q, want %q", c.status, got, c.want)
		}
	}
}
