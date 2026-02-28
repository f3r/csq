package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

	data, err := os.ReadFile(settingsPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading settings: %w", err)
	}

	content := string(data)

	if strings.Contains(content, scriptPath) {
		fmt.Fprintln(os.Stderr, "SessionStart hook already installed")
		return nil
	}

	hookEntry := fmt.Sprintf(`    {
      "type": "SessionStart",
      "command": "%s"
    }`, scriptPath)

	var result string

	switch {
	case len(data) == 0:
		result = fmt.Sprintf("{\n  \"hooks\": [\n%s\n  ]\n}", hookEntry)

	case strings.Contains(content, `"hooks"`):
		lastBracket := strings.LastIndex(content, "]")
		if lastBracket == -1 {
			return fmt.Errorf("malformed settings: found hooks key but no closing bracket")
		}
		before := content[:lastBracket]
		after := content[lastBracket:]
		trimmed := strings.TrimRight(before, " \t\n\r")
		if strings.HasSuffix(trimmed, "[") {
			result = trimmed + "\n" + hookEntry + "\n  " + after
		} else {
			result = trimmed + ",\n" + hookEntry + "\n  " + after
		}

	default:
		lastBrace := strings.LastIndex(content, "}")
		if lastBrace == -1 {
			return fmt.Errorf("malformed settings: no closing brace")
		}
		before := strings.TrimRight(content[:lastBrace], " \t\n\r")
		after := content[lastBrace:]
		separator := ","
		if strings.TrimSpace(before) == "{" {
			separator = ""
		}
		result = before + separator + "\n  \"hooks\": [\n" + hookEntry + "\n  ]\n" + after
	}

	var check interface{}
	if err := json.Unmarshal([]byte(result), &check); err != nil {
		return fmt.Errorf("hook insertion produced invalid JSON: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(settingsPath, []byte(result), 0644); err != nil {
		return fmt.Errorf("writing settings: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Installed SessionStart hook in %s\n", settingsPath)
	return nil
}
