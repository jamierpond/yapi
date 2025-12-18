// Package config handles project-level configuration for yapi.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// ProjectConfigV1 represents a yapi.config.yml file which defines environments for a project.
type ProjectConfigV1 struct {
	Yapi               string                 `yaml:"yapi"`                  // Must be "v1"
	Kind               string                 `yaml:"kind"`                  // Must be "project"
	DefaultEnvironment string                 `yaml:"default_environment"`   // Default environment to use when --env not specified
	Defaults           EnvironmentConfig      `yaml:"defaults"`              // Default vars applied to all environments
	Environments       map[string]Environment `yaml:"environments"`          // Named environments (dev, staging, prod, etc.)
}

// EnvironmentConfig holds variable definitions that can be shared or inherited.
type EnvironmentConfig struct {
	Vars     map[string]string `yaml:"vars"`      // Direct variable definitions
	EnvFiles []string          `yaml:"env_files"` // Paths to .env files (relative to project root)
}

// Environment represents a single environment configuration.
type Environment struct {
	Name     string            // Derived from map key
	Vars     map[string]string `yaml:"vars"`      // Environment-specific variables
	EnvFiles []string          `yaml:"env_files"` // Environment-specific .env files
}

// FindProjectRoot walks up the directory tree from startDir looking for yapi.config.yml.
// Returns the directory containing the config file, or an error if not found.
func FindProjectRoot(startDir string) (string, error) {
	// Resolve to absolute path
	absPath, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	// Walk up the directory tree
	currentDir := absPath
	for {
		// Check if yapi.config.yml exists in current directory
		configPath := filepath.Join(currentDir, "yapi.config.yml")
		if _, err := os.Stat(configPath); err == nil {
			return currentDir, nil
		}

		// Try .yaml extension as well
		configPath = filepath.Join(currentDir, "yapi.config.yaml")
		if _, err := os.Stat(configPath); err == nil {
			return currentDir, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)

		// If we've reached the root, stop
		if parentDir == currentDir {
			return "", fmt.Errorf("no yapi.config.yml found in directory tree")
		}

		currentDir = parentDir
	}
}

// LoadProject loads and parses a yapi.config.yml file from the given directory.
// The projectRoot should be the directory containing the yapi.config.yml file.
func LoadProject(projectRoot string) (*ProjectConfigV1, error) {
	// Try .yml first, then .yaml
	configPath := filepath.Join(projectRoot, "yapi.config.yml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		configPath = filepath.Join(projectRoot, "yapi.config.yaml")
		data, err = os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read project config: %w", err)
		}
	}

	// Peek at version to allow for future versioning
	var env Envelope
	if err := yaml.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	if env.Yapi != "v1" {
		return nil, fmt.Errorf("unsupported project config version: %s", env.Yapi)
	}

	var config ProjectConfigV1
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse project config: %w", err)
	}

	// Validate required fields
	if config.Kind != "project" {
		return nil, fmt.Errorf("invalid or missing 'kind' (expected 'project', got '%s')", config.Kind)
	}

	// Populate environment names from map keys
	for name, env := range config.Environments {
		env.Name = name
		config.Environments[name] = env
	}

	return &config, nil
}

// GetEnvironment retrieves a specific environment by name, returning an error if not found.
func (pc *ProjectConfigV1) GetEnvironment(name string) (*Environment, error) {
	env, ok := pc.Environments[name]
	if !ok {
		return nil, fmt.Errorf("environment '%s' not defined in project config", name)
	}
	return &env, nil
}

// ListEnvironments returns a list of all environment names defined in the project.
func (pc *ProjectConfigV1) ListEnvironments() []string {
	names := make([]string, 0, len(pc.Environments))
	for name := range pc.Environments {
		names = append(names, name)
	}
	return names
}

// ResolveEnvFiles resolves all .env file paths relative to the project root and loads them.
// Returns a merged map of all variables (defaults first, then environment-specific).
// OS environment variables take precedence over all loaded vars.
func (pc *ProjectConfigV1) ResolveEnvFiles(projectRoot string, envName string) (map[string]string, error) {
	result := make(map[string]string)

	// 1. Load default .env files
	for _, envFile := range pc.Defaults.EnvFiles {
		filePath := filepath.Join(projectRoot, envFile)
		vars, err := godotenv.Read(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load default env file '%s': %w", envFile, err)
		}
		// Merge into result
		for k, v := range vars {
			result[k] = v
		}
	}

	// 2. Add default vars (override env files)
	for k, v := range pc.Defaults.Vars {
		result[k] = v
	}

	// 3. Load environment-specific .env files
	if env, ok := pc.Environments[envName]; ok {
		for _, envFile := range env.EnvFiles {
			filePath := filepath.Join(projectRoot, envFile)
			vars, err := godotenv.Read(filePath)
			if err != nil {
				return nil, fmt.Errorf("failed to load env file '%s' for environment '%s': %w", envFile, envName, err)
			}
			// Merge into result
			for k, v := range vars {
				result[k] = v
			}
		}

		// 4. Add environment-specific vars (highest priority from config)
		for k, v := range env.Vars {
			result[k] = v
		}
	}

	// 5. OS environment variables override everything
	// This happens at resolution time, not here

	return result, nil
}
