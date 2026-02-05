// Package vars provides shared regex patterns and utilities for variable expansion.
package vars

import (
	"regexp"
)

// Expansion matches ${VAR} patterns only (strict form required).
// This prevents ambiguity with dollar signs in bcrypt hashes, dollar amounts, etc.
// Variables can contain dots for chain references: ${step.field}
// Group 1: contents inside ${...}
var Expansion = regexp.MustCompile(`\$\{([^}]+)\}`)

// EnvOnly matches ${VAR} patterns without dots (environment variables only).
// Group 1: contents inside ${...}
var EnvOnly = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// Resolver resolves a variable key to its value.
type Resolver func(key string) (string, error)

// ChainVar matches ${step.field} patterns (contains a dot).
var ChainVar = regexp.MustCompile(`\$\{[^}]*\.[^}]+\}`)

// HasChainVars returns true if the string contains chain variable references (${step.field}).
func HasChainVars(s string) bool {
	return ChainVar.MatchString(s)
}

// HasEnvVars returns true if the string contains environment variable references ($VAR or ${VAR}).
func HasEnvVars(s string) bool {
	return EnvOnly.MatchString(s)
}

// BareChainRef matches bare $word.word patterns (without braces) that look like
// chain variable references. Negative lookbehind for { ensures we don't match ${...}.
// We use a two-step approach: find all $word.word patterns, then exclude those inside ${}.
var BareChainRef = regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*\.[A-Za-z0-9_.]+)`)

// FindBareRefs returns bare chain-style references ($step.field) that are NOT
// wrapped in ${...}. These are likely user mistakes since only ${...} is substituted.
func FindBareRefs(s string) []string {
	var results []string
	for _, loc := range BareChainRef.FindAllStringIndex(s, -1) {
		start := loc[0]
		// Check that this is not inside ${...} by looking for a preceding {
		if start > 0 && s[start-1] == '{' {
			continue
		}
		results = append(results, s[start:loc[1]])
	}
	return results
}

// ExpandString replaces all ${VAR} occurrences in input using the resolver.
func ExpandString(input string, resolver Resolver) (string, error) {
	var capturedErr error

	result := Expansion.ReplaceAllStringFunc(input, func(match string) string {
		if capturedErr != nil {
			return match
		}

		// Extract key from ${key}
		key := match[2 : len(match)-1]

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
