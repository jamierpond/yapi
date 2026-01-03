// Package config handles parsing and loading yapi config files.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/vars"
)

// Envelope is used solely to peek at the version
type Envelope struct {
	Yapi string `yaml:"yapi"`
}

// ParseResult holds the output of parsing a yapi config file.
type ParseResult struct {
	Request  *domain.Request
	Warnings []string
	Chain    []ChainStep // Chain steps if this is a chain config
	Base     *ConfigV1   // Base config for chain merging
	Expect   Expectation // Expectations for single request validation
}

// LoadFromString parses a yapi config from raw YAML data.
// Used by tests across multiple packages.
//
//nolint:unused
func LoadFromString(data string) (*ParseResult, error) {
	return loadFromStringInternal(data, "", nil, nil)
}

// LoadFromStringWithPath parses a yapi config with path context for resolving relative env_files.
func LoadFromStringWithPath(data string, configPath string, resolver vars.Resolver, defaults *ConfigV1) (*ParseResult, error) {
	return loadFromStringInternal(data, configPath, resolver, defaults)
}

// loadFromStringInternal is the shared implementation for loading configs.
func loadFromStringInternal(data string, configPath string, resolver vars.Resolver, defaults *ConfigV1) (*ParseResult, error) {
	// 1. Peek at version
	var env Envelope
	if err := yaml.Unmarshal([]byte(data), &env); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	// 2. Dispatch based on version
	switch env.Yapi {
	case "v1":
		return parseV1WithOptions([]byte(data), configPath, resolver, defaults)
	case "":
		// Legacy support: Parse as V1 but warn
		res, err := parseV1WithOptions([]byte(data), configPath, resolver, defaults)
		if err == nil {
			res.Warnings = append(res.Warnings, "Missing 'yapi: v1' version tag. Defaulting to v1.")
		}
		return res, err
	default:
		return nil, fmt.Errorf("unsupported yapi version: %s", env.Yapi)
	}
}

func parseV1WithOptions(data []byte, configPath string, resolver vars.Resolver, defaults *ConfigV1) (*ParseResult, error) {
	var v1 ConfigV1
	if err := yaml.Unmarshal(data, &v1); err != nil {
		return nil, err
	}

	// Merge with environment defaults if provided
	if defaults != nil {
		v1 = v1.MergeWithDefaults(*defaults)
	}

	// Load env files if specified in the config
	if len(v1.EnvFiles) > 0 {
		envFileVars, err := loadEnvFiles(v1.EnvFiles, configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load env_files: %w", err)
		}

		// Create a combined resolver: env files vars < existing resolver < OS env
		resolver = buildEnvFileResolver(envFileVars, resolver)
	}

	// Check if this is a chain config
	if len(v1.Chain) > 0 {
		return &ParseResult{Chain: v1.Chain, Base: &v1}, nil
	}

	// Keep a copy of the original config before variable expansion
	// This allows re-expansion with different resolvers later
	baseCopy := v1

	var domainReq *domain.Request
	var err error

	// Use custom resolver if provided, otherwise use default ToDomain
	if resolver != nil {
		domainReq, err = v1.ToDomainWithResolver(resolver)
	} else {
		domainReq, err = v1.ToDomain()
	}

	if err != nil {
		return nil, err
	}

	return &ParseResult{Request: domainReq, Expect: v1.Expect, Base: &baseCopy}, nil
}

// loadEnvFiles loads variables from the specified .env files.
// Paths are resolved relative to the config file directory.
func loadEnvFiles(envFiles []string, configPath string) (map[string]string, error) {
	result := make(map[string]string)

	// Determine base directory for resolving relative paths
	baseDir := "."
	if configPath != "" {
		baseDir = filepath.Dir(configPath)
	}

	for _, envFile := range envFiles {
		// Resolve relative paths against the config file directory
		filePath := envFile
		if !filepath.IsAbs(envFile) {
			filePath = filepath.Join(baseDir, envFile)
		}

		vars, err := godotenv.Read(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to load env file %q (resolved to %q): %w", envFile, filePath, err)
		}

		// Merge into result (later files override earlier ones)
		for k, v := range vars {
			result[k] = v
		}
	}

	return result, nil
}

// buildEnvFileResolver creates a resolver that combines env file vars with an existing resolver.
// Priority order: OS env > existing resolver > env file vars
func buildEnvFileResolver(envFileVars map[string]string, existingResolver vars.Resolver) vars.Resolver {
	return func(key string) (string, error) {
		// 1. Check OS environment first (highest priority)
		if val, ok := os.LookupEnv(key); ok {
			return val, nil
		}

		// 2. Check existing resolver (project config vars, etc.)
		if existingResolver != nil {
			if val, err := existingResolver(key); err == nil && val != "" {
				return val, nil
			}
		}

		// 3. Check env file vars
		if val, ok := envFileVars[key]; ok {
			return val, nil
		}

		// 4. Return empty string (os.ExpandEnv behavior)
		return "", nil
	}
}
