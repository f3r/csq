package bootstrap

import (
	_ "embed"
	"fmt"
	"os"
)

//go:embed bootstrap.sh.tmpl
var bootstrapScript string

func WriteScript(path string) error {
	if err := os.WriteFile(path, []byte(bootstrapScript), 0755); err != nil {
		return fmt.Errorf("writing bootstrap script: %w", err)
	}
	return nil
}

func Script() string {
	return bootstrapScript
}
