// Package vars provides shared regex patterns and utilities for variable expansion.
package vars

import (
	"regexp"
	"strings"
)

// Expansion matches $VAR and ${VAR} patterns, including dots for chain references.
// The strict form ${VAR} is always recognized.
// The lazy form $VAR requires the variable to:
//  1. Start with a letter or underscore (not a digit)
//  2. Not be preceded by alphanumeric or underscore (checked in ExpandString)
//
// This prevents matching dollar signs in bcrypt hashes ($2a$12$...) or other literals.
// Group 1: contents inside ${...}
// Group 2: token after $... (must start with letter or underscore)
var Expansion = regexp.MustCompile(`\$\{([^}]+)\}|\$([a-zA-Z_][a-zA-Z0-9_\-\.]*)`)

// EnvOnly matches $VAR and ${VAR} patterns without dots (environment variables only).
// Group 1: contents inside ${...}
// Group 2: token after $...
var EnvOnly = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)

// Resolver resolves a variable key to its value.
type Resolver func(key string) (string, error)

// ChainVar matches ${step.field} patterns (contains a dot).
// Variables must start with a letter or underscore.
var ChainVar = regexp.MustCompile(`\$\{[^}]*\.[^}]+\}|\$[a-zA-Z_][a-zA-Z0-9_\-]*\.[a-zA-Z0-9_\-\.]+`)

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

			// For lazy form, check if preceded by alphanumeric/underscore
			// This prevents matching "$k..." in bcrypt hashes like "$2a$12$k..."
			matchIndex := strings.Index(input, match)
			if matchIndex > 0 {
				prevChar := input[matchIndex-1]
				if isAlphanumericOrUnderscore(prevChar) {
					// Skip this match - it's part of a larger token (e.g., bcrypt hash)
					return match
				}
			}
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

// isAlphanumericOrUnderscore checks if a byte is alphanumeric or underscore
func isAlphanumericOrUnderscore(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}
