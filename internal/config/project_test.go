package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadProject_WithEnvFilesInEnvironment(t *testing.T) {
	// This test ensures that env_files in an environment doesn't cause a panic
	// due to duplicate yaml keys from the embedded ConfigV1.
	tmpDir := t.TempDir()

	// Create a project config with env_files in an environment
	projectConfig := `yapi: v1
kind: project
default_environment: dev

environments:
  dev:
    env_files:
      - .env.dev
    vars:
      API_URL: http://localhost:3000
  prod:
    env_files:
      - .env.prod
    vars:
      API_URL: https://api.example.com
`
	configPath := filepath.Join(tmpDir, "yapi.config.yml")
	if err := os.WriteFile(configPath, []byte(projectConfig), 0600); err != nil {
		t.Fatal(err)
	}

	// Create dummy .env files so ResolveEnvFiles doesn't fail
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.dev"), []byte("DEV_VAR=dev"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.prod"), []byte("PROD_VAR=prod"), 0600); err != nil {
		t.Fatal(err)
	}

	// This should not panic
	project, err := LoadProject(tmpDir)
	if err != nil {
		t.Fatalf("LoadProject() failed: %v", err)
	}

	// Verify the config was loaded correctly
	if project.DefaultEnvironment != "dev" {
		t.Errorf("DefaultEnvironment = %q, want %q", project.DefaultEnvironment, "dev")
	}

	devEnv, err := project.GetEnvironment("dev")
	if err != nil {
		t.Fatalf("GetEnvironment(dev) failed: %v", err)
	}

	// env_files should be accessible via the embedded ConfigV1
	if len(devEnv.EnvFiles) != 1 || devEnv.EnvFiles[0] != ".env.dev" {
		t.Errorf("dev.EnvFiles = %v, want [.env.dev]", devEnv.EnvFiles)
	}

	// Test that ResolveEnvFiles works
	vars, err := project.ResolveEnvFiles(tmpDir, "dev")
	if err != nil {
		t.Fatalf("ResolveEnvFiles() failed: %v", err)
	}

	if vars["API_URL"] != "http://localhost:3000" {
		t.Errorf("API_URL = %q, want %q", vars["API_URL"], "http://localhost:3000")
	}

	if vars["DEV_VAR"] != "dev" {
		t.Errorf("DEV_VAR = %q, want %q", vars["DEV_VAR"], "dev")
	}
}

func TestLoadProject_WithEnvFilesInDefaults(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a project config with env_files in defaults
	projectConfig := `yapi: v1
kind: project
default_environment: dev

defaults:
  env_files:
    - .env.shared
  vars:
    SHARED_VAR: shared

environments:
  dev:
    vars:
      API_URL: http://localhost:3000
`
	configPath := filepath.Join(tmpDir, "yapi.config.yml")
	if err := os.WriteFile(configPath, []byte(projectConfig), 0600); err != nil {
		t.Fatal(err)
	}

	// Create dummy .env file
	if err := os.WriteFile(filepath.Join(tmpDir, ".env.shared"), []byte("FROM_FILE=loaded"), 0600); err != nil {
		t.Fatal(err)
	}

	project, err := LoadProject(tmpDir)
	if err != nil {
		t.Fatalf("LoadProject() failed: %v", err)
	}

	vars, err := project.ResolveEnvFiles(tmpDir, "dev")
	if err != nil {
		t.Fatalf("ResolveEnvFiles() failed: %v", err)
	}

	if vars["FROM_FILE"] != "loaded" {
		t.Errorf("FROM_FILE = %q, want %q", vars["FROM_FILE"], "loaded")
	}

	if vars["SHARED_VAR"] != "shared" {
		t.Errorf("SHARED_VAR = %q, want %q", vars["SHARED_VAR"], "shared")
	}
}
