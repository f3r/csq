package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"github.com/f3r/csq/internal/bootstrap"
	"github.com/f3r/csq/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize csq: create ~/.csq/, write config, install bootstrap hook",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	csqDir := config.CsqDirPath()
	realHome := config.RealHome()

	if err := os.MkdirAll(csqDir, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", csqDir, err)
	}

	cfg := config.DefaultConfig()
	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Created %s\n", config.ConfigPath())

	scriptPath := filepath.Join(csqDir, "bootstrap.sh")
	if err := bootstrap.WriteScript(scriptPath); err != nil {
		return fmt.Errorf("writing bootstrap script: %w", err)
	}
	fmt.Fprintf(os.Stderr, "Created %s\n", scriptPath)

	if err := installSessionStartHook(realHome, scriptPath); err != nil {
		return fmt.Errorf("installing hook: %w", err)
	}

	fmt.Fprintln(os.Stderr, "csq initialized successfully!")
	return nil
}

type hookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

type hookRule struct {
	Matcher string        `json:"matcher"`
	Hooks   []hookCommand `json:"hooks"`
}

func installSessionStartHook(realHome, scriptPath string) error {
	settingsPath := filepath.Join(realHome, ".claude", "settings.json")

	data, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading settings: %w", err)
	}

	settings := make(map[string]json.RawMessage)
	if len(data) > 0 {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parsing settings: %w", err)
		}
	}

	hooks := make(map[string][]hookRule)
	needsMigration := false
	if raw, ok := settings["hooks"]; ok {
		if err := json.Unmarshal(raw, &hooks); err != nil {
			// Old array format — discard and migrate
			hooks = make(map[string][]hookRule)
			needsMigration = true
		}
	}

	if !needsMigration && hookContainsCommand(hooks["SessionStart"], scriptPath) {
		fmt.Fprintln(os.Stderr, "SessionStart hook already installed")
		return nil
	}

	newRule := hookRule{
		Matcher: "",
		Hooks: []hookCommand{
			{Type: "command", Command: scriptPath},
		},
	}
	hooks["SessionStart"] = append(hooks["SessionStart"], newRule)

	hooksJSON, err := json.Marshal(hooks)
	if err != nil {
		return fmt.Errorf("marshaling hooks: %w", err)
	}
	settings["hooks"] = hooksJSON

	result, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(settingsPath, result, 0644); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Installed SessionStart hook in %s\n", settingsPath)
	return nil
}

func hookContainsCommand(rules []hookRule, command string) bool {
	for _, rule := range rules {
		for _, h := range rule.Hooks {
			if h.Command == command {
				return true
			}
		}
	}
	return false
}
