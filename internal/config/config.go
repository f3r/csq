package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	CsqDir     = ".csq"
	ConfigFile = "config.json"
)

type Config struct {
	Roots           []string `json:"roots"`
	MaxDepth        int      `json:"max_depth"`
	CSBinary        string   `json:"cs_binary"`
	HomeBase        string   `json:"home_base"`
	SymlinkDotfiles []string `json:"symlink_dotfiles"`
}

func DefaultConfig() Config {
	home := RealHome()
	return Config{
		Roots:    []string{filepath.Join(home, "code")},
		MaxDepth: 3,
		CSBinary: "cs",
		HomeBase: filepath.Join(home, CsqDir, "homes"),
		SymlinkDotfiles: []string{
			".gitconfig", ".ssh", ".claude", ".config",
			".zshrc", ".zshenv", "bin", ".local", ".nvm", ".cargo", ".aws",
		},
	}
}

func CsqDirPath() string {
	return filepath.Join(RealHome(), CsqDir)
}

func ConfigPath() string {
	return filepath.Join(CsqDirPath(), ConfigFile)
}

func RealHome() string {
	if h := os.Getenv("CSQ_REAL_HOME"); h != "" {
		return h
	}
	h, err := os.UserHomeDir()
	if err != nil {
		h = os.Getenv("HOME")
	}
	return h
}

func Load() (Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return Config{}, fmt.Errorf("reading config: %w", err)
	}

	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parsing config: %w", err)
	}

	return cfg, nil
}

func Save(cfg Config) error {
	dir := CsqDirPath()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating csq dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	return os.WriteFile(ConfigPath(), data, 0644)
}

func ExpandPath(p string) string {
	if len(p) > 0 && p[0] == '~' {
		return filepath.Join(RealHome(), p[1:])
	}
	return p
}

func (c Config) ExpandedRoots() []string {
	roots := make([]string, len(c.Roots))
	for i, r := range c.Roots {
		roots[i] = ExpandPath(r)
	}
	return roots
}
