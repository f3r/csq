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

func installSessionStartHook(realHome, scriptPath string) error {
	settingsPath := filepath.Join(realHome, ".claude", "settings.json")

	var settings map[string]interface{}
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("reading settings: %w", err)
		}
		settings = make(map[string]interface{})
	} else {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf("parsing settings: %w", err)
		}
	}

	hooks, _ := settings["hooks"].([]interface{})

	hookCommand := scriptPath
	for _, h := range hooks {
		hMap, ok := h.(map[string]interface{})
		if !ok {
			continue
		}
		if hMap["type"] == "SessionStart" && hMap["command"] == hookCommand {
			fmt.Fprintln(os.Stderr, "SessionStart hook already installed")
			return nil
		}
	}

	newHook := map[string]interface{}{
		"type":    "SessionStart",
		"command": hookCommand,
	}
	hooks = append(hooks, newHook)
	settings["hooks"] = hooks

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling settings: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(settingsPath, out, 0644); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Installed SessionStart hook in %s\n", settingsPath)
	return nil
}
