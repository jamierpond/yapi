package telemetry

import (
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// getMachineID returns a persistent unique identifier for this machine.
// On first run, generates a UUID and saves it to ~/.config/yapi/machine_id.
// Subsequent runs read the existing ID.
func getMachineID() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fallbackID()
	}

	idFile := filepath.Join(configDir, "yapi", "machine_id")

	// Try reading existing ID
	if data, err := os.ReadFile(idFile); err == nil && len(data) > 0 {
		return string(data)
	}

	// Generate new ID
	id := uuid.New().String()

	// Best effort save (directory should exist since config lives there)
	_ = os.MkdirAll(filepath.Dir(idFile), 0755)
	_ = os.WriteFile(idFile, []byte(id), 0644)

	return id
}

// fallbackID returns a fallback identifier when config dir is unavailable
func fallbackID() string {
	if hostname, err := os.Hostname(); err == nil && hostname != "" {
		return hostname
	}
	return "unknown"
}
