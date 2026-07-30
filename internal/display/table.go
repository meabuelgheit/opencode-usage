package display

import (
	"fmt"
	"os"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"github.com/abuelgheit/opencode-usage/internal/stats"
)

func newTable() table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleDefault)
	t.Style().Color.RowAlternate = text.Colors{text.Faint}
	t.Style().Options.SeparateFooter = true
	return t
}

// num formats an int64 with human-readable suffixes (K, M, B).
func num(n int64) string {
	switch {
	case n < 1000:
		return fmt.Sprintf("%d", n)
	case n < 1000000:
		return fmt.Sprintf("%.2fK", float64(n)/1000)
	case n < 1000000000:
		return fmt.Sprintf("%.2fM", float64(n)/1000000)
	default:
		return fmt.Sprintf("%.2fB", float64(n)/1000000000)
	}
}

// cost formats a float as dollar amount.
func costStr(c float64) string {
	return fmt.Sprintf("$%.2f", c)
}

// pctStr formats a float as a percentage string.
func pctStr(v float64) string {
	return fmt.Sprintf("%.1f%%", v)
}

// modelShort shortens a model name for display (take last 2 segments).
func modelShort(m string) string {
	parts := strings.Split(m, "/")
	if len(parts) > 2 {
		parts = parts[len(parts)-2:]
	}
	return strings.Join(parts, "/")
}

// PrintSessions renders session rows as a table.
func PrintSessions(rows []stats.SessionRow) {
	t := newTable()
	t.AppendHeader(table.Row{"Date", "Session", "Agent", "Model", "Input", "Output", "Total", "Cache Read", "Cache Read%", "MESSAGES", "Cost"})

	for _, r := range rows {
		t.AppendRow(table.Row{
			r.Date,
			r.Slug,
			r.Agent,
			modelShort(r.Model),
			num(r.InputTokens),
			num(r.OutputTokens),
			num(r.TotalTokens),
			num(r.CacheRead),
			pctStr(r.CacheReadPct),
			r.MessageCount,
			costStr(r.Cost),
		})
	}

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignRight},
		{Number: 7, Align: text.AlignRight},
		{Number: 8, Align: text.AlignRight},
		{Number: 9, Align: text.AlignRight},
		{Number: 10, Align: text.AlignRight},
		{Number: 11, Align: text.AlignRight},
	})

	var (
		totInput, totOutput, totCacheRead, totTotal int64
		totCost                                     float64
	)
	for _, r := range rows {
		totInput += r.InputTokens
		totOutput += r.OutputTokens
		totCacheRead += r.CacheRead
		totTotal += r.TotalTokens
		totCost += r.Cost
	}
	totReadPct := 0.0
	if totTotal > 0 {
		totReadPct = float64(totCacheRead) / float64(totTotal) * 100
	}
	t.AppendFooter(table.Row{
		fmt.Sprintf("TOTAL (%d)", len(rows)), "", "", "",
		num(totInput), num(totOutput), num(totTotal),
		num(totCacheRead), pctStr(totReadPct), costStr(totCost),
	})

	t.Render()
}

// PrintDaily renders daily aggregate rows as a table.
func PrintDaily(rows []stats.DailyRow) {
	t := newTable()
	t.AppendHeader(table.Row{"Date", "Sessions", "MESSAGES", "Input", "Output", "Total", "Cache Read", "Cache Write", "Cache Read%", "Cache Write%", "Cost"})

	for _, r := range rows {
		t.AppendRow(table.Row{
			r.Date, r.Sessions, r.MessageCount,
			num(r.InputTokens), num(r.OutputTokens),
			num(r.TotalTokens),
			num(r.CacheRead), num(r.CacheWrite),
			pctStr(r.CacheReadPct), pctStr(r.CacheWritePct),
			costStr(r.Cost),
		})
	}

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignRight},
		{Number: 7, Align: text.AlignRight},
		{Number: 8, Align: text.AlignRight},
		{Number: 9, Align: text.AlignRight},
		{Number: 10, Align: text.AlignRight},
		{Number: 11, Align: text.AlignRight},
	})

	var (
		totSessions, totMessages                    int
		totInput, totOutput, totCacheRead, totCacheWrite, totTotal int64
		totCost                                                     float64
	)
	for _, r := range rows {
		totSessions += r.Sessions
		totMessages += r.MessageCount
		totInput += r.InputTokens
		totOutput += r.OutputTokens
		totCacheRead += r.CacheRead
		totCacheWrite += r.CacheWrite
		totTotal += r.TotalTokens
		totCost += r.Cost
	}
	totReadPct := 0.0
	totWritePct := 0.0
	if totTotal > 0 {
		totReadPct = float64(totCacheRead) / float64(totTotal) * 100
		totWritePct = float64(totCacheWrite) / float64(totTotal) * 100
	}

	t.AppendFooter(table.Row{
		"TOTAL", totSessions, totMessages,
		num(totInput), num(totOutput), num(totTotal),
		num(totCacheRead), num(totCacheWrite),
		pctStr(totReadPct), pctStr(totWritePct), costStr(totCost),
	})

	t.Render()
}

// PrintModels renders model breakdown rows as a table.
func PrintModels(rows []stats.ModelRow) {
	t := newTable()
	t.AppendHeader(table.Row{"Model", "Sessions", "MESSAGES", "Input", "Output", "Total", "Cache Read", "Cache Write", "Cache Read%", "Cache Write%", "Cost"})

	for _, r := range rows {
		t.AppendRow(table.Row{
			modelShort(r.Model), r.Sessions, r.MessageCount,
			num(r.InputTokens), num(r.OutputTokens),
			num(r.TotalTokens),
			num(r.CacheRead), num(r.CacheWrite),
			pctStr(r.CacheReadPct), pctStr(r.CacheWritePct),
			costStr(r.Cost),
		})
	}

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignRight},
		{Number: 7, Align: text.AlignRight},
		{Number: 8, Align: text.AlignRight},
		{Number: 9, Align: text.AlignRight},
		{Number: 10, Align: text.AlignRight},
		{Number: 11, Align: text.AlignRight},
	})

	var (
		totSessions, totMessages                    int
		totInput, totOutput, totCacheRead, totCacheWrite, totTotal int64
		totCost                                                     float64
	)
	for _, r := range rows {
		totSessions += r.Sessions
		totMessages += r.MessageCount
		totInput += r.InputTokens
		totOutput += r.OutputTokens
		totCacheRead += r.CacheRead
		totCacheWrite += r.CacheWrite
		totTotal += r.TotalTokens
		totCost += r.Cost
	}
	totReadPct := 0.0
	totWritePct := 0.0
	if totTotal > 0 {
		totReadPct = float64(totCacheRead) / float64(totTotal) * 100
		totWritePct = float64(totCacheWrite) / float64(totTotal) * 100
	}

	t.AppendFooter(table.Row{
		"TOTAL", totSessions, totMessages,
		num(totInput), num(totOutput), num(totTotal),
		num(totCacheRead), num(totCacheWrite),
		pctStr(totReadPct), pctStr(totWritePct), costStr(totCost),
	})

	t.Render()
}

// PrintAgents renders agent breakdown rows as a table.
func PrintAgents(rows []stats.AgentRow) {
	t := newTable()
	t.AppendHeader(table.Row{"Agent", "Sessions", "MESSAGES", "Input", "Output", "Total", "Cache Read", "Cache Read%", "Cost"})

	for _, r := range rows {
		t.AppendRow(table.Row{
			r.Agent, r.Sessions, r.MessageCount,
			num(r.InputTokens), num(r.OutputTokens),
			num(r.TotalTokens), num(r.CacheRead),
			pctStr(r.CacheReadPct),
			costStr(r.Cost),
		})
	}

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignRight},
		{Number: 7, Align: text.AlignRight},
		{Number: 8, Align: text.AlignRight},
		{Number: 9, Align: text.AlignRight},
	})

	var (
		totSessions, totMessages                int
		totInput, totOutput, totCacheRead, totTotal int64
		totCost                                     float64
	)
	for _, r := range rows {
		totSessions += r.Sessions
		totMessages += r.MessageCount
		totInput += r.InputTokens
		totOutput += r.OutputTokens
		totCacheRead += r.CacheRead
		totTotal += r.TotalTokens
		totCost += r.Cost
	}
	totReadPct := 0.0
	if totTotal > 0 {
		totReadPct = float64(totCacheRead) / float64(totTotal) * 100
	}

	t.AppendFooter(table.Row{
		"TOTAL", totSessions, totMessages,
		num(totInput), num(totOutput), num(totTotal),
		num(totCacheRead), pctStr(totReadPct), costStr(totCost),
	})

	t.Render()
}

// PrintSummary renders the all-time summary as a single-row table.
func PrintSummary(s *stats.SummaryRow) {
	t := newTable()
	t.AppendHeader(table.Row{"Metric", "Value"})
	t.AppendRows([]table.Row{
		{"Total Sessions", s.TotalSessions},
		{"Active Days", s.ActiveDays},
		{"Total Input Tokens", num(s.TotalInputTokens)},
		{"Total Output Tokens", num(s.TotalOutputTokens)},
		{"Total Cache Read", num(s.TotalCacheRead)},
		{"Total Cache Write", num(s.TotalCacheWrite)},
		{"Cache Read Rate", pctStr(s.CacheReadPct)},
		{"Total Messages", s.MessageCount},
		{"Total Cost", costStr(s.TotalCost)},
	})

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 2, Align: text.AlignRight},
	})

	t.Render()
}

// PrintProjects renders project breakdown rows as a table.
func PrintProjects(rows []stats.ProjectRow) {
	t := newTable()
	t.AppendHeader(table.Row{"Project", "Sessions", "Input", "Output", "MESSAGES", "Cost"})

	for _, r := range rows {
		// Truncate long project paths
		proj := r.Project
		if len(proj) > 50 {
			proj = "..." + proj[len(proj)-47:]
		}
		t.AppendRow(table.Row{
			proj, r.Sessions,
			num(r.InputTokens), num(r.OutputTokens),
			r.MessageCount,
			costStr(r.Cost),
		})
	}

	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 3, Align: text.AlignRight},
		{Number: 4, Align: text.AlignRight},
		{Number: 5, Align: text.AlignRight},
		{Number: 6, Align: text.AlignRight},
	})

	var (
		totSessions, totMessages  int
		totInput, totOutput       int64
		totCost                   float64
	)
	for _, r := range rows {
		totSessions += r.Sessions
		totMessages += r.MessageCount
		totInput += r.InputTokens
		totOutput += r.OutputTokens
		totCost += r.Cost
	}

	t.AppendFooter(table.Row{
		"TOTAL", totSessions, totMessages,
		num(totInput), num(totOutput), costStr(totCost),
	})

	t.Render()
}
