package observability

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// UserConfig holds user preferences stored in ~/.config/yapi/config.json
type UserConfig struct {
	LoggingEnabled *bool `json:"logging_enabled,omitempty"`
}

// yapiConfigDir returns the yapi config directory (~/.yapi)
func yapiConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".yapi"), nil
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
		return &UserConfig{}, nil
	}

	return &cfg, nil
}

// SaveUserConfig saves the user config to disk
func SaveUserConfig(cfg *UserConfig) error {
	path, err := configPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// IsEnabled returns true if logging is enabled (default: true)
func IsEnabled() bool {
	cfg, err := LoadUserConfig()
	if err != nil || cfg.LoggingEnabled == nil {
		return true // Default to enabled
	}
	return *cfg.LoggingEnabled
}

// SetEnabled saves the logging preference
func SetEnabled(enabled bool) error {
	cfg, err := LoadUserConfig()
	if err != nil {
		cfg = &UserConfig{}
	}
	cfg.LoggingEnabled = &enabled
	return SaveUserConfig(cfg)
}
