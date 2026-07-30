package stats

import (
	"database/sql"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

// createTestDB creates an in-memory SQLite database with the session and project tables.
func createTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	// Create tables matching opencode schema
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS project (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			directory TEXT NOT NULL
		);

		CREATE TABLE IF NOT EXISTS session (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			time_created INTEGER NOT NULL,
			slug TEXT NOT NULL,
			agent TEXT,
			model TEXT,
			tokens_input INTEGER DEFAULT 0,
			tokens_output INTEGER DEFAULT 0,
			tokens_cache_read INTEGER DEFAULT 0,
			tokens_cache_write INTEGER DEFAULT 0,
			cost REAL DEFAULT 0,
			project_id INTEGER REFERENCES project(id),
			directory TEXT
		);
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	return db
}

func insertSession(t *testing.T, db *sql.DB, ts int64, slug string, agent, model interface{}, tin, tout, tcr, tcw int64, cost float64, projectID *int, directory string) {
	t.Helper()

	var err error
	if projectID != nil {
		_, err = db.Exec(
			`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, project_id, directory)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			ts, slug, agent, model, tin, tout, tcr, tcw, cost, *projectID, directory)
	} else {
		_, err = db.Exec(
			`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			ts, slug, agent, model, tin, tout, tcr, tcw, cost, directory)
	}
	if err != nil {
		t.Fatalf("failed to insert session: %v", err)
	}
}

func insertProject(t *testing.T, db *sql.DB, name, directory string) int64 {
	t.Helper()

	result, err := db.Exec(
		`INSERT INTO project (name, directory) VALUES (?, ?)`,
		name, directory)
	if err != nil {
		t.Fatalf("failed to insert project: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

func TestGetSessions(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()
	for i := 0; i < 5; i++ {
		ts := now.Add(-time.Duration(i) * time.Hour).UnixMilli()
		insertSession(t, db, ts, "session-"+string(rune('a'+i)), "coder", "gpt-4", 100*int64(i+1), 50*int64(i+1), 10*int64(i+1), 5*int64(i+1), float64(i+1)*0.01, nil, "/test/project")
	}

	rows, err := GetSessions(db, 3, 0)
	if err != nil {
		t.Fatalf("GetSessions failed: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("expected 3 rows, got %d", len(rows))
	}

	// Verify ordering (DESC by time_created)
	if len(rows) >= 2 {
		if rows[0].Date < rows[1].Date {
			t.Errorf("expected descending order by date, got %s before %s", rows[0].Date, rows[1].Date)
		}
	}
}

func TestGetSessionsWithDays(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// Insert a recent session
	insertSession(t, db, now.Add(-1*time.Hour).UnixMilli(), "recent", "coder", "gpt-4", 100, 50, 10, 5, 0.01, nil, "/test")

	// Insert an old session (20 days ago)
	insertSession(t, db, now.AddDate(0, 0, -20).UnixMilli(), "old", "coder", "gpt-4", 200, 100, 20, 10, 0.02, nil, "/test")

	// Query with days=7 should only return the recent one
	rows, err := GetSessions(db, 10, 7)
	if err != nil {
		t.Fatalf("GetSessions failed: %v", err)
	}

	if len(rows) != 1 {
		t.Errorf("expected 1 session within 7 days, got %d", len(rows))
	}

	if len(rows) > 0 && rows[0].Slug != "recent" {
		t.Errorf("expected 'recent' session, got %s", rows[0].Slug)
	}
}

func TestGetDaily(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// Insert sessions across 3 days
	// Day -2: 2 sessions
	insertSession(t, db, now.AddDate(0, 0, -2).Add(1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, nil, "/test")
	insertSession(t, db, now.AddDate(0, 0, -2).Add(2*time.Hour).UnixMilli(), "s2", "architect", "claude-3", 200, 100, 20, 10, 0.02, nil, "/test")

	// Day -1: 1 session
	insertSession(t, db, now.AddDate(0, 0, -1).Add(1*time.Hour).UnixMilli(), "s3", "coder", "gpt-4", 300, 150, 30, 15, 0.03, nil, "/test")

	// Day 0 (today): 1 session
	insertSession(t, db, now.Add(1*time.Hour).UnixMilli(), "s4", "coder", "gpt-4", 400, 200, 40, 20, 0.04, nil, "/test")

	rows, err := GetDaily(db, 7)
	if err != nil {
		t.Fatalf("GetDaily failed: %v", err)
	}

	// Should have 3 days
	if len(rows) != 3 {
		t.Errorf("expected 3 daily rows, got %d", len(rows))
	}

	// Find the day with 2 sessions
	for _, r := range rows {
		if r.Sessions == 2 {
			// That day should have 300 input tokens combined
			if r.InputTokens != 300 {
				t.Errorf("expected 300 input tokens for day with 2 sessions, got %d", r.InputTokens)
			}
			if r.Cost != 0.03 {
				t.Errorf("expected 0.03 cost for day with 2 sessions, got %.2f", r.Cost)
			}
		}
	}
}

func TestGetModels(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// 2 sessions with gpt-4
	insertSession(t, db, now.Add(-1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, nil, "/test")
	insertSession(t, db, now.Add(-2*time.Hour).UnixMilli(), "s2", "coder", "gpt-4", 200, 100, 20, 10, 0.02, nil, "/test")

	// 3 sessions with claude-3 (should be first, most sessions)
	insertSession(t, db, now.Add(-3*time.Hour).UnixMilli(), "s3", "architect", "claude-3", 300, 150, 30, 15, 0.03, nil, "/test")
	insertSession(t, db, now.Add(-4*time.Hour).UnixMilli(), "s4", "architect", "claude-3", 400, 200, 40, 20, 0.04, nil, "/test")
	insertSession(t, db, now.Add(-5*time.Hour).UnixMilli(), "s5", "architect", "claude-3", 500, 250, 50, 25, 0.05, nil, "/test")

	// 1 session with null model (should be 'unknown')
	insertSession(t, db, now.Add(-6*time.Hour).UnixMilli(), "s6", "coder", nil, 50, 25, 5, 2, 0.005, nil, "/test")

	rows, err := GetModels(db, 0)
	if err != nil {
		t.Fatalf("GetModels failed: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("expected 3 model rows, got %d", len(rows))
	}

	// claude-3 should be first (most sessions)
	if rows[0].Model != "claude-3" {
		t.Errorf("expected claude-3 as first model, got %s", rows[0].Model)
	}

	if rows[0].Sessions != 3 {
		t.Errorf("expected claude-3 to have 3 sessions, got %d", rows[0].Sessions)
	}

	// Check unknown model
	foundUnknown := false
	for _, r := range rows {
		if r.Model == "unknown" {
			foundUnknown = true
			if r.Sessions != 1 {
				t.Errorf("expected unknown model to have 1 session, got %d", r.Sessions)
			}
		}
	}
	if !foundUnknown {
		t.Error("expected unknown model row for empty model name")
	}
}

func TestGetAgents(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// 2 sessions with coder
	insertSession(t, db, now.Add(-1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, nil, "/test")
	insertSession(t, db, now.Add(-2*time.Hour).UnixMilli(), "s2", "coder", "gpt-4", 200, 100, 20, 10, 0.02, nil, "/test")

	// 3 sessions with architect
	insertSession(t, db, now.Add(-3*time.Hour).UnixMilli(), "s3", "architect", "claude-3", 300, 150, 30, 15, 0.03, nil, "/test")
	insertSession(t, db, now.Add(-4*time.Hour).UnixMilli(), "s4", "architect", "claude-3", 400, 200, 40, 20, 0.04, nil, "/test")
	insertSession(t, db, now.Add(-5*time.Hour).UnixMilli(), "s5", "architect", "claude-3", 500, 250, 50, 25, 0.05, nil, "/test")

	// 1 session with null agent (should be 'unknown')
	insertSession(t, db, now.Add(-6*time.Hour).UnixMilli(), "s6", nil, "gpt-4", 50, 25, 5, 2, 0.005, nil, "/test")

	rows, err := GetAgents(db, 0)
	if err != nil {
		t.Fatalf("GetAgents failed: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("expected 3 agent rows, got %d", len(rows))
	}

	// architect should be first (most sessions)
	if rows[0].Agent != "architect" {
		t.Errorf("expected architect as first agent, got %s", rows[0].Agent)
	}

	if rows[0].Sessions != 3 {
		t.Errorf("expected architect to have 3 sessions, got %d", rows[0].Sessions)
	}

	// Check unknown agent
	foundUnknown := false
	for _, r := range rows {
		if r.Agent == "unknown" {
			foundUnknown = true
		}
	}
	if !foundUnknown {
		t.Error("expected unknown agent row for empty agent name")
	}
}

func TestGetSummary(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// Insert sessions across 2 days
	insertSession(t, db, now.AddDate(0, 0, -1).Add(1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, nil, "/test")
	insertSession(t, db, now.AddDate(0, 0, -1).Add(2*time.Hour).UnixMilli(), "s2", "architect", "claude-3", 200, 100, 20, 10, 0.02, nil, "/test")
	insertSession(t, db, now.Add(1*time.Hour).UnixMilli(), "s3", "coder", "gpt-4", 300, 150, 30, 15, 0.03, nil, "/test")

	summary, err := GetSummary(db)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}

	if summary.TotalSessions != 3 {
		t.Errorf("expected 3 total sessions, got %d", summary.TotalSessions)
	}

	if summary.ActiveDays != 2 {
		t.Errorf("expected 2 active days, got %d", summary.ActiveDays)
	}

	expectedInput := int64(600)
	if summary.TotalInputTokens != expectedInput {
		t.Errorf("expected %d total input tokens, got %d", expectedInput, summary.TotalInputTokens)
	}

	expectedOutput := int64(300)
	if summary.TotalOutputTokens != expectedOutput {
		t.Errorf("expected %d total output tokens, got %d", expectedOutput, summary.TotalOutputTokens)
	}

	expectedCost := 0.06
	if summary.TotalCost != expectedCost {
		t.Errorf("expected %.2f total cost, got %.2f", expectedCost, summary.TotalCost)
	}
}

func TestGetSummary_EmptyDB(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	summary, err := GetSummary(db)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}

	if summary.TotalSessions != 0 {
		t.Errorf("expected 0 total sessions, got %d", summary.TotalSessions)
	}
	if summary.ActiveDays != 0 {
		t.Errorf("expected 0 active days, got %d", summary.ActiveDays)
	}
}

func TestGetProjects(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// Create projects
	proj1ID := insertProject(t, db, "my-project", "/home/user/my-project")
	proj2ID := insertProject(t, db, "another-project", "/home/user/another-project")

	proj1IDPtr := int(proj1ID)
	proj2IDPtr := int(proj2ID)

	// 2 sessions for proj1
	insertSession(t, db, now.Add(-1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, &proj1IDPtr, "/home/user/my-project")
	insertSession(t, db, now.Add(-2*time.Hour).UnixMilli(), "s2", "coder", "gpt-4", 200, 100, 20, 10, 0.02, &proj1IDPtr, "/home/user/my-project")

	// 3 sessions for proj2
	insertSession(t, db, now.Add(-3*time.Hour).UnixMilli(), "s3", "architect", "claude-3", 300, 150, 30, 15, 0.03, &proj2IDPtr, "/home/user/another-project")
	insertSession(t, db, now.Add(-4*time.Hour).UnixMilli(), "s4", "architect", "claude-3", 400, 200, 40, 20, 0.04, &proj2IDPtr, "/home/user/another-project")
	insertSession(t, db, now.Add(-5*time.Hour).UnixMilli(), "s5", "architect", "claude-3", 500, 250, 50, 25, 0.05, &proj2IDPtr, "/home/user/another-project")

	// 1 session with no project_id (should use directory)
	insertSession(t, db, now.Add(-6*time.Hour).UnixMilli(), "s6", "coder", "gpt-4", 50, 25, 5, 2, 0.005, nil, "/some/other/dir")

	rows, err := GetProjects(db, 0)
	if err != nil {
		t.Fatalf("GetProjects failed: %v", err)
	}

	if len(rows) != 3 {
		t.Errorf("expected 3 project rows, got %d", len(rows))
	}

	// another-project should be first (most sessions)
	if rows[0].Project != "another-project" {
		t.Errorf("expected another-project as first, got %s", rows[0].Project)
	}

	if rows[0].Sessions != 3 {
		t.Errorf("expected another-project to have 3 sessions, got %d", rows[0].Sessions)
	}

	// Check the unlinked session uses directory
	foundDir := false
	for _, r := range rows {
		if r.Project == "/some/other/dir" {
			foundDir = true
			if r.Sessions != 1 {
				t.Errorf("expected directory project to have 1 session, got %d", r.Sessions)
			}
		}
	}
	if !foundDir {
		t.Error("expected a project row using directory path for unlinked session")
	}
}

func TestGetProjectsWithDays(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	projID := insertProject(t, db, "my-project", "/test")
	projIDPtr := int(projID)

	// Recent session
	insertSession(t, db, now.Add(-1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, &projIDPtr, "/test")

	// Old session (20 days ago)
	insertSession(t, db, now.AddDate(0, 0, -20).UnixMilli(), "s2", "coder", "gpt-4", 200, 100, 20, 10, 0.02, &projIDPtr, "/test")

	rows, err := GetProjects(db, 7)
	if err != nil {
		t.Fatalf("GetProjects failed: %v", err)
	}

	if len(rows) != 1 {
		t.Errorf("expected 1 project in last 7 days, got %d", len(rows))
	}

	if rows[0].Sessions != 1 {
		t.Errorf("expected 1 session for project in last 7 days, got %d", rows[0].Sessions)
	}
}

func TestDefaultDBPath(t *testing.T) {
	path := DefaultDBPath()
	expected := "~/.local/share/opencode/opencode.db"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}
