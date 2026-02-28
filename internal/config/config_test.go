package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.MaxDepth != 3 {
		t.Errorf("expected MaxDepth=3, got %d", cfg.MaxDepth)
	}
	if cfg.CSBinary != "cs" {
		t.Errorf("expected CSBinary=cs, got %s", cfg.CSBinary)
	}
	if len(cfg.Roots) != 1 {
		t.Errorf("expected 1 root, got %d", len(cfg.Roots))
	}
	if len(cfg.SymlinkDotfiles) == 0 {
		t.Error("expected symlink dotfiles to be populated")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CSQ_REAL_HOME", tmpDir)

	cfg := DefaultConfig()
	cfg.MaxDepth = 5
	cfg.Roots = []string{"/tmp/test"}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loaded.MaxDepth != 5 {
		t.Errorf("expected MaxDepth=5, got %d", loaded.MaxDepth)
	}
	if len(loaded.Roots) != 1 || loaded.Roots[0] != "/tmp/test" {
		t.Errorf("unexpected roots: %v", loaded.Roots)
	}
}

func TestLoadMissing(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CSQ_REAL_HOME", tmpDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load should not fail on missing config: %v", err)
	}
	if cfg.CSBinary != "cs" {
		t.Errorf("expected default CSBinary, got %s", cfg.CSBinary)
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CSQ_REAL_HOME", tmpDir)

	csqDir := filepath.Join(tmpDir, CsqDir)
	if err := os.MkdirAll(csqDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(csqDir, ConfigFile), []byte("{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load()
	if err == nil {
		t.Error("expected error on invalid JSON")
	}
}

func TestExpandPath(t *testing.T) {
	home := RealHome()
	expanded := ExpandPath("~/code")
	expected := filepath.Join(home, "code")
	if expanded != expected {
		t.Errorf("expected %s, got %s", expected, expanded)
	}

	abs := ExpandPath("/absolute/path")
	if abs != "/absolute/path" {
		t.Errorf("expected /absolute/path, got %s", abs)
	}
}

func TestExpandedRoots(t *testing.T) {
	cfg := Config{Roots: []string{"~/code", "/absolute"}}
	roots := cfg.ExpandedRoots()
	if len(roots) != 2 {
		t.Fatalf("expected 2 roots, got %d", len(roots))
	}
	if roots[1] != "/absolute" {
		t.Errorf("expected /absolute, got %s", roots[1])
	}
}

func TestConfigJSON(t *testing.T) {
	cfg := DefaultConfig()
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	var decoded Config
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.MaxDepth != cfg.MaxDepth {
		t.Errorf("roundtrip mismatch: %d != %d", decoded.MaxDepth, cfg.MaxDepth)
	}
}
