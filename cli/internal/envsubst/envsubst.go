package envsubst

import (
	"os"
	"regexp"
)

// envVarPattern matches ${VAR_NAME} patterns
var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)\}`)

// Substitute replaces all ${VAR_NAME} patterns with their environment variable values.
// Returns the substituted string and any error if a referenced var is not set.
func Substitute(s string) string {
	return envVarPattern.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name from ${VAR_NAME}
		varName := match[2 : len(match)-1]
		if val, ok := os.LookupEnv(varName); ok {
			return val
		}
		// Return original if not set (validation catches this separately)
		return match
	})
}

// FindMissing returns a list of env var names referenced in the string that are not set.
func FindMissing(s string) []string {
	var missing []string
	seen := make(map[string]bool)

	matches := envVarPattern.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		varName := match[1]
		if seen[varName] {
			continue
		}
		seen[varName] = true

		if _, ok := os.LookupEnv(varName); !ok {
			missing = append(missing, varName)
		}
	}

	return missing
}

// FindAll returns all env var names referenced in the string.
func FindAll(s string) []string {
	var vars []string
	seen := make(map[string]bool)

	matches := envVarPattern.FindAllStringSubmatch(s, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		varName := match[1]
		if seen[varName] {
			continue
		}
		seen[varName] = true
		vars = append(vars, varName)
	}

	return vars
}

// FindAllWithPositions returns all env var references with their positions in the string.
type EnvVarRef struct {
	Name  string
	Start int
	End   int
}

func FindAllWithPositions(s string) []EnvVarRef {
	var refs []EnvVarRef

	matches := envVarPattern.FindAllStringSubmatchIndex(s, -1)
	for _, match := range matches {
		if len(match) < 4 {
			continue
		}
		// match[0:1] is the full match ${VAR}, match[2:3] is the capture group VAR
		refs = append(refs, EnvVarRef{
			Name:  s[match[2]:match[3]],
			Start: match[0],
			End:   match[1],
		})
	}

	return refs
}
