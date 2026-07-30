package stats

import (
	"database/sql"
	"fmt"
	"time"
)

// DefaultDBPath returns the default opencode database path.
func DefaultDBPath() string {
	return "~/.local/share/opencode/opencode.db"
}

// GetSessions returns recent sessions, ordered by time_created DESC.
// limit: max rows to return. daysAgo: only include sessions from last N days (0 = all time).
func GetSessions(db *sql.DB, limit int, daysAgo int) ([]SessionRow, error) {
	var rows *sql.Rows
	var err error

	if daysAgo > 0 {
		since := time.Now().AddDate(0, 0, -daysAgo).UnixMilli()
		rows, err = db.Query(`
			SELECT s.time_created, s.slug, COALESCE(s.agent, 'unknown'), COALESCE(s.model, 'unknown'),
			       s.tokens_input, s.tokens_output, s.tokens_cache_read, s.cost,
			       (SELECT COUNT(*) FROM message WHERE message.session_id = s.id) as msg_count
			FROM session s
			WHERE s.time_created >= ?
			ORDER BY s.time_created DESC
			LIMIT ?
		`, since, limit)
	} else {
		rows, err = db.Query(`
			SELECT s.time_created, s.slug, COALESCE(s.agent, 'unknown'), COALESCE(s.model, 'unknown'),
			       s.tokens_input, s.tokens_output, s.tokens_cache_read, s.cost,
			       (SELECT COUNT(*) FROM message WHERE message.session_id = s.id) as msg_count
			FROM session s
			ORDER BY s.time_created DESC
			LIMIT ?
		`, limit)
	}
	if err != nil {
		return nil, fmt.Errorf("query sessions: %w", err)
	}
	defer rows.Close()

	var results []SessionRow
	for rows.Next() {
		var r SessionRow
		var ts int64
		if err := rows.Scan(&ts, &r.Slug, &r.Agent, &r.Model,
			&r.InputTokens, &r.OutputTokens, &r.CacheRead, &r.Cost, &r.MessageCount); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		r.Date = time.UnixMilli(ts).Format("2006-01-02")
		results = append(results, r)
	}
	for i := range results {
		results[i].TotalTokens = results[i].InputTokens + results[i].CacheRead
		if results[i].TotalTokens > 0 {
			results[i].CacheReadPct = float64(results[i].CacheRead) / float64(results[i].TotalTokens) * 100
		}
	}
	return results, rows.Err()
}

// GetDaily returns daily aggregated stats.
func GetDaily(db *sql.DB, daysAgo int) ([]DailyRow, error) {
	since := time.Now().AddDate(0, 0, -daysAgo).UnixMilli()
	rows, err := db.Query(`
		SELECT date(s.time_created/1000, 'unixepoch') as day,
		       COUNT(*) as sessions,
		       SUM(s.tokens_input),
		       SUM(s.tokens_output),
		       SUM(s.tokens_cache_read),
		       SUM(s.tokens_cache_write),
		       SUM(s.cost),
		       SUM((SELECT COUNT(*) FROM message WHERE message.session_id = s.id)) as msg_count
		FROM session s
		WHERE s.time_created >= ?
		GROUP BY day
		ORDER BY day DESC
	`, since)
	if err != nil {
		return nil, fmt.Errorf("query daily: %w", err)
	}
	defer rows.Close()

	var results []DailyRow
	for rows.Next() {
		var r DailyRow
		if err := rows.Scan(&r.Date, &r.Sessions, &r.InputTokens, &r.OutputTokens,
			&r.CacheRead, &r.CacheWrite, &r.Cost, &r.MessageCount); err != nil {
			return nil, fmt.Errorf("scan daily: %w", err)
		}
		results = append(results, r)
	}
	for i := range results {
		results[i].TotalTokens = results[i].InputTokens + results[i].CacheRead
		if results[i].TotalTokens > 0 {
			results[i].CacheReadPct = float64(results[i].CacheRead) / float64(results[i].TotalTokens) * 100
			results[i].CacheWritePct = float64(results[i].CacheWrite) / float64(results[i].TotalTokens) * 100
		}
	}
	return results, rows.Err()
}

// GetModels returns breakdown by model.
func GetModels(db *sql.DB, daysAgo int) ([]ModelRow, error) {
	var rows *sql.Rows
	var err error

	if daysAgo > 0 {
		since := time.Now().AddDate(0, 0, -daysAgo).UnixMilli()
		rows, err = db.Query(`
			SELECT COALESCE(s.model, 'unknown'),
			       COUNT(*) as sessions,
			       SUM(s.tokens_input), SUM(s.tokens_output),
			       SUM(s.tokens_cache_read), SUM(s.tokens_cache_write),
			       SUM(s.cost),
			       SUM((SELECT COUNT(*) FROM message WHERE message.session_id = s.id)) as msg_count
			FROM session s
			WHERE s.time_created >= ?
			GROUP BY s.model
			ORDER BY sessions DESC
		`, since)
	} else {
		rows, err = db.Query(`
			SELECT COALESCE(s.model, 'unknown'),
			       COUNT(*) as sessions,
			       SUM(s.tokens_input), SUM(s.tokens_output),
			       SUM(s.tokens_cache_read), SUM(s.tokens_cache_write),
			       SUM(s.cost),
			       SUM((SELECT COUNT(*) FROM message WHERE message.session_id = s.id)) as msg_count
			FROM session s
			GROUP BY s.model
			ORDER BY sessions DESC
		`)
	}
	if err != nil {
		return nil, fmt.Errorf("query models: %w", err)
	}
	defer rows.Close()

	var results []ModelRow
	for rows.Next() {
		var r ModelRow
		if err := rows.Scan(&r.Model, &r.Sessions, &r.InputTokens, &r.OutputTokens,
			&r.CacheRead, &r.CacheWrite, &r.Cost, &r.MessageCount); err != nil {
			return nil, fmt.Errorf("scan models: %w", err)
		}
		results = append(results, r)
	}
	for i := range results {
		results[i].TotalTokens = results[i].InputTokens + results[i].CacheRead
		if results[i].TotalTokens > 0 {
			results[i].CacheReadPct = float64(results[i].CacheRead) / float64(results[i].TotalTokens) * 100
			results[i].CacheWritePct = float64(results[i].CacheWrite) / float64(results[i].TotalTokens) * 100
		}
	}
	return results, rows.Err()
}

// GetAgents returns breakdown by agent type.
func GetAgents(db *sql.DB, daysAgo int) ([]AgentRow, error) {
	var rows *sql.Rows
	var err error

	if daysAgo > 0 {
		since := time.Now().AddDate(0, 0, -daysAgo).UnixMilli()
		rows, err = db.Query(`
			SELECT COALESCE(s.agent, 'unknown'),
			       COUNT(*) as sessions,
			       SUM(s.tokens_input), SUM(s.tokens_output),
			       SUM(s.cost),
			       SUM((SELECT COUNT(*) FROM message WHERE message.session_id = s.id)) as msg_count
			FROM session s
			WHERE s.time_created >= ?
			GROUP BY s.agent
			ORDER BY sessions DESC
		`, since)
	} else {
		rows, err = db.Query(`
			SELECT COALESCE(s.agent, 'unknown'),
			       COUNT(*) as sessions,
			       SUM(s.tokens_input), SUM(s.tokens_output),
			       SUM(s.cost),
			       SUM((SELECT COUNT(*) FROM message WHERE message.session_id = s.id)) as msg_count
			FROM session s
			GROUP BY s.agent
			ORDER BY sessions DESC
		`)
	}
	if err != nil {
		return nil, fmt.Errorf("query agents: %w", err)
	}
	defer rows.Close()

	var results []AgentRow
	for rows.Next() {
		var r AgentRow
		if err := rows.Scan(&r.Agent, &r.Sessions, &r.InputTokens, &r.OutputTokens, &r.Cost, &r.MessageCount); err != nil {
			return nil, fmt.Errorf("scan agents: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// GetSummary returns all-time summary statistics.
func GetSummary(db *sql.DB) (*SummaryRow, error) {
	var r SummaryRow
	err := db.QueryRow(`
		SELECT COUNT(*),
		       COUNT(DISTINCT date(time_created/1000, 'unixepoch')),
		       COALESCE(SUM(tokens_input), 0),
		       COALESCE(SUM(tokens_output), 0),
		       COALESCE(SUM(tokens_cache_read), 0),
		       COALESCE(SUM(tokens_cache_write), 0),
		       COALESCE(SUM(cost), 0),
		       COALESCE(SUM((SELECT COUNT(*) FROM message WHERE message.session_id = session.id)), 0)
		FROM session
	`).Scan(&r.TotalSessions, &r.ActiveDays, &r.TotalInputTokens, &r.TotalOutputTokens,
		&r.TotalCacheRead, &r.TotalCacheWrite, &r.TotalCost, &r.MessageCount)
	if err != nil {
		return nil, fmt.Errorf("query summary: %w", err)
	}

	totalTokens := r.TotalInputTokens + r.TotalCacheRead
	if totalTokens > 0 {
		r.CacheReadPct = float64(r.TotalCacheRead) / float64(totalTokens) * 100
		r.CacheWritePct = float64(r.TotalCacheWrite) / float64(totalTokens) * 100
	}

	return &r, nil
}

// GetProjects returns breakdown by project.
func GetProjects(db *sql.DB, daysAgo int) ([]ProjectRow, error) {
	var rows *sql.Rows
	var err error

	if daysAgo > 0 {
		since := time.Now().AddDate(0, 0, -daysAgo).UnixMilli()
		rows, err = db.Query(`
			SELECT COALESCE(p.name, s.directory) as project,
			       COUNT(s.id) as sessions,
			       SUM(s.tokens_input), SUM(s.tokens_output),
			       SUM(s.cost),
			       SUM((SELECT COUNT(*) FROM message WHERE message.session_id = s.id)) as msg_count
			FROM session s
			LEFT JOIN project p ON s.project_id = p.id
			WHERE s.time_created >= ?
			GROUP BY project
			ORDER BY sessions DESC
		`, since)
	} else {
		rows, err = db.Query(`
			SELECT COALESCE(p.name, s.directory) as project,
			       COUNT(s.id) as sessions,
			       SUM(s.tokens_input), SUM(s.tokens_output),
			       SUM(s.cost),
			       SUM((SELECT COUNT(*) FROM message WHERE message.session_id = s.id)) as msg_count
			FROM session s
			LEFT JOIN project p ON s.project_id = p.id
			GROUP BY project
			ORDER BY sessions DESC
		`)
	}
	if err != nil {
		return nil, fmt.Errorf("query projects: %w", err)
	}
	defer rows.Close()

	var results []ProjectRow
	for rows.Next() {
		var r ProjectRow
		if err := rows.Scan(&r.Project, &r.Sessions, &r.InputTokens, &r.OutputTokens, &r.Cost, &r.MessageCount); err != nil {
			return nil, fmt.Errorf("scan projects: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
