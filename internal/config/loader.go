// Package config handles parsing and loading yapi config files.
package config

import (
	"fmt"

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
func LoadFromString(data string) (*ParseResult, error) {
	return LoadFromStringWithOptions(data, nil, nil)
}

// LoadFromStringWithResolver parses a yapi config from raw YAML data using a custom variable resolver.
// If resolver is nil, uses the default EnvResolver.
func LoadFromStringWithResolver(data string, resolver vars.Resolver) (*ParseResult, error) {
	return LoadFromStringWithOptions(data, resolver, nil)
}

// LoadFromStringWithOptions parses a yapi config with optional resolver and environment defaults.
func LoadFromStringWithOptions(data string, resolver vars.Resolver, defaults *ConfigV1) (*ParseResult, error) {
	// 1. Peek at version
	var env Envelope
	if err := yaml.Unmarshal([]byte(data), &env); err != nil {
		return nil, fmt.Errorf("invalid yaml: %w", err)
	}

	// 2. Dispatch based on version
	switch env.Yapi {
	case "v1":
		return parseV1WithOptions([]byte(data), resolver, defaults)
	case "":
		// Legacy support: Parse as V1 but warn
		res, err := parseV1WithOptions([]byte(data), resolver, defaults)
		if err == nil {
			res.Warnings = append(res.Warnings, "Missing 'yapi: v1' version tag. Defaulting to v1.")
		}
		return res, err
	default:
		return nil, fmt.Errorf("unsupported yapi version: %s", env.Yapi)
	}
}

func parseV1WithOptions(data []byte, resolver vars.Resolver, defaults *ConfigV1) (*ParseResult, error) {
	var v1 ConfigV1
	if err := yaml.Unmarshal(data, &v1); err != nil {
		return nil, err
	}

	// Merge with environment defaults if provided
	if defaults != nil {
		v1 = v1.MergeWithDefaults(*defaults)
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
