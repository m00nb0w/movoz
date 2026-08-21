package syncer

import (
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		url = "postgres://localhost/scout_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", url)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping: test database not available: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

// githubCall records the arguments a single FetchPRStats invocation was made
// with, so tests can assert on the exact since/until boundaries the syncer
// computed rather than just trusting the code looks right.
type githubCall struct {
	username string
	repos    []string
	since    time.Time
	until    time.Time
}

// fakeGitHub is a GitHubStatsFetcher test double. It is safe for concurrent
// use (guarded by mu) because the scheduler tests exercise it from a
// background goroutine started by StartScheduler while the test goroutine
// reads back call counts.
type fakeGitHub struct {
	raised, reviewed int
	err              error

	mu    sync.Mutex
	calls []githubCall
}

func (f *fakeGitHub) FetchPRStats(ctx context.Context, username string, repos []string, since, until time.Time) (int, int, error) {
	f.mu.Lock()
	f.calls = append(f.calls, githubCall{username: username, repos: repos, since: since, until: until})
	f.mu.Unlock()
	return f.raised, f.reviewed, f.err
}

func (f *fakeGitHub) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func (f *fakeGitHub) call(i int) githubCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[i]
}

// jiraCall records the arguments a single FetchTicketStats invocation was
// made with (see githubCall).
type jiraCall struct {
	accountID string
	projects  []string
	since     time.Time
	until     time.Time
}

// fakeJira is a JiraStatsFetcher test double; see fakeGitHub for why it is
// mutex-guarded.
type fakeJira struct {
	closed     int
	complexity float64
	err        error

	mu    sync.Mutex
	calls []jiraCall
}

func (f *fakeJira) FetchTicketStats(ctx context.Context, accountID string, projects []string, since, until time.Time) (int, float64, error) {
	f.mu.Lock()
	f.calls = append(f.calls, jiraCall{accountID: accountID, projects: projects, since: since, until: until})
	f.mu.Unlock()
	return f.closed, f.complexity, f.err
}

func (f *fakeJira) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func (f *fakeJira) call(i int) jiraCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls[i]
}
