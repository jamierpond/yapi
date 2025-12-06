// Package vars provides shared regex patterns and utilities for variable expansion.
package vars

import (
	"regexp"
	"strings"
)

// Expansion matches $VAR and ${VAR} patterns, including dots for chain references.
// Group 1: contents inside ${...}
// Group 2: token after $...
var Expansion = regexp.MustCompile(`\$\{([^}]+)\}|\$([a-zA-Z0-9_\-\.]+)`)

// EnvOnly matches $VAR and ${VAR} patterns without dots (environment variables only).
// Group 1: contents inside ${...}
// Group 2: token after $...
var EnvOnly = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

// Resolver resolves a variable key to its value.
type Resolver func(key string) (string, error)

// ChainVar matches ${step.field} patterns (contains a dot).
var ChainVar = regexp.MustCompile(`\$\{[^}]*\.[^}]+\}|\$[a-zA-Z0-9_\-]+\.[a-zA-Z0-9_\-\.]+`)

// HasChainVars returns true if the string contains chain variable references (${step.field}).
func HasChainVars(s string) bool {
	return ChainVar.MatchString(s)
}

// HasEnvVars returns true if the string contains environment variable references ($VAR or ${VAR}).
func HasEnvVars(s string) bool {
	return EnvOnly.MatchString(s)
}

// ExpandString replaces all $VAR and ${VAR} occurrences in input using the resolver.
func ExpandString(input string, resolver Resolver) (string, error) {
	var capturedErr error

	result := Expansion.ReplaceAllStringFunc(input, func(match string) string {
		if capturedErr != nil {
			return match
		}

		var key string
		if strings.HasPrefix(match, "${") {
			// Strict: ${key}
			key = match[2 : len(match)-1]
		} else {
			// Lazy: $key
			key = match[1:]
		}

		val, err := resolver(key)
		if err != nil {
			capturedErr = err
			return match
		}
		return val
	})

	if capturedErr != nil {
		return "", capturedErr
	}
	return result, nil
}
