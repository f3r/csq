package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestReadSessionsEmpty(t *testing.T) {
	tmpDir := t.TempDir()

	instances, err := ReadSessions(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if instances != nil {
		t.Errorf("expected nil for missing state, got %v", instances)
	}
}

func TestReadSessions(t *testing.T) {
	tmpDir := t.TempDir()
	csDir := filepath.Join(tmpDir, ".claude-squad")
	if err := os.MkdirAll(csDir, 0755); err != nil {
		t.Fatal(err)
	}

	instances := []Instance{
		{Title: "fix auth", Status: "running", Branch: "cs/fix-auth"},
		{Title: "add tests", Status: "paused", Branch: "cs/add-tests"},
	}

	data, _ := json.Marshal(instances)
	if err := os.WriteFile(filepath.Join(csDir, "state.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	result, err := ReadSessions(tmpDir)
	if err != nil {
		t.Fatalf("ReadSessions failed: %v", err)
	}

	if len(result) != 2 {
		t.Errorf("expected 2 instances, got %d", len(result))
	}
	if result[0].Title != "fix auth" {
		t.Errorf("expected 'fix auth', got %q", result[0].Title)
	}
}

func TestCountSessions(t *testing.T) {
	tmpDir := t.TempDir()

	if count := CountSessions(tmpDir); count != 0 {
		t.Errorf("expected 0 for missing state, got %d", count)
	}

	csDir := filepath.Join(tmpDir, ".claude-squad")
	if err := os.MkdirAll(csDir, 0755); err != nil {
		t.Fatal(err)
	}

	instances := []Instance{{Title: "test", Status: "running"}}
	data, _ := json.Marshal(instances)
	if err := os.WriteFile(filepath.Join(csDir, "state.json"), data, 0644); err != nil {
		t.Fatal(err)
	}

	if count := CountSessions(tmpDir); count != 1 {
		t.Errorf("expected 1, got %d", count)
	}
}

func TestAllSessions(t *testing.T) {
	tmpDir := t.TempDir()

	// Create two project homes with sessions
	for _, name := range []string{"proj-a", "proj-b"} {
		csDir := filepath.Join(tmpDir, name, ".claude-squad")
		if err := os.MkdirAll(csDir, 0755); err != nil {
			t.Fatal(err)
		}
		instances := []Instance{{Title: name + " session", Status: "running"}}
		data, _ := json.Marshal(instances)
		if err := os.WriteFile(filepath.Join(csDir, "state.json"), data, 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Empty project (no sessions)
	if err := os.MkdirAll(filepath.Join(tmpDir, "proj-c"), 0755); err != nil {
		t.Fatal(err)
	}

	sessions, err := AllSessions(tmpDir)
	if err != nil {
		t.Fatalf("AllSessions failed: %v", err)
	}

	if len(sessions) != 2 {
		t.Errorf("expected 2 projects with sessions, got %d", len(sessions))
	}
}
