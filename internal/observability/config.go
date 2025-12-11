package observability

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// UserConfig holds user preferences stored in ~/.config/yapi/config.json
type UserConfig struct {
	FileLoggingEnabled *bool `json:"file_logging_enabled,omitempty"`
	PosthogEnabled     *bool `json:"posthog_enabled,omitempty"`
}

// yapiConfigDir returns the yapi config directory (~/.config/yapi)
func yapiConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "yapi"), nil
}

// configPath returns the path to the user config file
func configPath() (string, error) {
	dir, err := yapiConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.json"), nil
}

// LoadUserConfig loads the user config from disk
func LoadUserConfig() (*UserConfig, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &UserConfig{}, nil
		}
		return nil, err
	}

	var cfg UserConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return &UserConfig{}, nil // Return empty config on parse error
	}

	return &cfg, nil
}

// SaveUserConfig saves the user config to disk
func SaveUserConfig(cfg *UserConfig) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// IsFirstRun returns true if this is the first time yapi is being run
// (no config file exists and preferences haven't been set)
func IsFirstRun() bool {
	cfg, err := LoadUserConfig()
	if err != nil {
		return true
	}
	return cfg.FileLoggingEnabled == nil && cfg.PosthogEnabled == nil
}

// SetFileLoggingEnabled saves the user's file logging preference
func SetFileLoggingEnabled(enabled bool) error {
	cfg, err := LoadUserConfig()
	if err != nil {
		cfg = &UserConfig{}
	}
	cfg.FileLoggingEnabled = &enabled
	return SaveUserConfig(cfg)
}

// SetPosthogEnabled saves the user's PostHog preference
func SetPosthogEnabled(enabled bool) error {
	cfg, err := LoadUserConfig()
	if err != nil {
		cfg = &UserConfig{}
	}
	cfg.PosthogEnabled = &enabled
	return SaveUserConfig(cfg)
}

// GetFileLoggingPreference returns the user's file logging preference.
// Returns nil if not yet set (first run).
func GetFileLoggingPreference() *bool {
	cfg, err := LoadUserConfig()
	if err != nil {
		return nil
	}
	return cfg.FileLoggingEnabled
}

// GetPosthogPreference returns the user's PostHog preference.
// Returns nil if not yet set (first run).
func GetPosthogPreference() *bool {
	cfg, err := LoadUserConfig()
	if err != nil {
		return nil
	}
	return cfg.PosthogEnabled
}
