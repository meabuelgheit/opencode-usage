package cmd

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/abuelgheit/opencode-stats/internal/db"
	"github.com/abuelgheit/opencode-stats/internal/stats"
)

var (
	dbPath string
)

// openDB is a helper that opens the DB and handles errors.
func openDB() (*sql.DB, error) {
	database, err := db.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return database, nil
}

// Execute runs the root command.
func Execute() {
	rootCmd := &cobra.Command{
		Use:   "opencode-stats",
		Short: "Display OpenCode usage statistics",
		Long:  `opencode-stats reads the OpenCode SQLite database and displays usage statistics including sessions, token counts, costs, and breakdowns by model, agent, and project.`,
	}

	rootCmd.PersistentFlags().StringVar(&dbPath, "db", stats.DefaultDBPath(), "Path to opencode database")

	rootCmd.AddCommand(newSessionCmd())
	rootCmd.AddCommand(newDailyCmd())
	rootCmd.AddCommand(newModelsCmd())
	rootCmd.AddCommand(newAgentsCmd())
	rootCmd.AddCommand(newSummaryCmd())
	rootCmd.AddCommand(newProjectsCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
