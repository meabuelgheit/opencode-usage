package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/abuelgheit/opencode-stats/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("opencode-stats", version.String())
		},
	}
}
