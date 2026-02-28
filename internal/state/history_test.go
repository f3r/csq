package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/f3r/csq/internal/discovery"
)

func TestRecordLaunch_CreatesFileIfMissing(t *testing.T) {
	dir := t.TempDir()

	if err := RecordLaunch(dir, "acme/api"); err != nil {
		t.Fatalf("RecordLaunch failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, historyFile))
	if err != nil {
		t.Fatalf("reading history file: %v", err)
	}

	var history map[string]time.Time
	if err := json.Unmarshal(data, &history); err != nil {
		t.Fatalf("parsing history: %v", err)
	}

	if _, ok := history["acme/api"]; !ok {
		t.Fatal("expected acme/api in history")
	}
}

func TestRecordLaunch_UpdatesExistingEntry(t *testing.T) {
	dir := t.TempDir()

	if err := RecordLaunch(dir, "acme/api"); err != nil {
		t.Fatalf("first RecordLaunch failed: %v", err)
	}

	first := readTimestamp(t, dir, "acme/api")

	// Ensure time progresses
	time.Sleep(10 * time.Millisecond)

	if err := RecordLaunch(dir, "acme/api"); err != nil {
		t.Fatalf("second RecordLaunch failed: %v", err)
	}

	second := readTimestamp(t, dir, "acme/api")

	if !second.After(first) {
		t.Fatalf("expected second timestamp (%v) to be after first (%v)", second, first)
	}
}

func TestRecordLaunch_PreservesOtherEntries(t *testing.T) {
	dir := t.TempDir()

	if err := RecordLaunch(dir, "acme/api"); err != nil {
		t.Fatalf("RecordLaunch failed: %v", err)
	}
	if err := RecordLaunch(dir, "org/web"); err != nil {
		t.Fatalf("RecordLaunch failed: %v", err)
	}

	history := readHistory(t, dir)
	if len(history) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(history))
	}
	if _, ok := history["acme/api"]; !ok {
		t.Fatal("expected acme/api in history")
	}
	if _, ok := history["org/web"]; !ok {
		t.Fatal("expected org/web in history")
	}
}

func TestSortByRecency_RecentFirst(t *testing.T) {
	dir := t.TempDir()

	now := time.Now()
	history := map[string]time.Time{
		"old-project": now.Add(-2 * time.Hour),
		"new-project": now.Add(-1 * time.Minute),
	}
	writeHistory(t, dir, history)

	projects := []discovery.Project{
		{Name: "old-project", Path: "/old"},
		{Name: "new-project", Path: "/new"},
		{Name: "unknown", Path: "/unknown"},
	}

	sorted := SortByRecency(projects, dir)

	expected := []string{"new-project", "old-project", "unknown"}
	for i, name := range expected {
		if sorted[i].Name != name {
			t.Errorf("position %d: expected %s, got %s", i, name, sorted[i].Name)
		}
	}
}

func TestSortByRecency_UnknownsAlphabetical(t *testing.T) {
	dir := t.TempDir()

	projects := []discovery.Project{
		{Name: "zebra", Path: "/z"},
		{Name: "alpha", Path: "/a"},
		{Name: "middle", Path: "/m"},
	}

	sorted := SortByRecency(projects, dir)

	expected := []string{"alpha", "middle", "zebra"}
	for i, name := range expected {
		if sorted[i].Name != name {
			t.Errorf("position %d: expected %s, got %s", i, name, sorted[i].Name)
		}
	}
}

func TestSortByRecency_DoesNotMutateInput(t *testing.T) {
	dir := t.TempDir()

	now := time.Now()
	writeHistory(t, dir, map[string]time.Time{"b": now})

	projects := []discovery.Project{
		{Name: "a", Path: "/a"},
		{Name: "b", Path: "/b"},
	}

	sorted := SortByRecency(projects, dir)

	if sorted[0].Name != "b" {
		t.Fatal("sorted should have b first")
	}
	if projects[0].Name != "a" {
		t.Fatal("original slice should not be mutated")
	}
}

func readTimestamp(t *testing.T, dir, name string) time.Time {
	t.Helper()
	history := readHistory(t, dir)
	ts, ok := history[name]
	if !ok {
		t.Fatalf("expected %s in history", name)
	}
	return ts
}

func readHistory(t *testing.T, dir string) map[string]time.Time {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, historyFile))
	if err != nil {
		t.Fatalf("reading history: %v", err)
	}
	var history map[string]time.Time
	if err := json.Unmarshal(data, &history); err != nil {
		t.Fatalf("parsing history: %v", err)
	}
	return history
}

func writeHistory(t *testing.T, dir string, history map[string]time.Time) {
	t.Helper()
	data, err := json.Marshal(history)
	if err != nil {
		t.Fatalf("marshaling history: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, historyFile), data, 0644); err != nil {
		t.Fatalf("writing history: %v", err)
	}
}
