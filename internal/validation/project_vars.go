package validation

import (
	"fmt"
	"os"
	"strings"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/vars"
)

// VarDiagnosis holds information about a variable's availability across environments.
type VarDiagnosis struct {
	Name          string
	DefinedInEnvs []string
	MissingInEnvs []string
	IsInDefaults  bool
	IsOS          bool
}

// ValidateProjectVars performs matrix validation of variables across all environments.
// This is the "smart validation" that enables diagnostics like "API_URL is missing in 'staging'".
func ValidateProjectVars(text string, project *config.ProjectConfigV1, projectRoot string) []Diagnostic {
	if project == nil {
		// Fallback to legacy OS env check
		return validateEnvVars(text)
	}

	// 1. Extract all ${VAR} tokens from the config file (excluding chain refs)
	varNames := extractEnvVarNames(text)
	if len(varNames) == 0 {
		return nil
	}

	// 2. Build matrix: check each variable against each environment
	varMatrix := make(map[string]*VarDiagnosis)
	for varName := range varNames {
		varMatrix[varName] = &VarDiagnosis{
			Name:          varName,
			DefinedInEnvs: []string{},
			MissingInEnvs: []string{},
			IsInDefaults:  false,
			IsOS:          false,
		}
	}

	// 3. Check if variables are in defaults
	for varName := range varNames {
		// Check defaults.vars
		if _, ok := project.Defaults.Vars[varName]; ok {
			varMatrix[varName].IsInDefaults = true
			continue
		}

		// Check defaults.env_files
		for _, envFile := range project.Defaults.EnvFiles {
			defaultVars, err := project.ResolveEnvFiles(projectRoot, "")
			if err == nil {
				if _, ok := defaultVars[varName]; ok {
					varMatrix[varName].IsInDefaults = true
					break
				}
			}
		}

		// Check if it's an OS environment variable
		if _, ok := os.LookupEnv(varName); ok {
			varMatrix[varName].IsOS = true
		}
	}

	// 4. Check each environment
	for envName := range project.Environments {
		envVars, err := project.ResolveEnvFiles(projectRoot, envName)
		if err != nil {
			// If we can't load the env, skip it
			continue
		}

		for varName := range varNames {
			// Check if variable is defined in this environment
			_, inEnvVars := envVars[varName]
			_, inEnvConfig := project.Environments[envName].Vars[varName]

			if inEnvVars || inEnvConfig || varMatrix[varName].IsInDefaults {
				varMatrix[varName].DefinedInEnvs = append(varMatrix[varName].DefinedInEnvs, envName)
			} else {
				varMatrix[varName].MissingInEnvs = append(varMatrix[varName].MissingInEnvs, envName)
			}
		}
	}

	// 5. Generate diagnostics based on the matrix
	var diags []Diagnostic

	for varName, diagnosis := range varMatrix {
		// Case A: Variable is in defaults or OS -> No error
		if diagnosis.IsInDefaults || diagnosis.IsOS {
			continue
		}

		// Case B: Variable is defined in ALL environments -> No error
		if len(diagnosis.MissingInEnvs) == 0 {
			continue
		}

		// Case C: Variable is missing in SOME environments -> Warning
		if len(diagnosis.DefinedInEnvs) > 0 && len(diagnosis.MissingInEnvs) > 0 {
			line := findVarLine(text, varName)
			envList := strings.Join(diagnosis.MissingInEnvs, ", ")
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Field:    varName,
				Message:  fmt.Sprintf("variable '%s' is missing in environment(s): %s", varName, envList),
				Line:     line,
				Col:      0,
			})
			continue
		}

		// Case D: Variable is not found in ANY environment -> Error
		if len(diagnosis.DefinedInEnvs) == 0 {
			line := findVarLine(text, varName)
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    varName,
				Message:  fmt.Sprintf("variable '%s' is not defined in any environment or defaults", varName),
				Line:     line,
				Col:      0,
			})
		}
	}

	return diags
}

// extractEnvVarNames extracts all unique environment variable names from the text.
// Excludes chain references (${step.field}) and JQ variables ($_var).
func extractEnvVarNames(text string) map[string]bool {
	result := make(map[string]bool)
	refs := FindEnvVarRefs(text)

	for _, ref := range refs {
		// Skip if it's a chain reference
		if strings.Contains(ref.Name, ".") {
			continue
		}
		// Skip JQ variables (start with underscore)
		if strings.HasPrefix(ref.Name, "_") {
			continue
		}
		result[ref.Name] = true
	}

	return result
}

// findVarLine finds the line number where a variable is first referenced in the text.
func findVarLine(text, varName string) int {
	lines := strings.Split(text, "\n")

	// Look for ${VAR} or $VAR patterns
	patterns := []string{
		fmt.Sprintf("${%s}", varName),
		fmt.Sprintf("$%s", varName),
	}

	for lineNum, line := range lines {
		for _, pattern := range patterns {
			if strings.Contains(line, pattern) {
				// Make sure it's not a chain reference
				if !strings.Contains(line, pattern+".") {
					return lineNum
				}
			}
		}
	}

	return -1
}

// BuildProjectResolver creates a Resolver that uses project environment variables.
// The resolver checks: 1) OS env, 2) Project env vars, 3) Empty string fallback.
func BuildProjectResolver(projectVars map[string]string) vars.Resolver {
	return func(key string) (string, error) {
		// OS environment variables take precedence
		if val, ok := os.LookupEnv(key); ok {
			return val, nil
		}

		// Then check project vars
		if val, ok := projectVars[key]; ok {
			return val, nil
		}

		// Return empty string if not found (os.ExpandEnv behavior)
		return "", nil
	}
}
