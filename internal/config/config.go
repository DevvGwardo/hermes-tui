package config

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

// Config holds the user's preferences and session state.
type Config struct {
	Theme       string `json:"theme"`       // "ocean", "amber", "rose", "forest", "aquarium"
	SessionID   string `json:"session_id"`  // last active session key
	Thinking    bool   `json:"thinking"`    // show thinking indicator
	HistorySize int    `json:"history_size"` // number of history entries to keep
}

// Default returns a new Config with sensible defaults.
func Default() Config {
	return Config{
		Theme:       "ocean",
		SessionID:   "",
		Thinking:    true,
		HistorySize: 100,
	}
}

// Path returns the config file path: ~/.config/hermes-tui/config.json
func configPath() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(usr.HomeDir, ".config", "hermes-tui")
	return filepath.Join(dir, "config.json"), nil
}

// Load reads the config file. Returns Default() if the file doesn't exist or is invalid.
func Load() (Config, error) {
	path, err := configPath()
	if err != nil {
		return Default(), err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return Default(), nil // treat missing file as defaults
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Default(), err
	}
	// Ensure defaults for missing fields
	if cfg.Theme == "" {
		cfg.Theme = "ocean"
	}
	if cfg.HistorySize == 0 {
		cfg.HistorySize = 100
	}
	return cfg, nil
}

// Save writes the config to the config file.
func (c *Config) Save() error {
	path, err := configPath()
	if err != nil {
		return err
	}
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
