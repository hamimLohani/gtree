// Package config handles loading and creating the gtree configuration file.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds all gtree configuration values.
type Config struct {
	DefaultPath  string   `yaml:"default_path"`
	DefaultDepth int      `yaml:"default_depth"`
	DefaultSort  string   `yaml:"default_sort"`
	Theme        string   `yaml:"theme"`
	IgnoreDirs   []string `yaml:"ignore_dirs"`
}

// defaultIgnoreDirs lists directories that are never useful to scan for git repos.
var defaultIgnoreDirs = []string{
	"node_modules",
	".cache",
	"vendor",
	"target",
	"dist",
	"build",
	".venv",
	"Library",
	".Trash",
	"proc",
	"sys",
	"dev",
}

// defaults returns a Config populated with sensible starting values.
func defaults() Config {
	return Config{
		DefaultPath:  "",
		DefaultDepth: 0, // 0 = unlimited
		DefaultSort:  "name",
		Theme:        "auto",
		IgnoreDirs:   defaultIgnoreDirs,
	}
}

// FilePath returns the canonical path to the config file.
func FilePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "gtree", "config.yml")
}

// LoadOrCreate reads the config file, creating it with defaults if it doesn't exist.
func LoadOrCreate() (*Config, error) {
	path := FilePath()

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		cfg := defaults()
		if writeErr := write(path, &cfg); writeErr != nil {
			// Non-fatal: just use defaults in memory.
			fmt.Fprintf(os.Stderr, "note: could not create config file at %s: %v\n", path, writeErr)
		}
		return &cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	// Start with defaults so missing keys get sensible values.
	cfg := defaults()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}
	return &cfg, nil
}

// write serialises cfg as YAML to path, creating parent directories if needed.
func write(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	header := []byte("# gtree configuration file\n# https://github.com/hamimlohani/gtree\n\n")
	return os.WriteFile(path, append(header, data...), 0o644)
}
