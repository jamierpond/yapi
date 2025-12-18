package validation

import (
	"fmt"
	"os"
	"sort"
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
	// Load all default variables once
	defaultVars := make(map[string]string)
	if len(project.Defaults.EnvFiles) > 0 || len(project.Defaults.Vars) > 0 {
		vars, err := project.ResolveEnvFiles(projectRoot, "")
		if err == nil {
			defaultVars = vars
		}
	}

	for varName := range varNames {
		// Check defaults (from both vars and env_files)
		if _, ok := defaultVars[varName]; ok {
			varMatrix[varName].IsInDefaults = true
			continue
		}
		if _, ok := project.Defaults.Vars[varName]; ok {
			varMatrix[varName].IsInDefaults = true
			continue
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
		// Case A: Variable is in defaults -> No diagnostic needed
		if diagnosis.IsInDefaults {
			continue
		}

		// Case B: Variable is only in OS environment -> Warning (non-reproducible)
		// This makes validation deterministic and encourages explicit config
		if diagnosis.IsOS && len(diagnosis.DefinedInEnvs) == 0 {
			line := findVarLine(text, varName)
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Field:    varName,
				Message:  fmt.Sprintf("variable '%s' only defined in OS environment (not in project config). Consider adding to yapi.config.yml for reproducibility", varName),
				Line:     line,
				Col:      0,
			})
			continue
		}

		// Case C: Variable is defined in ALL environments -> No error
		if len(diagnosis.MissingInEnvs) == 0 {
			continue
		}

		// Case D: Variable is missing in SOME environments -> Warning
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

		// Case E: Variable is not found in ANY environment and not in OS -> Error
		if len(diagnosis.DefinedInEnvs) == 0 && !diagnosis.IsOS {
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
// Excludes chain references (${step.field}) and known JQ built-in variables.
func extractEnvVarNames(text string) map[string]bool {
	result := make(map[string]bool)
	refs := FindEnvVarRefs(text)

	for _, ref := range refs {
		// Skip if it's a chain reference
		if strings.Contains(ref.Name, ".") {
			continue
		}
		// Skip known JQ built-in variables (_headers, _body, etc.)
		// Note: FindEnvVarRefs already filters these out, but we check again for safety
		if isJQBuiltin(ref.Name) {
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

// EnvironmentRequirement describes whether a config requires an environment
type EnvironmentRequirement struct {
	Required         bool              // True if an environment is required
	MissingVariables []string          // Variables that are not defined anywhere
	PartialVariables map[string]string // Variables defined in some envs (var -> envs CSV)
	Message          string            // Helpful error message
}

// CheckEnvironmentRequirement analyzes a config to determine if it needs an environment.
// Returns requirement info including which variables are missing and where they're defined.
func CheckEnvironmentRequirement(text string, project *config.ProjectConfigV1, projectRoot string) *EnvironmentRequirement {
	// Extract all variables used in the config
	varNames := extractEnvVarNames(text)
	if len(varNames) == 0 {
		// No variables used - no environment needed
		return &EnvironmentRequirement{Required: false}
	}

	// Build matrix of where each variable is defined
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

	// Check defaults
	defaultVars := make(map[string]string)
	if len(project.Defaults.EnvFiles) > 0 || len(project.Defaults.Vars) > 0 {
		vars, err := project.ResolveEnvFiles(projectRoot, "")
		if err == nil {
			defaultVars = vars
		}
	}

	for varName := range varNames {
		// Check OS environment
		if _, ok := os.LookupEnv(varName); ok {
			varMatrix[varName].IsOS = true
			continue
		}

		// Check defaults
		if _, ok := defaultVars[varName]; ok {
			varMatrix[varName].IsInDefaults = true
			continue
		}

		// Check each environment
		for envName := range project.Environments {
			envVars, err := project.ResolveEnvFiles(projectRoot, envName)
			if err != nil {
				continue
			}
			if _, ok := envVars[varName]; ok {
				varMatrix[varName].DefinedInEnvs = append(varMatrix[varName].DefinedInEnvs, envName)
			} else {
				varMatrix[varName].MissingInEnvs = append(varMatrix[varName].MissingInEnvs, envName)
			}
		}
	}

	// Analyze results
	var missingVars []string
	partialVars := make(map[string]string)
	var anyEnvSpecificVar bool

	for varName, diagnosis := range varMatrix {
		// Skip if in OS env or defaults (always available)
		if diagnosis.IsOS || diagnosis.IsInDefaults {
			continue
		}

		// Variable is not defined anywhere - critical error
		if len(diagnosis.DefinedInEnvs) == 0 {
			missingVars = append(missingVars, varName)
			continue
		}

		// Variable is only in environments (not in OS or defaults)
		// This means we need to select an environment
		anyEnvSpecificVar = true

		// Variable is defined in some but not all environments
		if len(diagnosis.MissingInEnvs) > 0 {
			sort.Strings(diagnosis.DefinedInEnvs)
			partialVars[varName] = strings.Join(diagnosis.DefinedInEnvs, ", ")
		} else {
			// Variable is in all environments - still need to pick one
			sort.Strings(diagnosis.DefinedInEnvs)
			partialVars[varName] = strings.Join(diagnosis.DefinedInEnvs, ", ")
		}
	}

	// If all variables are satisfied by OS env or defaults, no environment needed
	if !anyEnvSpecificVar && len(missingVars) == 0 {
		return &EnvironmentRequirement{Required: false}
	}

	// Build helpful error message
	var msg strings.Builder
	msg.WriteString("This config requires environment variables that are not currently defined.\n")

	if len(missingVars) > 0 {
		sort.Strings(missingVars)
		msg.WriteString("\nNot defined in any environment:\n")
		for _, varName := range missingVars {
			msg.WriteString(fmt.Sprintf("  - %s\n", varName))
		}
	}

	if len(partialVars) > 0 {
		msg.WriteString("\nDefined in some environments:\n")
		// Sort for consistent output
		varList := make([]string, 0, len(partialVars))
		for varName := range partialVars {
			varList = append(varList, varName)
		}
		sort.Strings(varList)
		for _, varName := range varList {
			msg.WriteString(fmt.Sprintf("  - %s (available in: %s)\n", varName, partialVars[varName]))
		}
	}

	availableEnvs := project.ListEnvironments()
	sort.Strings(availableEnvs)
	msg.WriteString("\nUse --env <name> to select an environment.\n")
	msg.WriteString(fmt.Sprintf("Available environments: %s", strings.Join(availableEnvs, ", ")))

	return &EnvironmentRequirement{
		Required:         true,
		MissingVariables: missingVars,
		PartialVariables: partialVars,
		Message:          msg.String(),
	}
}
