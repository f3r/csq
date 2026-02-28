package cmd

import (
	"fmt"
	"os"

	"github.com/f3r/csq/internal/config"
	"github.com/f3r/csq/internal/discovery"
	"github.com/f3r/csq/internal/launcher"
	"github.com/f3r/csq/internal/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "csq [project]",
	Short: "Multi-project Claude Squad launcher",
	Long:  "Launch Claude Squad with per-project state isolation and fuzzy project search.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runRoot,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("refresh", "r", false, "Refresh project cache before searching")
}

func runRoot(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	refresh, _ := cmd.Flags().GetBool("refresh")

	var projects []discovery.Project
	if refresh {
		projects, err = discovery.DiscoverFresh(cfg)
	} else {
		projects, err = discovery.Discover(cfg)
	}
	if err != nil {
		return fmt.Errorf("discovering projects: %w", err)
	}

	if len(projects) == 0 {
		return fmt.Errorf("no projects found in %v (max_depth: %d)", cfg.Roots, cfg.MaxDepth)
	}

	var selected discovery.Project
	csArgs := extractCSArgs(cmd)

	if len(args) == 1 {
		matches := discovery.FuzzyMatch(projects, args[0])
		switch len(matches) {
		case 0:
			return fmt.Errorf("no project matching %q", args[0])
		case 1:
			selected = matches[0]
		default:
			picked, err := tui.Pick(matches, cfg)
			if err != nil {
				return err
			}
			if picked == nil {
				return nil
			}
			selected = *picked
		}
	} else {
		picked, err := tui.Pick(projects, cfg)
		if err != nil {
			return err
		}
		if picked == nil {
			return nil
		}
		selected = *picked
	}

	fmt.Fprintf(os.Stderr, "Launching cs for %s...\n", selected.Name)
	return launcher.Launch(selected.Name, selected.Path, cfg, csArgs)
}

func extractCSArgs(cmd *cobra.Command) []string {
	args := os.Args
	for i, a := range args {
		if a == "--" {
			return args[i+1:]
		}
	}
	return nil
}
