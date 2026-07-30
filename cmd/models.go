package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/abuelgheit/opencode-stats/internal/display"
	"github.com/abuelgheit/opencode-stats/internal/stats"
)

func newModelsCmd() *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "models",
		Short: "Show breakdown by model",
		Long:  `Display usage statistics grouped by model.`,
		Run: func(cmd *cobra.Command, args []string) {
			database, err := openDB()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			defer database.Close()

			rows, err := stats.GetModels(database, days)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}

			if len(rows) == 0 {
				fmt.Println("No sessions found.")
				return
			}

			display.PrintModels(rows)
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 0, "Show from last N days (0 = all)")

	return cmd
}
