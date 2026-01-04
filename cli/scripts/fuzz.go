package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fuzzTime := os.Getenv("FUZZTIME")
	if fuzzTime == "" {
		fuzzTime = "5s"
	}

	// Get all packages
	cmd := exec.Command("go", "list", "./...")
	out, err := cmd.Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list packages: %v\n", err)
		os.Exit(1)
	}

	packages := strings.Split(strings.TrimSpace(string(out)), "\n")

	for _, pkg := range packages {
		// Check if package has fuzz tests
		cmd := exec.Command("go", "test", "-list", "^Fuzz", pkg)
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "Fuzz") {
				fmt.Printf("Fuzzing %s in %s\n", line, pkg)
				fuzzCmd := exec.Command("go", "test", "-fuzz="+line, "-fuzztime="+fuzzTime, pkg)
				fuzzCmd.Stdout = os.Stdout
				fuzzCmd.Stderr = os.Stderr
				if err := fuzzCmd.Run(); err != nil {
					var exitErr *exec.ExitError
					if errors.As(err, &exitErr) {
						os.Exit(exitErr.ExitCode())
					}
					fmt.Fprintf(os.Stderr, "Fuzz failed: %v\n", err)
					os.Exit(1)
				}
			}
		}
	}
}
