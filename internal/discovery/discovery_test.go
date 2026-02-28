package discovery

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/f3r/csq/internal/config"
)

func setupTestRepos(t *testing.T) (string, config.Config) {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("CSQ_REAL_HOME", tmpDir)

	repos := []string{
		"org1/repoA",
		"org1/repoB",
		"org2/project-x",
		"standalone",
	}

	for _, r := range repos {
		repoPath := filepath.Join(tmpDir, "code", r)
		gitDir := filepath.Join(repoPath, ".git")
		if err := os.MkdirAll(gitDir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Non-git dir should be skipped
	if err := os.MkdirAll(filepath.Join(tmpDir, "code", "not-a-repo"), 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		Roots:    []string{filepath.Join(tmpDir, "code")},
		MaxDepth: 3,
		CSBinary: "cs",
		HomeBase: filepath.Join(tmpDir, ".csq", "homes"),
	}

	return tmpDir, cfg
}

func TestDiscoverFindsRepos(t *testing.T) {
	_, cfg := setupTestRepos(t)

	projects, err := scan(cfg)
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	if len(projects) != 4 {
		t.Errorf("expected 4 projects, got %d", len(projects))
		for _, p := range projects {
			t.Logf("  %s -> %s", p.Name, p.Path)
		}
	}
}

func TestDiscoverSorted(t *testing.T) {
	_, cfg := setupTestRepos(t)

	projects, err := scan(cfg)
	if err != nil {
		t.Fatal(err)
	}

	for i := 1; i < len(projects); i++ {
		if projects[i].Name < projects[i-1].Name {
			t.Errorf("projects not sorted: %s before %s", projects[i-1].Name, projects[i].Name)
		}
	}
}

func TestDiscoverRespectsMaxDepth(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CSQ_REAL_HOME", tmpDir)

	deepRepo := filepath.Join(tmpDir, "code", "a", "b", "c", "d", "deep-repo", ".git")
	if err := os.MkdirAll(deepRepo, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		Roots:    []string{filepath.Join(tmpDir, "code")},
		MaxDepth: 2,
	}

	projects, err := scan(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if len(projects) != 0 {
		t.Errorf("expected 0 projects with max_depth=2, got %d", len(projects))
	}
}

func TestDiscoverSkipsNodeModules(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("CSQ_REAL_HOME", tmpDir)

	nmRepo := filepath.Join(tmpDir, "code", "node_modules", "some-pkg", ".git")
	if err := os.MkdirAll(nmRepo, 0755); err != nil {
		t.Fatal(err)
	}

	realRepo := filepath.Join(tmpDir, "code", "real-project", ".git")
	if err := os.MkdirAll(realRepo, 0755); err != nil {
		t.Fatal(err)
	}

	cfg := config.Config{
		Roots:    []string{filepath.Join(tmpDir, "code")},
		MaxDepth: 3,
	}

	projects, err := scan(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 project (skipping node_modules), got %d", len(projects))
	}
}

func TestFuzzyMatch(t *testing.T) {
	projects := []Project{
		{Name: "acme/api", Path: "/code/acme/api"},
		{Name: "acme/web-app", Path: "/code/acme/web-app"},
		{Name: "personal/core-utils", Path: "/code/personal/core-utils"},
	}

	tests := []struct {
		query    string
		expected int
	}{
		{"", 3},
		{"core", 1},
		{"api", 1},
		{"nonexistent", 0},
		{"acme", 2},
	}

	for _, tt := range tests {
		matches := FuzzyMatch(projects, tt.query)
		if len(matches) != tt.expected {
			t.Errorf("FuzzyMatch(%q): expected %d results, got %d", tt.query, tt.expected, len(matches))
		}
	}
}

func TestCaching(t *testing.T) {
	_, cfg := setupTestRepos(t)

	projects1, err := Discover(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// Second call should use cache
	projects2, err := Discover(cfg)
	if err != nil {
		t.Fatal(err)
	}

	if len(projects1) != len(projects2) {
		t.Errorf("cache mismatch: %d vs %d", len(projects1), len(projects2))
	}
}
