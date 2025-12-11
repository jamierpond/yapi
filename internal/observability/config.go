package telemetry

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// UserConfig holds user preferences stored in ~/.config/yapi/config.json
type UserConfig struct {
	TelemetryEnabled *bool `json:"telemetry_enabled,omitempty"`
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
// (no config file exists and telemetry preference hasn't been set)
func IsFirstRun() bool {
	cfg, err := LoadUserConfig()
	if err != nil {
		return true
	}
	return cfg.TelemetryEnabled == nil
}

// SetTelemetryEnabled saves the user's telemetry preference
func SetTelemetryEnabled(enabled bool) error {
	cfg, err := LoadUserConfig()
	if err != nil {
		cfg = &UserConfig{}
	}
	cfg.TelemetryEnabled = &enabled
	return SaveUserConfig(cfg)
}

// GetTelemetryPreference returns the user's telemetry preference.
// Returns nil if not yet set (first run).
func GetTelemetryPreference() *bool {
	cfg, err := LoadUserConfig()
	if err != nil {
		return nil
	}
	return cfg.TelemetryEnabled
}
