package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultsArePopulated(t *testing.T) {
	cfg := defaults()

	if cfg.DefaultSort == "" {
		t.Error("default sort should not be empty")
	}
	if cfg.Theme == "" {
		t.Error("theme should not be empty")
	}
	if len(cfg.IgnoreDirs) == 0 {
		t.Error("ignore_dirs should not be empty")
	}
}

func TestLoadOrCreate_CreatesFile(t *testing.T) {
	// Point viper / FilePath at a temp dir.
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmp)
	defer func() { os.Setenv("HOME", origHome) }()

	cfg, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	// File should now exist.
	cfgPath := filepath.Join(tmp, ".config", "gtree", "config.yml")
	if _, err := os.Stat(cfgPath); err != nil {
		t.Errorf("config file was not created: %v", err)
	}
}

func TestLoadOrCreate_ReadsExisting(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	t.Setenv("HOME", tmp)
	defer func() { os.Setenv("HOME", origHome) }()

	// Write a custom config.
	dir := filepath.Join(tmp, ".config", "gtree")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "default_sort: dirty\ntheme: light\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadOrCreate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DefaultSort != "dirty" {
		t.Errorf("expected sort=dirty, got %q", cfg.DefaultSort)
	}
	if cfg.Theme != "light" {
		t.Errorf("expected theme=light, got %q", cfg.Theme)
	}
	// Fields not in YAML should keep defaults.
	if len(cfg.IgnoreDirs) == 0 {
		t.Error("ignore_dirs should fall back to defaults")
	}
}

func TestIgnoreDirList(t *testing.T) {
	cfg := defaults()
	ignoreSet := make(map[string]bool)
	for _, d := range cfg.IgnoreDirs {
		ignoreSet[d] = true
	}
	for _, mustIgnore := range []string{"node_modules", "vendor", ".venv"} {
		if !ignoreSet[mustIgnore] {
			t.Errorf("%q should be in default ignore list", mustIgnore)
		}
	}
}
