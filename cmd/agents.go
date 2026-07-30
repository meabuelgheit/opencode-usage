package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/abuelgheit/opencode-usage/internal/display"
	"github.com/abuelgheit/opencode-usage/internal/stats"
)

func newAgentsCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Show breakdown by agent",
		Long:  `Display usage statistics grouped by agent type.`,
		Run: func(cmd *cobra.Command, args []string) {
			database, err := openDB()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer database.Close()

			rows, err := stats.GetAgents(database, days)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(rows) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			display.PrintAgents(rows)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 0, "Show from last N days (0 = all)")

	return cmd
}
