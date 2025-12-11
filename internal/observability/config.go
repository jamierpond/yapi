package observability

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Telemetry mode constants matching Go's philosophy
const (
	ModeOff   = "off"   // Completely disabled
	ModeLocal = "local" // Default: write to file, do not upload
	ModeOn    = "on"    // Upload enabled
)

// UserConfig holds user preferences stored in ~/.config/yapi/config.json
type UserConfig struct {
	TelemetryMode string `json:"telemetry_mode,omitempty"`
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
// (no config file exists or telemetry mode hasn't been set)
func IsFirstRun() bool {
	cfg, err := LoadUserConfig()
	if err != nil {
		return true
	}
	return cfg.TelemetryMode == ""
}

// GetMode returns the current telemetry mode, defaulting to "local"
func GetMode() string {
	cfg, err := LoadUserConfig()
	if err != nil || cfg.TelemetryMode == "" {
		return ModeLocal // Go vibe: default to local collection
	}
	return cfg.TelemetryMode
}

// SetMode saves the user's telemetry mode preference
func SetMode(mode string) error {
	cfg, err := LoadUserConfig()
	if err != nil {
		cfg = &UserConfig{}
	}
	cfg.TelemetryMode = mode
	return SaveUserConfig(cfg)
}
