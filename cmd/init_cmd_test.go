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

	hooks := readHooks(t, tmpDir)
	rules, ok := hooks["SessionStart"]
	if !ok {
		t.Fatal("expected SessionStart key in hooks")
	}
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(rules))
	}
	if rules[0].Hooks[0].Command != scriptPath {
		t.Errorf("expected command %s, got %s", scriptPath, rules[0].Hooks[0].Command)
	}
	if rules[0].Hooks[0].Type != "command" {
		t.Errorf("expected type 'command', got %s", rules[0].Hooks[0].Type)
	}
}

func TestInstallSessionStartHook_ExistingHooksRecord(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	existing := `{
  "permissions": {
    "allow": ["Bash"]
  },
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "hooks": [
          {
            "type": "command",
            "command": "echo hello"
          }
        ]
      }
    ]
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

	hooks := readHooks(t, tmpDir)

	if _, ok := hooks["PreToolUse"]; !ok {
		t.Error("existing PreToolUse hook was lost")
	}

	rules, ok := hooks["SessionStart"]
	if !ok {
		t.Fatal("expected SessionStart key in hooks")
	}
	if rules[0].Hooks[0].Command != scriptPath {
		t.Errorf("expected command %s, got %s", scriptPath, rules[0].Hooks[0].Command)
	}

	data, _ := os.ReadFile(settingsPath)
	if !strings.Contains(string(data), `"permissions"`) {
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

	hooks := readHooks(t, tmpDir)
	rules, ok := hooks["SessionStart"]
	if !ok {
		t.Fatal("expected SessionStart key in hooks")
	}
	if rules[0].Hooks[0].Command != scriptPath {
		t.Errorf("expected command %s, got %s", scriptPath, rules[0].Hooks[0].Command)
	}

	data, _ := os.ReadFile(settingsPath)
	if !strings.Contains(string(data), `"permissions"`) {
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
  "hooks": {
    "SessionStart": [
      {
        "matcher": "",
        "hooks": [
          {
            "type": "command",
            "command": "` + scriptPath + `"
          }
        ]
      }
    ]
  }
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

func TestInstallSessionStartHook_PreservesOtherKeys(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	existing := `{
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

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := parsed["model"]; !ok {
		t.Error("model key was lost")
	}
	if _, ok := parsed["permissions"]; !ok {
		t.Error("permissions key was lost")
	}
	if _, ok := parsed["hooks"]; !ok {
		t.Error("hooks key was not added")
	}
}

func TestInstallSessionStartHook_MigratesOldArrayFormat(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	existing := `{
  "hooks": [
    {
      "type": "SessionStart",
      "command": "/old/path/bootstrap.sh"
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

	hooks := readHooks(t, tmpDir)
	rules, ok := hooks["SessionStart"]
	if !ok {
		t.Fatal("expected SessionStart key in hooks")
	}
	if rules[0].Hooks[0].Command != scriptPath {
		t.Errorf("expected command %s, got %s", scriptPath, rules[0].Hooks[0].Command)
	}
}

func readHooks(t *testing.T, homeDir string) map[string][]hookRule {
	t.Helper()
	settingsPath := filepath.Join(homeDir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("reading settings: %v", err)
	}

	var parsed map[string]json.RawMessage
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("invalid JSON: %v\ncontent: %s", err, string(data))
	}

	raw, ok := parsed["hooks"]
	if !ok {
		t.Fatal("no hooks key in settings")
	}

	var hooks map[string][]hookRule
	if err := json.Unmarshal(raw, &hooks); err != nil {
		t.Fatalf("invalid hooks format: %v\ncontent: %s", err, string(raw))
	}

	return hooks
}
