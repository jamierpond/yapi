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

// EnvFileStatus represents the validation state of an env file reference
type EnvFileStatus struct {
	Path     string // Original path from config
	Resolved string // Absolute path after resolution
	Exists   bool   // Whether file exists
	Readable bool   // Whether file is readable (if exists)
	Error    error  // Error if not readable (permission denied, etc.)
	Line     int    // Line number in source YAML
	Col      int    // Column number in source YAML
}

// EnvFileLoadResult contains the result of loading env files
type EnvFileLoadResult struct {
	Variables  map[string]string // Merged variables from all valid files
	Warnings   []string          // Warnings for missing files
	FileStatus []EnvFileStatus   // Status of each env file
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
	return loadFromStringInternal(data, "", nil, nil, ResolverOptions{})
}

// LoadFromStringWithPath parses a yapi config with path context for resolving relative env_files.
func LoadFromStringWithPath(data string, configPath string, resolver vars.Resolver, defaults *ConfigV1) (*ParseResult, error) {
	return loadFromStringInternal(data, configPath, resolver, defaults, ResolverOptions{})
}

// LoadOptions contains options for loading a config.
type LoadOptions struct {
	ConfigPath string        // Path to config file for relative env_files resolution
	Resolver   vars.Resolver // Custom variable resolver
	Defaults   *ConfigV1     // Default config values
	StrictEnv  bool          // Strict mode: error on missing env files, no OS fallback
}

// LoadFromStringWithOptions parses a yapi config with full options support.
func LoadFromStringWithOptions(data string, opts LoadOptions) (*ParseResult, error) {
	return loadFromStringInternal(data, opts.ConfigPath, opts.Resolver, opts.Defaults, ResolverOptions{StrictEnv: opts.StrictEnv})
}

// loadFromStringInternal is the shared implementation for loading configs.
func loadFromStringInternal(data string, configPath string, resolver vars.Resolver, defaults *ConfigV1, opts ResolverOptions) (*ParseResult, error) {
	// 1. Peek at version
	var env Envelope
	if err := yaml.Unmarshal([]byte(data), &env); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	// 2. Dispatch based on version
	switch env.Yapi {
	case "v1":
		return parseV1WithOptions([]byte(data), configPath, resolver, defaults, opts)
	case "":
		// Legacy support: Parse as V1 but warn
		res, err := parseV1WithOptions([]byte(data), configPath, resolver, defaults, opts)
		if err == nil {
			res.Warnings = append(res.Warnings, "Missing 'yapi: v1' version tag. Defaulting to v1.")
		}
		return res, err
	default:
		return nil, fmt.Errorf("unsupported yapi version: %s", env.Yapi)
	}
}

func parseV1WithOptions(data []byte, configPath string, resolver vars.Resolver, defaults *ConfigV1, opts ResolverOptions) (*ParseResult, error) {
	var v1 ConfigV1
	if err := yaml.Unmarshal(data, &v1); err != nil {
		return nil, err
	}

	// Merge with environment defaults if provided
	if defaults != nil {
		v1 = v1.MergeWithDefaults(*defaults)
	}

	// Collect warnings from env file loading
	var envFileWarnings []string

	// Load env files if specified in the config
	if len(v1.EnvFiles) > 0 {
		envFileVars, warnings, err := loadEnvFilesWithOptions(v1.EnvFiles, configPath, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to load env_files: %w", err)
		}
		envFileWarnings = warnings

		// Create a combined resolver: env files vars > existing resolver > OS env (fallback)
		resolver = buildEnvFileResolverWithOptions(envFileVars, resolver, opts)
	}

	// Check if this is a chain config
	if len(v1.Chain) > 0 {
		return &ParseResult{Chain: v1.Chain, Base: &v1, Warnings: envFileWarnings}, nil
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

	return &ParseResult{Request: domainReq, Expect: v1.Expect, Base: &baseCopy, Warnings: envFileWarnings}, nil
}

// loadEnvFiles loads variables from the specified .env files.
// Paths are resolved relative to the config file directory.
// Returns variables, warnings for missing files, and error for permission/parse issues.
//
//nolint:unused
func loadEnvFiles(envFiles []string, configPath string) (map[string]string, []string, error) {
	return loadEnvFilesWithOptions(envFiles, configPath, ResolverOptions{})
}

// loadEnvFilesWithOptions loads env files with configurable options.
// In strict mode, missing files are errors instead of warnings.
// configPath should be the path to a config file (directory is extracted via filepath.Dir).
func loadEnvFilesWithOptions(envFiles []string, configPath string, opts ResolverOptions) (map[string]string, []string, error) {
	// Determine base directory for resolving relative paths
	baseDir := "."
	if configPath != "" {
		baseDir = filepath.Dir(configPath)
	}
	return loadEnvFilesFromDir(envFiles, baseDir, opts)
}

// loadEnvFilesFromDir loads env files with paths resolved relative to baseDir.
// This is the internal implementation used by both loadEnvFilesWithOptions (for config files)
// and ResolveEnvFilesWithWarnings (for project directories).
func loadEnvFilesFromDir(envFiles []string, baseDir string, opts ResolverOptions) (map[string]string, []string, error) {
	result := make(map[string]string)
	var warnings []string
	seen := make(map[string]bool) // Track seen files to avoid duplicate warnings

	for _, envFile := range envFiles {
		// Resolve relative paths against the config file directory
		filePath := envFile
		if !filepath.IsAbs(envFile) {
			filePath = filepath.Join(baseDir, envFile)
		}

		// Skip duplicate entries
		if seen[filePath] {
			continue
		}
		seen[filePath] = true

		// Check if file exists
		fileInfo, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			if opts.StrictEnv {
				return nil, nil, fmt.Errorf("env file %q not found", envFile)
			}
			warnings = append(warnings, fmt.Sprintf("env file '%s' not found", envFile))
			continue
		}

		// Check if it's a directory (not a file)
		if fileInfo != nil && fileInfo.IsDir() {
			return nil, nil, fmt.Errorf("env file %q is a directory, not a file", envFile)
		}

		// Check readability by trying to open it
		f, err := os.Open(filePath) // #nosec G304 -- filePath is constructed from configPath and envFile
		if err != nil {
			// Permission error or other read error - always an error, not a warning
			return nil, nil, fmt.Errorf("cannot read env file '%s': %w", envFile, err)
		}
		_ = f.Close()

		// Parse the env file
		envVars, err := godotenv.Read(filePath)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse env file %q: %w", envFile, err)
		}

		// Merge into result (later files override earlier ones)
		for k, v := range envVars {
			result[k] = v
		}
	}

	return result, warnings, nil
}

// ResolutionSource indicates where a variable was resolved from
type ResolutionSource int

const (
	// SourceNotFound indicates the variable was not found in any source
	SourceNotFound ResolutionSource = iota
	// SourceEnvFile indicates the variable was found in an env file
	SourceEnvFile
	// SourceProjectConfig indicates the variable was found in project config vars
	SourceProjectConfig
	// SourceOSEnv indicates the variable was found in OS environment
	SourceOSEnv
)

// ResolverOptions configures the behavior of the env file resolver
type ResolverOptions struct {
	StrictEnv bool // If true, don't fall back to OS env and error on missing files
}

// ResolverWithTracking is a resolver that also tracks where variables were resolved from
type ResolverWithTracking struct {
	Resolver         vars.Resolver
	ResolutionSource map[string]ResolutionSource // Tracks where each var was resolved from
	OSFallbackVars   []string                    // Variables that fell back to OS env
}

// buildEnvFileResolver creates a resolver that combines env file vars with an existing resolver.
// Priority order: env file vars > existing resolver > OS env (fallback only)
//
//nolint:unused
func buildEnvFileResolver(envFileVars map[string]string, existingResolver vars.Resolver) vars.Resolver {
	return buildEnvFileResolverWithOptions(envFileVars, existingResolver, ResolverOptions{})
}

// buildEnvFileResolverWithOptions creates a resolver with configurable options.
// Priority order: env file vars > existing resolver > OS env (fallback disabled in strict mode)
func buildEnvFileResolverWithOptions(envFileVars map[string]string, existingResolver vars.Resolver, opts ResolverOptions) vars.Resolver {
	return func(key string) (string, error) {
		// 1. Check env file vars first (highest priority from config)
		if val, ok := envFileVars[key]; ok {
			return val, nil
		}

		// 2. Check existing resolver (project config vars, etc.)
		if existingResolver != nil {
			if val, err := existingResolver(key); err == nil && val != "" {
				return val, nil
			}
		}

		// 3. Check OS environment (fallback only, unless strict mode)
		if !opts.StrictEnv {
			if val, ok := os.LookupEnv(key); ok {
				return val, nil
			}
		}

		// 4. Return empty string (os.ExpandEnv behavior)
		return "", nil
	}
}

// buildEnvFileResolverWithTracking creates a resolver that tracks resolution sources.
// Returns a ResolverWithTracking that includes the resolver and tracking information.
// Used by LSP for showing resolution source info diagnostics.
//
//nolint:unused
func buildEnvFileResolverWithTracking(envFileVars map[string]string, existingResolver vars.Resolver, opts ResolverOptions) *ResolverWithTracking {
	tracking := &ResolverWithTracking{
		ResolutionSource: make(map[string]ResolutionSource),
		OSFallbackVars:   []string{},
	}

	tracking.Resolver = func(key string) (string, error) {
		// 1. Check env file vars first (highest priority from config)
		if val, ok := envFileVars[key]; ok {
			tracking.ResolutionSource[key] = SourceEnvFile
			return val, nil
		}

		// 2. Check existing resolver (project config vars, etc.)
		if existingResolver != nil {
			if val, err := existingResolver(key); err == nil && val != "" {
				tracking.ResolutionSource[key] = SourceProjectConfig
				return val, nil
			}
		}

		// 3. Check OS environment (fallback only, unless strict mode)
		if !opts.StrictEnv {
			if val, ok := os.LookupEnv(key); ok {
				tracking.ResolutionSource[key] = SourceOSEnv
				tracking.OSFallbackVars = append(tracking.OSFallbackVars, key)
				return val, nil
			}
		}

		// 4. Not found
		tracking.ResolutionSource[key] = SourceNotFound
		return "", nil
	}

	return tracking
}
