package launcher

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/f3r/csq/internal/config"
)

func SanitizeName(name string) string {
	r := strings.NewReplacer("/", "--", " ", "_", ".", "_")
	return r.Replace(name)
}

func UnsanitizeName(sanitized string) string {
	r := strings.NewReplacer("--", "/", "_", ".")
	return r.Replace(sanitized)
}

func HomeDir(projectName string, cfg config.Config) string {
	homeBase := config.ExpandPath(cfg.HomeBase)
	return filepath.Join(homeBase, SanitizeName(projectName))
}

func EnsureHome(projectName string, cfg config.Config) (string, error) {
	homeDir := HomeDir(projectName, cfg)
	realHome := config.RealHome()

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		return "", fmt.Errorf("creating home dir: %w", err)
	}

	for _, dotfile := range cfg.SymlinkDotfiles {
		src := filepath.Join(realHome, dotfile)
		dst := filepath.Join(homeDir, dotfile)

		if _, err := os.Lstat(dst); err == nil {
			continue
		}

		if _, err := os.Stat(src); os.IsNotExist(err) {
			continue
		}

		if err := os.Symlink(src, dst); err != nil {
			return "", fmt.Errorf("symlinking %s: %w", dotfile, err)
		}
	}

	if err := ensureClaudeSquadConfig(homeDir, realHome); err != nil {
		return "", fmt.Errorf("setting up claude-squad config: %w", err)
	}

	return homeDir, nil
}

func ensureClaudeSquadConfig(homeDir, realHome string) error {
	csDir := filepath.Join(homeDir, ".claude-squad")
	if err := os.MkdirAll(csDir, 0755); err != nil {
		return err
	}

	srcConfig := filepath.Join(realHome, ".claude-squad", "config.json")
	dstConfig := filepath.Join(csDir, "config.json")

	if _, err := os.Stat(dstConfig); err == nil {
		return nil
	}

	data, err := os.ReadFile(srcConfig)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return os.WriteFile(dstConfig, data, 0644)
}

func Launch(projectName, projectPath string, cfg config.Config, csArgs []string) error {
	homeDir, err := EnsureHome(projectName, cfg)
	if err != nil {
		return fmt.Errorf("ensuring home: %w", err)
	}

	csBinary, err := exec.LookPath(config.ExpandPath(cfg.CSBinary))
	if err != nil {
		return fmt.Errorf("cs binary not found: %w", err)
	}

	if err := os.Chdir(projectPath); err != nil {
		return fmt.Errorf("changing to project dir: %w", err)
	}

	env := buildEnv(homeDir, projectPath)
	args := append([]string{csBinary}, csArgs...)

	return syscall.Exec(csBinary, args, env)
}

func buildEnv(homeDir, projectPath string) []string {
	realHome := config.RealHome()
	var env []string

	for _, e := range os.Environ() {
		key := e[:strings.IndexByte(e, '=')]
		switch key {
		case "HOME", "PWD":
			continue
		default:
			env = append(env, e)
		}
	}

	env = append(env,
		"HOME="+homeDir,
		"PWD="+projectPath,
		"CSQ_REAL_HOME="+realHome,
		"CSQ_PROJECT="+filepath.Base(projectPath),
		"GH_CONFIG_DIR="+filepath.Join(realHome, ".config", "gh"),
	)

	return env
}
