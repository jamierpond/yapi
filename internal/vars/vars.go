// Package vars provides shared regex patterns for variable expansion.
package vars

import "regexp"

// Expansion matches $VAR and ${VAR} patterns, including dots for chain references.
// Group 1: contents inside ${...}
// Group 2: token after $...
var Expansion = regexp.MustCompile(`\$\{([^}]+)\}|\$([a-zA-Z0-9_\-\.]+)`)

// EnvOnly matches $VAR and ${VAR} patterns without dots (environment variables only).
// Group 1: contents inside ${...}
// Group 2: token after $...
var EnvOnly = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}|\$([A-Za-z_][A-Za-z0-9_]*)`)
