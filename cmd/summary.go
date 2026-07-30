package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/abuelgheit/opencode-usage/internal/display"
	"github.com/abuelgheit/opencode-usage/internal/stats"
)

func newSummaryCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "summary",
		Short: "Show summary statistics",
		Long:  `Display totals for sessions, tokens, and cost. Use --days to filter.`,
		Run: func(cmd *cobra.Command, args []string) {
			database, err := openDB()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer database.Close()

			s, err := stats.GetSummary(database, days)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			display.PrintSummary(s)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 0, "Show from last N days (0 = all)")

	return cmd
}
