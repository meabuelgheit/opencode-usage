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
	TotalTokens int64   // input + cache_read (effective prompt size)
	CacheReadPct float64 // cache_read / (input + cache_read) * 100
	Cost         float64 // cost in dollars
	MessageCount int     // number of messages
}

// DailyRow represents daily aggregated stats.
type DailyRow struct {
	Date          string
	Sessions      int
	InputTokens   int64
	OutputTokens  int64
	CacheRead     int64
	CacheWrite    int64
	TotalTokens   int64 // input + cache_read
	CacheReadPct  float64 // cache_read / (input + cache_read) * 100
	CacheWritePct float64 // cache_write / total * 100
	Cost          float64
	MessageCount  int
}

// ModelRow represents per-model breakdown.
type ModelRow struct {
	Model         string
	Sessions      int
	InputTokens   int64
	OutputTokens  int64
	CacheRead     int64
	CacheWrite    int64
	TotalTokens   int64 // input + cache_read
	CacheReadPct  float64 // cache_read / (input + cache_read) * 100
	CacheWritePct float64 // cache_write / total * 100
	Cost          float64
	BlendCostPerM float64 // blended cost per 1M tokens (cost / ((input+output+cache_read)/1M))
	MessageCount  int
}

// AgentRow represents per-agent breakdown.
type AgentRow struct {
	Agent         string
	Sessions      int
	InputTokens   int64
	OutputTokens  int64
	CacheRead     int64
	CacheWrite    int64
	TotalTokens   int64   // InputTokens + CacheRead (computed in Go)
	CacheReadPct  float64 // CacheRead / TotalTokens * 100
	CacheWritePct float64 // CacheWrite / TotalTokens * 100
	Cost          float64
	MessageCount  int
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
	CostPerMTokens   float64 // cost per million input+output tokens
	EffectiveCostPerMTokens float64 // cost per million tokens including cache reads
	CacheReadPct     float64 // aggregate
	CacheWritePct    float64 // aggregate
	MessageCount     int
}

// ProjectRow represents per-project breakdown.
type ProjectRow struct {
	Project     string // project name, fallback to directory
	Sessions    int
	InputTokens int64
	OutputTokens int64
	Cost         float64
	MessageCount int
}
