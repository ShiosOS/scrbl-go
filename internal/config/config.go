package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultConfigFile = "config.json"
	legacyConfigFile  = "config.yaml"
)

type Config struct {
	NotesDir  string `json:"notes_dir"`
	ServerURL string `json:"server_url"`
	APIKey    string `json:"api_key"`
}

func Load() (Config, error) {
	cfg := Config{NotesDir: DefaultNotesDir()}

	b, err := os.ReadFile(Path())
	if err != nil {
		if os.IsNotExist(err) {
			legacyCfg, ok, legacyErr := loadLegacyConfig()
			if legacyErr != nil {
				return Config{}, legacyErr
			}
			if ok {
				if strings.TrimSpace(legacyCfg.NotesDir) != "" {
					cfg.NotesDir = legacyCfg.NotesDir
				}
				cfg.ServerURL = legacyCfg.ServerURL
				cfg.APIKey = legacyCfg.APIKey
			}
			return finalize(cfg), nil
		}
		return Config{}, fmt.Errorf("read config: %w", err)
	}

	if len(strings.TrimSpace(string(b))) == 0 {
		return finalize(cfg), nil
	}

	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}

	return finalize(cfg), nil
}

func Save(cfg Config) error {
	cfg = finalize(cfg)

	dir := filepath.Dir(Path())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}

	if err := os.WriteFile(Path(), b, 0o644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

func Path() string {
	if p := strings.TrimSpace(os.Getenv("SCRBL_CONFIG")); p != "" {
		return expandPath(p)
	}
	return filepath.Join(baseDir(), defaultConfigFile)
}

func DefaultNotesDir() string {
	return filepath.Join(baseDir(), "notes")
}

func loadLegacyConfig() (Config, bool, error) {
	legacyPath := filepath.Join(baseDir(), legacyConfigFile)
	b, err := os.ReadFile(legacyPath)
	if err != nil {
		if os.IsNotExist(err) {
			return Config{}, false, nil
		}
		return Config{}, false, fmt.Errorf("read legacy config: %w", err)
	}

	cfg := Config{}
	content := strings.ReplaceAll(string(b), "\r\n", "\n")

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		val = strings.Trim(val, `"'`)

		switch key {
		case "notes_dir":
			cfg.NotesDir = val
		case "server_url":
			cfg.ServerURL = val
		case "api_key":
			cfg.APIKey = val
		}
	}

	return cfg, true, nil
}

func finalize(cfg Config) Config {
	if strings.TrimSpace(cfg.NotesDir) == "" {
		cfg.NotesDir = DefaultNotesDir()
	}

	cfg.NotesDir = expandPath(cfg.NotesDir)
	cfg.ServerURL = strings.TrimRight(strings.TrimSpace(cfg.ServerURL), "/")
	cfg.APIKey = strings.TrimSpace(cfg.APIKey)

	return cfg
}

func baseDir() string {
	return filepath.Join(homeDir(), ".scrbl")
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil || strings.TrimSpace(home) == "" {
		return "."
	}
	return home
}

func expandPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}
	if p == "~" {
		return homeDir()
	}
	if strings.HasPrefix(p, "~/") || strings.HasPrefix(p, `~\`) {
		return filepath.Join(homeDir(), p[2:])
	}
	return p
}
