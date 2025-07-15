package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config holds user configuration settings
type Config struct {
	// Editor settings
	TabSize     int  `json:"tab_size"`
	WordWrap    int  `json:"word_wrap"`
	ShowNumbers bool `json:"show_line_numbers"`

	// Theme settings
	Theme      string `json:"theme"`
	DarkMode   bool   `json:"dark_mode"`

	// Behavior settings
	AutoSave   bool `json:"auto_save"`
	BlinkRate  int  `json:"cursor_blink_rate_ms"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		TabSize:     4,
		WordWrap:    DefaultWordWrap,
		ShowNumbers: false,
		Theme:       "auto",
		DarkMode:    true,
		AutoSave:    false,
		BlinkRate:   500,
	}
}

// LoadConfig loads configuration from the user's home directory
func LoadConfig() Config {
	config := DefaultConfig()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return config
	}

	configPath := filepath.Join(homeDir, ".config", "hani", "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		// Config file doesn't exist, return defaults
		return config
	}

	if err := json.Unmarshal(data, &config); err != nil {
		// Invalid config file, return defaults
		return config
	}

	return config
}

// SaveConfig saves the configuration to the user's home directory
func SaveConfig(config Config) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".config", "hani")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}
