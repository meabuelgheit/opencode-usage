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

		CREATE TABLE IF NOT EXISTS message (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id INTEGER NOT NULL REFERENCES session(id),
			time_created INTEGER NOT NULL,
			data TEXT
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

func insertMessage(t *testing.T, db *sql.DB, sessionID int64, count int) {
	t.Helper()
	for i := 0; i < count; i++ {
		_, err := db.Exec(`INSERT INTO message (session_id, time_created, data) VALUES (?, ?, ?)`,
			sessionID, time.Now().UnixMilli(), "test message")
		if err != nil {
			t.Fatalf("failed to insert message: %v", err)
		}
	}
}

func TestGetSessions(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// We need session IDs. Insert one at a time and track IDs.
	var sessionIDs []int64
	for i := 0; i < 5; i++ {
		ts := now.Add(-time.Duration(i) * time.Hour).UnixMilli()
		slug := "s" + string(rune('a'+i))
		result, err := db.Exec(
			`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			ts, slug, "coder", "gpt-4", 100*int64(i+1), 50*int64(i+1), 10*int64(i+1), 5*int64(i+1), float64(i+1)*0.01, "/test/project")
		if err != nil {
			t.Fatalf("failed to insert session: %v", err)
		}
		id, _ := result.LastInsertId()
		sessionIDs = append(sessionIDs, id)
	}

	// Insert messages: 3 for first session, 1 for second, 0 for rest
	insertMessage(t, db, sessionIDs[0], 3)
	insertMessage(t, db, sessionIDs[1], 1)

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

	// Check TotalTokens = InputTokens + CacheRead
	for _, r := range rows {
		expectedTotal := r.InputTokens + r.CacheRead
		if r.TotalTokens != expectedTotal {
			t.Errorf("expected TotalTokens %d, got %d (input=%d, cache_read=%d)", expectedTotal, r.TotalTokens, r.InputTokens, r.CacheRead)
		}
	}

	// Check CacheReadPct is computed for rows with cache reads
	for _, r := range rows {
		if r.CacheRead > 0 && r.CacheReadPct <= 0 {
			t.Errorf("expected positive CacheReadPct for row with cache reads, got %.1f", r.CacheReadPct)
		}
	}

	// Check MessageCount: first row (newest, session[4]) has 0 msgs, second (session[3]) has 0, third (session[2]) has 0
	// Actually sessionIDs[4] is the most recent (i=0 is most recent because i=0 means 0 hours ago)
	// Wait: i=0 is now, i=1 is 1hr ago, i=2 is 2hr ago... i=4 is 4hr ago
	// ORDER BY time_created DESC means rows[0] = sessionIDs[0], rows[1] = sessionIDs[1], etc.
	// We added 3 msgs to sessionIDs[0] and 1 msg to sessionIDs[1]
	if len(rows) >= 1 && rows[0].MessageCount != 3 {
		t.Errorf("expected MessageCount 3 for first session, got %d", rows[0].MessageCount)
	}
	if len(rows) >= 2 && rows[1].MessageCount != 1 {
		t.Errorf("expected MessageCount 1 for second session, got %d", rows[1].MessageCount)
	}
	if len(rows) >= 3 && rows[2].MessageCount != 0 {
		t.Errorf("expected MessageCount 0 for third session, got %d", rows[2].MessageCount)
	}
}

func TestGetSessionsWithDays(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// We need the session IDs, so use db.Exec directly
	result, err := db.Exec(
		`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now.Add(-1*time.Hour).UnixMilli(), "recent", "coder", "gpt-4", 100, 50, 10, 5, 0.01, "/test")
	if err != nil {
		t.Fatalf("failed to insert session: %v", err)
	}
	recentID, _ := result.LastInsertId()

	result, err = db.Exec(
		`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		now.AddDate(0, 0, -20).UnixMilli(), "old", "coder", "gpt-4", 200, 100, 20, 10, 0.02, "/test")
	if err != nil {
		t.Fatalf("failed to insert session: %v", err)
	}
	oldID, _ := result.LastInsertId()

	// Add messages to both sessions
	insertMessage(t, db, recentID, 5)
	insertMessage(t, db, oldID, 2)

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

	if len(rows) > 0 && rows[0].MessageCount != 5 {
		t.Errorf("expected MessageCount 5 for recent session, got %d", rows[0].MessageCount)
	}
}

func TestGetDaily(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// Insert sessions across 3 days, track IDs
	// Day -2: 2 sessions
	r1, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.AddDate(0, 0, -2).Add(1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, "/test")
	id1, _ := r1.LastInsertId()
	r2, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.AddDate(0, 0, -2).Add(2*time.Hour).UnixMilli(), "s2", "architect", "claude-3", 200, 100, 20, 10, 0.02, "/test")
	_, _ = r2.LastInsertId()

	// Day -1: 1 session
	r3, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.AddDate(0, 0, -1).Add(1*time.Hour).UnixMilli(), "s3", "coder", "gpt-4", 300, 150, 30, 15, 0.03, "/test")
	id3, _ := r3.LastInsertId()

	// Day 0 (today): 1 session
	r4, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(1*time.Hour).UnixMilli(), "s4", "coder", "gpt-4", 400, 200, 40, 20, 0.04, "/test")
	_, _ = r4.LastInsertId()

	// Add messages: 3 to s1, 1 to s3
	insertMessage(t, db, id1, 3)
	insertMessage(t, db, id3, 1)

	rows, err := GetDaily(db, 7)
	if err != nil {
		t.Fatalf("GetDaily failed: %v", err)
	}

	// Should have 3 days
	if len(rows) != 3 {
		t.Errorf("expected 3 daily rows, got %d", len(rows))
	}

	// Check MessageCount across days
	for _, r := range rows {
		if r.Sessions == 2 {
			// That day should have 300 input tokens combined
			if r.InputTokens != 300 {
				t.Errorf("expected 300 input tokens for day with 2 sessions, got %d", r.InputTokens)
			}
			if r.Cost != 0.03 {
				t.Errorf("expected 0.03 cost for day with 2 sessions, got %.2f", r.Cost)
			}
			// Day -2: s1 has 3 messages, s2 has 0 messages => total 3
			if r.MessageCount != 3 {
				t.Errorf("expected MessageCount 3 for day with 2 sessions, got %d", r.MessageCount)
			}
		} else if r.Sessions == 1 {
			// Could be day -1 (s3 has 1 msg) or today (s4 has 0 msgs)
			// Check based on TotalTokens: day -1 has 300+30=330, today has 400+40=440
			if r.TotalTokens == 330 && r.MessageCount != 1 {
				t.Errorf("expected MessageCount 1 for day -1, got %d", r.MessageCount)
			}
			if r.TotalTokens == 440 && r.MessageCount != 0 {
				t.Errorf("expected MessageCount 0 for today, got %d", r.MessageCount)
			}
		}
		// Check TotalTokens = InputTokens + CacheRead
		expectedTotal := r.InputTokens + r.CacheRead
		if r.TotalTokens != expectedTotal {
			t.Errorf("expected TotalTokens %d, got %d (input=%d, cache_read=%d)", expectedTotal, r.TotalTokens, r.InputTokens, r.CacheRead)
		}

		// Check cache percentages are computed for days with cache data
		if r.CacheRead > 0 && r.CacheReadPct <= 0 {
			t.Errorf("expected positive CacheReadPct, got %.1f", r.CacheReadPct)
		}
		if r.CacheWrite > 0 && r.CacheWritePct <= 0 {
			t.Errorf("expected positive CacheWritePct, got %.1f", r.CacheWritePct)
		}
	}
}

func TestGetModels(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// 2 sessions with gpt-4
	r1, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, "/test")
	id1, _ := r1.LastInsertId()
	r2, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-2*time.Hour).UnixMilli(), "s2", "coder", "gpt-4", 200, 100, 20, 10, 0.02, "/test")
	_, _ = r2.LastInsertId()

	// 3 sessions with claude-3 (should be first, most sessions)
	r3, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-3*time.Hour).UnixMilli(), "s3", "architect", "claude-3", 300, 150, 30, 15, 0.03, "/test")
	id3, _ := r3.LastInsertId()
	r4, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-4*time.Hour).UnixMilli(), "s4", "architect", "claude-3", 400, 200, 40, 20, 0.04, "/test")
	_, _ = r4.LastInsertId()
	r5, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-5*time.Hour).UnixMilli(), "s5", "architect", "claude-3", 500, 250, 50, 25, 0.05, "/test")
	_, _ = r5.LastInsertId()

	// 1 session with null model (should be 'unknown')
	r6, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-6*time.Hour).UnixMilli(), "s6", "coder", nil, 50, 25, 5, 2, 0.005, "/test")
	id6, _ := r6.LastInsertId()

	// Add messages: 2 to gpt-4 s1, 4 to claude-3 s3, 1 to unknown s6
	insertMessage(t, db, id1, 2)
	insertMessage(t, db, id3, 4)
	insertMessage(t, db, id6, 1)

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

	// Check MessageCount
	for _, r := range rows {
		if r.Model == "claude-3" && r.MessageCount != 4 {
			t.Errorf("expected MessageCount 4 for claude-3, got %d", r.MessageCount)
		}
		if r.Model == "gpt-4" && r.MessageCount != 2 {
			t.Errorf("expected MessageCount 2 for gpt-4, got %d", r.MessageCount)
		}
	}

	// Check unknown model
	foundUnknown := false
	for _, r := range rows {
		if r.Model == "unknown" {
			foundUnknown = true
			if r.Sessions != 1 {
				t.Errorf("expected unknown model to have 1 session, got %d", r.Sessions)
			}
			if r.MessageCount != 1 {
				t.Errorf("expected MessageCount 1 for unknown model, got %d", r.MessageCount)
			}
		}
		// Check TotalTokens = InputTokens + CacheRead
		expectedTotal := r.InputTokens + r.CacheRead
		if r.TotalTokens != expectedTotal {
			t.Errorf("expected TotalTokens %d, got %d (input=%d, cache_read=%d)", expectedTotal, r.TotalTokens, r.InputTokens, r.CacheRead)
		}

		// Verify cache percentages are computed
		if r.CacheRead > 0 && r.CacheReadPct <= 0 {
			t.Errorf("expected positive CacheReadPct for model %s, got %.1f", r.Model, r.CacheReadPct)
		}
		if r.CacheWrite > 0 && r.CacheWritePct <= 0 {
			t.Errorf("expected positive CacheWritePct for model %s, got %.1f", r.Model, r.CacheWritePct)
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
	r1, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, "/test")
	id1, _ := r1.LastInsertId()
	_, _ = db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-2*time.Hour).UnixMilli(), "s2", "coder", "gpt-4", 200, 100, 20, 10, 0.02, "/test")

	// 3 sessions with architect
	r3, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-3*time.Hour).UnixMilli(), "s3", "architect", "claude-3", 300, 150, 30, 15, 0.03, "/test")
	id3, _ := r3.LastInsertId()
	_, _ = db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-4*time.Hour).UnixMilli(), "s4", "architect", "claude-3", 400, 200, 40, 20, 0.04, "/test")
	_, _ = db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-5*time.Hour).UnixMilli(), "s5", "architect", "claude-3", 500, 250, 50, 25, 0.05, "/test")

	// 1 session with null agent (should be 'unknown')
	r6, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-6*time.Hour).UnixMilli(), "s6", nil, "gpt-4", 50, 25, 5, 2, 0.005, "/test")
	id6, _ := r6.LastInsertId()

	// Add messages: 3 to coder s1, 5 to architect s3, 2 to unknown s6
	insertMessage(t, db, id1, 3)
	insertMessage(t, db, id3, 5)
	insertMessage(t, db, id6, 2)

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

	// Check MessageCount
	for _, r := range rows {
		if r.Agent == "architect" && r.MessageCount != 5 {
			t.Errorf("expected MessageCount 5 for architect, got %d", r.MessageCount)
		}
		if r.Agent == "coder" && r.MessageCount != 3 {
			t.Errorf("expected MessageCount 3 for coder, got %d", r.MessageCount)
		}
		// Check TotalTokens = InputTokens + CacheRead
		expectedTotal := r.InputTokens + r.CacheRead
		if r.TotalTokens != expectedTotal {
			t.Errorf("expected TotalTokens %d, got %d (input=%d, cache_read=%d)", expectedTotal, r.TotalTokens, r.InputTokens, r.CacheRead)
		}
		// Verify cache percentages are computed for rows with cache data
		if r.CacheRead > 0 && r.CacheReadPct <= 0 {
			t.Errorf("expected positive CacheReadPct for agent %s, got %.1f", r.Agent, r.CacheReadPct)
		}
		if r.CacheWrite > 0 && r.CacheWritePct <= 0 {
			t.Errorf("expected positive CacheWritePct for agent %s, got %.1f", r.Agent, r.CacheWritePct)
		}
	}

	// Check unknown agent
	foundUnknown := false
	for _, r := range rows {
		if r.Agent == "unknown" {
			foundUnknown = true
			if r.MessageCount != 2 {
				t.Errorf("expected MessageCount 2 for unknown agent, got %d", r.MessageCount)
			}
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

	// Insert sessions across 2 days, track IDs
	r1, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.AddDate(0, 0, -1).Add(1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, "/test")
	id1, _ := r1.LastInsertId()
	r2, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.AddDate(0, 0, -1).Add(2*time.Hour).UnixMilli(), "s2", "architect", "claude-3", 200, 100, 20, 10, 0.02, "/test")
	id2, _ := r2.LastInsertId()
	r3, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(1*time.Hour).UnixMilli(), "s3", "coder", "gpt-4", 300, 150, 30, 15, 0.03, "/test")
	id3, _ := r3.LastInsertId()

	// Add messages: 2 to s1, 3 to s2, 1 to s3
	insertMessage(t, db, id1, 2)
	insertMessage(t, db, id2, 3)
	insertMessage(t, db, id3, 1)

	summary, err := GetSummary(db, 0)
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

	if summary.MessageCount != 6 {
		t.Errorf("expected 6 total messages, got %d", summary.MessageCount)
	}

	if summary.CacheReadPct <= 0 {
		t.Errorf("expected positive CacheReadPct, got %.1f", summary.CacheReadPct)
	}
	if summary.CacheWritePct <= 0 {
		t.Errorf("expected positive CacheWritePct, got %.1f", summary.CacheWritePct)
	}

	// Verify CostPerMTokens: totalCost=0.06, totalBilledTokens=600+300=900
	// 0.06 / (900 / 1_000_000) = 0.06 / 0.0009 = 66.666...
	expectedCostPerM := 0.06 / (float64(900) / 1_000_000)
	if summary.CostPerMTokens != expectedCostPerM {
		t.Errorf("expected CostPerMTokens %.2f, got %.2f", expectedCostPerM, summary.CostPerMTokens)
	}

	// Verify EffectiveCostPerMTokens: totalCost=0.06, effectiveTokens=600+300+60=960
	// 0.06 / (960 / 1_000_000) = 0.06 / 0.00096 = 62.50
	expectedEffectiveCostPerM := 0.06 / (float64(960) / 1_000_000)
	if summary.EffectiveCostPerMTokens != expectedEffectiveCostPerM {
		t.Errorf("expected EffectiveCostPerMTokens %.2f, got %.2f", expectedEffectiveCostPerM, summary.EffectiveCostPerMTokens)
	}

	// Verify specific values: total cache read = 60, total input = 600, total cache write = 30
	// CacheReadPct = 60 / (600 + 60) * 100 = 60/660*100 ≈ 9.1%
	expectedReadPct := float64(60) / float64(600+60) * 100
	if summary.CacheReadPct != expectedReadPct {
		t.Errorf("expected CacheReadPct %.1f, got %.1f", expectedReadPct, summary.CacheReadPct)
	}
	// CacheWritePct = 30 / (600 + 60) * 100 = 30/660*100 ≈ 4.5%
	expectedWritePct := float64(30) / float64(600+60) * 100
	if summary.CacheWritePct != expectedWritePct {
		t.Errorf("expected CacheWritePct %.1f, got %.1f", expectedWritePct, summary.CacheWritePct)
	}
}

func TestGetSummary_EmptyDB(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	summary, err := GetSummary(db, 0)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}

	if summary.TotalSessions != 0 {
		t.Errorf("expected 0 total sessions, got %d", summary.TotalSessions)
	}
	if summary.ActiveDays != 0 {
		t.Errorf("expected 0 active days, got %d", summary.ActiveDays)
	}
	if summary.MessageCount != 0 {
		t.Errorf("expected 0 MessageCount for empty DB, got %d", summary.MessageCount)
	}
	if summary.CacheReadPct != 0 {
		t.Errorf("expected 0 CacheReadPct for empty DB, got %.1f", summary.CacheReadPct)
	}
	if summary.CacheWritePct != 0 {
		t.Errorf("expected 0 CacheWritePct for empty DB, got %.1f", summary.CacheWritePct)
	}
	if summary.CostPerMTokens != 0 {
		t.Errorf("expected 0 CostPerMTokens for empty DB, got %.2f", summary.CostPerMTokens)
	}
	if summary.EffectiveCostPerMTokens != 0 {
		t.Errorf("expected 0 EffectiveCostPerMTokens for empty DB, got %.2f", summary.EffectiveCostPerMTokens)
	}
}

func TestGetSummaryWithDays(t *testing.T) {
	db := createTestDB(t)
	defer db.Close()

	now := time.Now()

	// Insert a recent session (1 hour ago)
	r1, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-1*time.Hour).UnixMilli(), "recent", "coder", "gpt-4", 100, 50, 10, 5, 0.01, "/test")
	id1, _ := r1.LastInsertId()

	// Insert an old session (20 days ago)
	r2, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.AddDate(0, 0, -20).UnixMilli(), "old", "architect", "claude-3", 200, 100, 20, 10, 0.02, "/test")
	id2, _ := r2.LastInsertId()

	// Add messages to both
	insertMessage(t, db, id1, 5)
	insertMessage(t, db, id2, 3)

	// Query with days=7 — only recent session should be included
	summary, err := GetSummary(db, 7)
	if err != nil {
		t.Fatalf("GetSummary failed: %v", err)
	}

	if summary.TotalSessions != 1 {
		t.Errorf("expected 1 session within 7 days, got %d", summary.TotalSessions)
	}

	if summary.ActiveDays != 1 {
		t.Errorf("expected 1 active day within 7 days, got %d", summary.ActiveDays)
	}

	expectedInput := int64(100)
	if summary.TotalInputTokens != expectedInput {
		t.Errorf("expected %d input tokens within 7 days, got %d", expectedInput, summary.TotalInputTokens)
	}

	expectedOutput := int64(50)
	if summary.TotalOutputTokens != expectedOutput {
		t.Errorf("expected %d output tokens within 7 days, got %d", expectedOutput, summary.TotalOutputTokens)
	}

	expectedCost := 0.01
	if summary.TotalCost != expectedCost {
		t.Errorf("expected %.2f cost within 7 days, got %.2f", expectedCost, summary.TotalCost)
	}

	if summary.MessageCount != 5 {
		t.Errorf("expected 5 messages within 7 days, got %d", summary.MessageCount)
	}

	// Cache percentages should still be computed
	if summary.CacheReadPct <= 0 {
		t.Errorf("expected positive CacheReadPct, got %.1f", summary.CacheReadPct)
	}
	if summary.CacheWritePct <= 0 {
		t.Errorf("expected positive CacheWritePct, got %.1f", summary.CacheWritePct)
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
	r1, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, project_id, directory) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, &proj1IDPtr, "/home/user/my-project")
	id1, _ := r1.LastInsertId()
	_, _ = db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, project_id, directory) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-2*time.Hour).UnixMilli(), "s2", "coder", "gpt-4", 200, 100, 20, 10, 0.02, &proj1IDPtr, "/home/user/my-project")

	// 3 sessions for proj2
	r3, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, project_id, directory) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-3*time.Hour).UnixMilli(), "s3", "architect", "claude-3", 300, 150, 30, 15, 0.03, &proj2IDPtr, "/home/user/another-project")
	id3, _ := r3.LastInsertId()
	_, _ = db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, project_id, directory) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-4*time.Hour).UnixMilli(), "s4", "architect", "claude-3", 400, 200, 40, 20, 0.04, &proj2IDPtr, "/home/user/another-project")
	_, _ = db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, project_id, directory) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-5*time.Hour).UnixMilli(), "s5", "architect", "claude-3", 500, 250, 50, 25, 0.05, &proj2IDPtr, "/home/user/another-project")

	// 1 session with no project_id (should use directory)
	r6, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, directory) VALUES (?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-6*time.Hour).UnixMilli(), "s6", "coder", "gpt-4", 50, 25, 5, 2, 0.005, "/some/other/dir")
	id6, _ := r6.LastInsertId()

	// Add messages: 1 to s1, 2 to s3, 3 to s6
	insertMessage(t, db, id1, 1)
	insertMessage(t, db, id3, 2)
	insertMessage(t, db, id6, 3)

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

	// Check MessageCount
	if rows[0].Project == "another-project" && rows[0].MessageCount != 2 {
		t.Errorf("expected MessageCount 2 for another-project, got %d", rows[0].MessageCount)
	}

	// Check the unlinked session uses directory
	foundDir := false
	for _, r := range rows {
		if r.Project == "/some/other/dir" {
			foundDir = true
			if r.Sessions != 1 {
				t.Errorf("expected directory project to have 1 session, got %d", r.Sessions)
			}
			if r.MessageCount != 3 {
				t.Errorf("expected MessageCount 3 for directory project, got %d", r.MessageCount)
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
	r1, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, project_id, directory) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		now.Add(-1*time.Hour).UnixMilli(), "s1", "coder", "gpt-4", 100, 50, 10, 5, 0.01, &projIDPtr, "/test")
	id1, _ := r1.LastInsertId()

	// Old session (20 days ago)
	r2, _ := db.Exec(`INSERT INTO session (time_created, slug, agent, model, tokens_input, tokens_output, tokens_cache_read, tokens_cache_write, cost, project_id, directory) VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		now.AddDate(0, 0, -20).UnixMilli(), "s2", "coder", "gpt-4", 200, 100, 20, 10, 0.02, &projIDPtr, "/test")
	id2, _ := r2.LastInsertId()

	// Add messages
	insertMessage(t, db, id1, 7)
	insertMessage(t, db, id2, 1)

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

	if rows[0].MessageCount != 7 {
		t.Errorf("expected MessageCount 7 for project in last 7 days, got %d", rows[0].MessageCount)
	}
}

func TestDefaultDBPath(t *testing.T) {
	path := DefaultDBPath()
	expected := "~/.local/share/opencode/opencode.db"
	if path != expected {
		t.Errorf("expected %s, got %s", expected, path)
	}
}
