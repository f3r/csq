package cmd

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/f3r/csq/internal/config"
	"github.com/f3r/csq/internal/launcher"
	"github.com/f3r/csq/internal/state"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show all active sessions across all projects",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	homeBase := config.ExpandPath(cfg.HomeBase)
	sessions, err := state.AllSessions(homeBase)
	if err != nil {
		return err
	}

	if len(sessions) == 0 {
		fmt.Println("No active sessions.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROJECT\tTITLE\tSTATUS\tBRANCH")

	total := 0
	for _, s := range sessions {
		name := launcher.UnsanitizeName(s.ProjectName)
		for _, inst := range s.Instances {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
				name,
				truncate(inst.Title, 40),
				inst.Status,
				inst.Branch,
			)
			total++
		}
	}

	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\n%d session(s) across %d project(s)\n", total, len(sessions))
	return nil
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
