package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/f3r/csq/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config [key] [value]",
	Short: "Show or edit configuration",
	Long:  "With no args: show config. With key: show value. With key+value: set value.",
	RunE:  runConfig,
}

func init() {
	rootCmd.AddCommand(configCmd)
}

func runConfig(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	switch len(args) {
	case 0:
		return showConfig(cfg)
	case 1:
		return showConfigKey(cfg, args[0])
	case 2:
		return setConfigKey(&cfg, args[0], args[1])
	default:
		return fmt.Errorf("usage: csq config [key] [value]")
	}
}

func showConfig(cfg config.Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func showConfigKey(cfg config.Config, key string) error {
	data, _ := json.Marshal(cfg)
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}

	val, ok := m[key]
	if !ok {
		return fmt.Errorf("unknown config key: %s", key)
	}

	switch v := val.(type) {
	case []interface{}:
		for _, item := range v {
			fmt.Println(item)
		}
	default:
		fmt.Println(v)
	}
	return nil
}

func setConfigKey(cfg *config.Config, key, value string) error {
	switch key {
	case "max_depth":
		n, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("max_depth must be an integer")
		}
		cfg.MaxDepth = n
	case "cs_binary":
		cfg.CSBinary = value
	case "home_base":
		cfg.HomeBase = value
	case "roots":
		cfg.Roots = strings.Split(value, ",")
	default:
		return fmt.Errorf("cannot set key %q (supported: roots, max_depth, cs_binary, home_base)", key)
	}

	if err := config.Save(*cfg); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "Set %s = %s\n", key, value)
	return nil
}
