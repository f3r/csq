package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/f3r/csq/internal/discovery"
)

const historyFile = "history.json"

func historyPath(csqDir string) string {
	return filepath.Join(csqDir, historyFile)
}

func loadHistory(csqDir string) (map[string]time.Time, error) {
	data, err := os.ReadFile(historyPath(csqDir))
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]time.Time), nil
		}
		return nil, err
	}

	var history map[string]time.Time
	if err := json.Unmarshal(data, &history); err != nil {
		return make(map[string]time.Time), nil
	}

	return history, nil
}

func saveHistory(csqDir string, history map[string]time.Time) error {
	if err := os.MkdirAll(csqDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyPath(csqDir), data, 0644)
}

func RecordLaunch(csqDir, projectName string) error {
	history, err := loadHistory(csqDir)
	if err != nil {
		return err
	}

	history[projectName] = time.Now()

	return saveHistory(csqDir, history)
}

func SortByRecency(projects []discovery.Project, csqDir string) []discovery.Project {
	history, err := loadHistory(csqDir)
	if err != nil {
		history = make(map[string]time.Time)
	}

	sorted := make([]discovery.Project, len(projects))
	copy(sorted, projects)

	sort.SliceStable(sorted, func(i, j int) bool {
		ti, oki := history[sorted[i].Name]
		tj, okj := history[sorted[j].Name]

		switch {
		case oki && okj:
			return ti.After(tj)
		case oki:
			return true
		case okj:
			return false
		default:
			return sorted[i].Name < sorted[j].Name
		}
	})

	return sorted
}
