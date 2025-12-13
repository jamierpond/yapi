// Package config handles parsing and loading yapi config files.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
	"yapi.run/cli/internal/domain"
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

// Load reads and parses a yapi config file from the given path.
func Load(path string) (*ParseResult, error) {
	data, err := os.ReadFile(path) //nolint:gosec // user-provided config file
	if err != nil {
		return nil, err
	}
	return LoadFromString(string(data))
}

// LoadFromString parses a yapi config from raw YAML data.
func LoadFromString(data string) (*ParseResult, error) {
	// 1. Peek at version
	var env Envelope
	if err := yaml.Unmarshal([]byte(data), &env); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	// 2. Dispatch based on version
	switch env.Yapi {
	case "v1":
		return parseV1([]byte(data))
	case "":
		// Legacy support: Parse as V1 but warn
		res, err := parseV1([]byte(data))
		if err == nil {
			res.Warnings = append(res.Warnings, "Missing 'yapi: v1' version tag. Defaulting to v1.")
		}
		return res, err
	default:
		return nil, fmt.Errorf("unsupported yapi version: %s", env.Yapi)
	}
}

func parseV1(data []byte) (*ParseResult, error) {
	var v1 ConfigV1
	if err := yaml.Unmarshal(data, &v1); err != nil {
		return nil, err
	}

	// Check if this is a chain config
	if len(v1.Chain) > 0 {
		return &ParseResult{Chain: v1.Chain, Base: &v1}, nil
	}

	domainReq, err := v1.ToDomain()
	if err != nil {
		return nil, err
	}

	return &ParseResult{Request: domainReq, Expect: v1.Expect, Base: &v1}, nil
}
