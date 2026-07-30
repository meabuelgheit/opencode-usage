package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/abuelgheit/opencode-usage/internal/display"
	"github.com/abuelgheit/opencode-usage/internal/stats"
)

func newSessionCmd() *cobra.Command {
	var limit int
	var days int

	cmd := &cobra.Command{
		Use:   "session",
		Short: "Show recent sessions",
		Long:  `Display recent OpenCode sessions with token usage and cost.`,
		Run: func(cmd *cobra.Command, args []string) {
			database, err := openDB()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer database.Close()

			rows, err := stats.GetSessions(database, limit, days)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(rows) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			display.PrintSessions(rows)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of sessions to show")
	cmd.Flags().IntVarP(&days, "days", "d", 0, "Show sessions from last N days (0 = all)")

	return cmd
}
