package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/f3r/csq/internal/config"
	"github.com/f3r/csq/internal/discovery"
	"github.com/f3r/csq/internal/launcher"
	"github.com/f3r/csq/internal/state"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all discovered projects with session counts",
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().BoolP("refresh", "r", false, "Refresh project cache")
}

func runList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	refresh, _ := cmd.Flags().GetBool("refresh")

	var projects []discovery.Project
	if refresh {
		projects, err = discovery.DiscoverFresh(cfg)
	} else {
		projects, err = discovery.Discover(cfg)
	}
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "PROJECT\tSESSIONS\tPATH")

	for _, p := range projects {
		homeDir := launcher.HomeDir(p.Name, cfg)
		count := state.CountSessions(homeDir)
		sessStr := "-"
		if count > 0 {
			sessStr = fmt.Sprintf("%d", count)
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, sessStr, p.Path)
	}

	return w.Flush()
}
