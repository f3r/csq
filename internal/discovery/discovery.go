package discovery

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/f3r/csq/internal/config"
)

type Project struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type cache struct {
	Projects  []Project `json:"projects"`
	Timestamp time.Time `json:"timestamp"`
}

const cacheTTL = 5 * time.Minute

func cachePath() string {
	return filepath.Join(config.CsqDirPath(), "cache.json")
}

func Discover(cfg config.Config) ([]Project, error) {
	if projects, ok := loadCache(); ok {
		return projects, nil
	}

	projects, err := scan(cfg)
	if err != nil {
		return nil, err
	}

	saveCache(projects)
	return projects, nil
}

func DiscoverFresh(cfg config.Config) ([]Project, error) {
	projects, err := scan(cfg)
	if err != nil {
		return nil, err
	}
	saveCache(projects)
	return projects, nil
}

func scan(cfg config.Config) ([]Project, error) {
	var projects []Project
	seen := make(map[string]bool)

	for _, root := range cfg.ExpandedRoots() {
		root = filepath.Clean(root)
		err := walkForGitRepos(root, root, cfg.MaxDepth, 0, func(p Project) {
			if !seen[p.Path] {
				seen[p.Path] = true
				projects = append(projects, p)
			}
		})
		if err != nil {
			return nil, fmt.Errorf("scanning %s: %w", root, err)
		}
	}

	sort.Slice(projects, func(i, j int) bool {
		return projects[i].Name < projects[j].Name
	})

	return projects, nil
}

func walkForGitRepos(root, dir string, maxDepth, currentDepth int, fn func(Project)) error {
	if currentDepth > maxDepth {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsPermission(err) {
			return nil
		}
		return err
	}

	hasGit := false
	for _, e := range entries {
		if e.Name() == ".git" {
			hasGit = true
			break
		}
	}

	if hasGit {
		name := relativeName(root, dir)
		fn(Project{Name: name, Path: dir})
		return nil
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" {
			continue
		}
		if err := walkForGitRepos(root, filepath.Join(dir, name), maxDepth, currentDepth+1, fn); err != nil {
			return err
		}
	}

	return nil
}

func relativeName(root, dir string) string {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return filepath.Base(dir)
	}
	return rel
}

func loadCache() ([]Project, bool) {
	data, err := os.ReadFile(cachePath())
	if err != nil {
		return nil, false
	}

	var c cache
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, false
	}

	if time.Since(c.Timestamp) > cacheTTL {
		return nil, false
	}

	return c.Projects, true
}

func saveCache(projects []Project) {
	c := cache{
		Projects:  projects,
		Timestamp: time.Now(),
	}
	data, _ := json.MarshalIndent(c, "", "  ")
	_ = os.MkdirAll(filepath.Dir(cachePath()), 0755)
	_ = os.WriteFile(cachePath(), data, 0644)
}

func FuzzyMatch(projects []Project, query string) []Project {
	if query == "" {
		return projects
	}

	query = strings.ToLower(query)
	type scored struct {
		project Project
		score   int
	}

	var matches []scored
	for _, p := range projects {
		name := strings.ToLower(p.Name)
		if strings.Contains(name, query) {
			score := 100 - len(name)
			if strings.HasSuffix(name, query) || strings.HasSuffix(name, "/"+query) {
				score += 50
			}
			if name == query {
				score += 100
			}
			matches = append(matches, scored{project: p, score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].score > matches[j].score
	})

	result := make([]Project, len(matches))
	for i, m := range matches {
		result[i] = m.project
	}
	return result
}
