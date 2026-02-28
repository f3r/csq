package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	claudeSquadDir = ".claude-squad"
	stateFile      = "state.json"
)

type Instance struct {
	Title  string `json:"title"`
	Status string `json:"status"`
	Branch string `json:"branch"`
	Path   string `json:"worktreePath"`
}

type SessionInfo struct {
	ProjectName string
	HomeDir     string
	Instances   []Instance
}

func ReadSessions(homeDir string) ([]Instance, error) {
	path := filepath.Join(homeDir, claudeSquadDir, stateFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading state: %w", err)
	}

	var instances []Instance
	if err := json.Unmarshal(data, &instances); err != nil {
		return nil, fmt.Errorf("parsing state: %w", err)
	}

	return instances, nil
}

func CountSessions(homeDir string) int {
	instances, err := ReadSessions(homeDir)
	if err != nil {
		return 0
	}
	return len(instances)
}

func AllSessions(homeBase string) ([]SessionInfo, error) {
	entries, err := os.ReadDir(homeBase)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading homes dir: %w", err)
	}

	var sessions []SessionInfo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		homeDir := filepath.Join(homeBase, e.Name())
		instances, err := ReadSessions(homeDir)
		if err != nil {
			continue
		}
		if len(instances) > 0 {
			sessions = append(sessions, SessionInfo{
				ProjectName: e.Name(),
				HomeDir:     homeDir,
				Instances:   instances,
			})
		}
	}

	return sessions, nil
}
