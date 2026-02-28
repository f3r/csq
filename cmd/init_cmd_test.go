package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallSessionStartHook_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := "/tmp/csq/bootstrap.sh"

	err := installSessionStartHook(tmpDir, scriptPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	settingsPath := filepath.Join(tmpDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("reading settings: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if !strings.Contains(string(data), scriptPath) {
		t.Error("hook command not found in output")
	}
}

func TestInstallSessionStartHook_ExistingHooksArray(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	existing := `{
  "permissions": {
    "allow": ["Bash"]
  },
  "hooks": [
    {
      "type": "PreToolUse",
      "command": "echo hello"
    }
  ]
}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	scriptPath := "/tmp/csq/bootstrap.sh"
	err := installSessionStartHook(tmpDir, scriptPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("reading settings: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v\ncontent: %s", err, string(data))
	}

	content := string(data)
	if !strings.Contains(content, "echo hello") {
		t.Error("existing hook was lost")
	}
	if !strings.Contains(content, scriptPath) {
		t.Error("new hook not found in output")
	}
	if !strings.Contains(content, `"permissions"`) {
		t.Error("existing permissions key was lost")
	}
}

func TestInstallSessionStartHook_NoHooksKey(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	existing := `{
  "permissions": {
    "allow": ["Bash"]
  }
}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	scriptPath := "/tmp/csq/bootstrap.sh"
	err := installSessionStartHook(tmpDir, scriptPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("reading settings: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v\ncontent: %s", err, string(data))
	}

	content := string(data)
	if !strings.Contains(content, scriptPath) {
		t.Error("hook not found in output")
	}
	if !strings.Contains(content, `"permissions"`) {
		t.Error("existing permissions key was lost")
	}
}

func TestInstallSessionStartHook_AlreadyInstalled(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	scriptPath := "/tmp/csq/bootstrap.sh"
	existing := `{
  "hooks": [
    {
      "type": "SessionStart",
      "command": "` + scriptPath + `"
    }
  ]
}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	err := installSessionStartHook(tmpDir, scriptPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != existing {
		t.Error("file was modified when hook was already installed")
	}
}

func TestInstallSessionStartHook_PreservesFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	existing := `{
  "apiKey": "test-key",
  "model": "sonnet",
  "permissions": {
    "allow": ["Bash", "Read"]
  }
}`
	settingsPath := filepath.Join(claudeDir, "settings.json")
	if err := os.WriteFile(settingsPath, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	scriptPath := "/tmp/csq/bootstrap.sh"
	err := installSessionStartHook(tmpDir, scriptPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatal(err)
	}

	content := string(data)
	if !strings.Contains(content, `"apiKey": "test-key"`) {
		t.Error("apiKey was reformatted or lost")
	}
	if !strings.Contains(content, `"model": "sonnet"`) {
		t.Error("model was reformatted or lost")
	}
}
