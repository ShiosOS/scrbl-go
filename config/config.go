package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	NotesDir  string `mapstructure:"notes_dir"`
	ServerURL string `mapstructure:"server_url"`
	APIKey    string `mapstructure:"api_key"`
	Editor    string `mapstructure:"editor"`
}

// DefaultDir returns the default scrbl config/data directory.
func DefaultDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".scrbl")
}

// DefaultNotesDir returns the default notes directory.
func DefaultNotesDir() string {
	return filepath.Join(DefaultDir(), "notes")
}

// Load reads the config file and returns a Config.
func Load() (*Config, error) {
	configDir := DefaultDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, err
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configDir)

	// Defaults
	viper.SetDefault("notes_dir", DefaultNotesDir())
	viper.SetDefault("server_url", "")
	viper.SetDefault("api_key", "")
	viper.SetDefault("editor", "nvim")

	// Read config file (ignore error if it doesn't exist)
	_ = viper.ReadInConfig()

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save writes the current config values to the config file.
func Save(cfg *Config) error {
	configDir := DefaultDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	viper.Set("notes_dir", cfg.NotesDir)
	viper.Set("server_url", cfg.ServerURL)
	viper.Set("api_key", cfg.APIKey)
	viper.Set("editor", cfg.Editor)

	return viper.WriteConfigAs(filepath.Join(configDir, "config.yaml"))
}

// Exists returns true if a config file already exists.
func Exists() bool {
	path := filepath.Join(DefaultDir(), "config.yaml")
	_, err := os.Stat(path)
	return err == nil
}
