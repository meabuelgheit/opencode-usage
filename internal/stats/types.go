package stats

// SessionRow represents a single session row.
type SessionRow struct {
	Date        string  // formatted date string YYYY-MM-DD
	Slug        string  // session slug/name
	Agent       string  // agent type (architect, coder, etc.)
	Model       string  // model name (shortened for display)
	InputTokens int64   // tokens_input
	OutputTokens int64  // tokens_output
	CacheRead   int64   // tokens_cache_read
	Cost        float64 // cost in dollars
}

// DailyRow represents daily aggregated stats.
type DailyRow struct {
	Date        string
	Sessions    int
	InputTokens int64
	OutputTokens int64
	CacheRead   int64
	CacheWrite  int64
	Cost        float64
}

// ModelRow represents per-model breakdown.
type ModelRow struct {
	Model       string
	Sessions    int
	InputTokens int64
	OutputTokens int64
	CacheRead   int64
	CacheWrite  int64
	Cost        float64
}

// AgentRow represents per-agent breakdown.
type AgentRow struct {
	Agent       string
	Sessions    int
	InputTokens int64
	OutputTokens int64
	Cost        float64
}

// SummaryRow represents all-time totals.
type SummaryRow struct {
	TotalSessions    int
	ActiveDays       int
	TotalInputTokens int64
	TotalOutputTokens int64
	TotalCacheRead   int64
	TotalCacheWrite  int64
	TotalCost        float64
}

// ProjectRow represents per-project breakdown.
type ProjectRow struct {
	Project     string // project name, fallback to directory
	Sessions    int
	InputTokens int64
	OutputTokens int64
	Cost        float64
}
