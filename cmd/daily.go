package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/abuelgheit/opencode-stats/internal/display"
	"github.com/abuelgheit/opencode-stats/internal/stats"
)

func newDailyCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "daily",
		Short: "Show daily usage breakdown",
		Long:  `Display daily aggregates of sessions, token usage, and cost.`,
		Run: func(cmd *cobra.Command, args []string) {
			database, err := openDB()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer database.Close()

			rows, err := stats.GetDaily(database, days)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(rows) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			display.PrintDaily(rows)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 30, "Show last N days")

	return cmd
}
