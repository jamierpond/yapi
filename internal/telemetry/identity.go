package telemetry

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
)

// getMachineID returns a stable identifier for this machine.
// It's a hash of the user's home directory and hostname.
func getMachineID() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "unknown-home"
	}

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}

	h := sha256.New()
	h.Write([]byte(home + ":" + hostname))
	return hex.EncodeToString(h.Sum(nil))
}
