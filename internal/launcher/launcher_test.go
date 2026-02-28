package launcher

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/f3r/csq/internal/config"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"acme/api", "acme--api"},
		{"simple", "simple"},
		{"org/sub.project", "org--sub_project"},
		{"has spaces/here", "has_spaces--here"},
	}

	for _, tt := range tests {
		got := SanitizeName(tt.input)
		if got != tt.expected {
			t.Errorf("SanitizeName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestUnsanitizeName(t *testing.T) {
	got := UnsanitizeName("acme--api")
	if got != "acme/api" {
		t.Errorf("UnsanitizeName = %q, want acme/api", got)
	}
}

func TestEnsureHomeCreatesDir(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CSQ_REAL_HOME", tmpDir)

	cfg := config.Config{
		HomeBase:        filepath.Join(tmpDir, ".csq", "homes"),
		SymlinkDotfiles: []string{".gitconfig"},
	}

	// Create the dotfile to symlink
	if err := os.WriteFile(filepath.Join(tmpDir, ".gitconfig"), []byte("[user]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	homeDir, err := EnsureHome("test/project", cfg)
	if err != nil {
		t.Fatalf("EnsureHome failed: %v", err)
	}

	if _, err := os.Stat(homeDir); os.IsNotExist(err) {
		t.Error("home dir was not created")
	}

	// Check symlink
	link := filepath.Join(homeDir, ".gitconfig")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("reading symlink: %v", err)
	}
	expected := filepath.Join(tmpDir, ".gitconfig")
	if target != expected {
		t.Errorf("symlink target = %q, want %q", target, expected)
	}
}

func TestEnsureHomeIdempotent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CSQ_REAL_HOME", tmpDir)

	cfg := config.Config{
		HomeBase:        filepath.Join(tmpDir, ".csq", "homes"),
		SymlinkDotfiles: []string{".gitconfig"},
	}

	if err := os.WriteFile(filepath.Join(tmpDir, ".gitconfig"), []byte("[user]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	home1, err := EnsureHome("test/project", cfg)
	if err != nil {
		t.Fatal(err)
	}

	home2, err := EnsureHome("test/project", cfg)
	if err != nil {
		t.Fatal(err)
	}

	if home1 != home2 {
		t.Errorf("idempotency failed: %s != %s", home1, home2)
	}
}

func TestEnsureHomeSkipsMissingDotfiles(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CSQ_REAL_HOME", tmpDir)

	cfg := config.Config{
		HomeBase:        filepath.Join(tmpDir, ".csq", "homes"),
		SymlinkDotfiles: []string{".nonexistent-file"},
	}

	_, err := EnsureHome("test/project", cfg)
	if err != nil {
		t.Fatalf("should not fail on missing dotfile: %v", err)
	}
}

func TestBuildEnvSetsRequiredVars(t *testing.T) {
	env := buildEnv("/fake/home", "/fake/project")

	envMap := make(map[string]string)
	for _, e := range env {
		for i, c := range e {
			if c == '=' {
				envMap[e[:i]] = e[i+1:]
				break
			}
		}
	}

	if envMap["HOME"] != "/fake/home" {
		t.Errorf("HOME = %q, want /fake/home", envMap["HOME"])
	}
	if envMap["PWD"] != "/fake/project" {
		t.Errorf("PWD = %q, want /fake/project", envMap["PWD"])
	}
	if _, ok := envMap["CSQ_REAL_HOME"]; !ok {
		t.Error("CSQ_REAL_HOME not set")
	}
	if _, ok := envMap["CSQ_PROJECT"]; !ok {
		t.Error("CSQ_PROJECT not set")
	}
}
