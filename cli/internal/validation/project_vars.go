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

// buildVarMatrix builds a matrix of variable availability across all environments.
// This is the single source of truth for variable resolution logic.
// requestEnvFileVars is an optional set of variable names from the request's own env_files.
func buildVarMatrix(text string, project *config.ProjectConfigV1, projectRoot string, requestEnvFileVars map[string]bool) map[string]*VarDiagnosis {
	varNames := extractEnvVarNames(text)
	if len(varNames) == 0 {
		return nil
	}

	varMatrix := make(map[string]*VarDiagnosis)
	for varName := range varNames {
		varMatrix[varName] = &VarDiagnosis{
			Name:          varName,
			DefinedInEnvs: []string{},
			MissingInEnvs: []string{},
		}
	}

	// Load project defaults
	defaultVars := make(map[string]string)
	if project != nil && (len(project.Defaults.EnvFiles) > 0 || len(project.Defaults.Vars) > 0) {
		vars, err := project.ResolveEnvFiles(projectRoot, "")
		if err == nil {
			defaultVars = vars
		}
	}

	for varName := range varNames {
		// Variables from request's own env_files are treated as defined
		if requestEnvFileVars != nil && requestEnvFileVars[varName] {
			varMatrix[varName].IsInDefaults = true
			continue
		}

		// Check project defaults
		if _, ok := defaultVars[varName]; ok {
			varMatrix[varName].IsInDefaults = true
			continue
		}
		if project != nil {
			if _, ok := project.Defaults.Vars[varName]; ok {
				varMatrix[varName].IsInDefaults = true
				continue
			}
		}

		// Check OS environment
		if _, ok := os.LookupEnv(varName); ok {
			varMatrix[varName].IsOS = true
		}
	}

	// Check each project environment
	if project != nil {
		for envName := range project.Environments {
			envVars, err := project.ResolveEnvFiles(projectRoot, envName)
			if err != nil {
				continue
			}

			for varName := range varNames {
				_, inEnvVars := envVars[varName]
				_, inEnvConfig := project.Environments[envName].Vars[varName]

				if inEnvVars || inEnvConfig || varMatrix[varName].IsInDefaults {
					varMatrix[varName].DefinedInEnvs = append(varMatrix[varName].DefinedInEnvs, envName)
				} else {
					varMatrix[varName].MissingInEnvs = append(varMatrix[varName].MissingInEnvs, envName)
				}
			}
		}
	}

	return varMatrix
}

// ValidateProjectVars performs matrix validation of variables across all environments.
// requestEnvFileVars is an optional set of variable names from the request's own env_files.
func ValidateProjectVars(text string, project *config.ProjectConfigV1, projectRoot string, requestEnvFileVars map[string]bool) []Diagnostic {
	if project == nil {
		return validateEnvVars(text)
	}

	varMatrix := buildVarMatrix(text, project, projectRoot, requestEnvFileVars)
	if varMatrix == nil {
		return nil
	}

	var diags []Diagnostic
	for varName, diagnosis := range varMatrix {
		if diagnosis.IsInDefaults {
			continue
		}

		// OS-only variable: warn about non-reproducibility
		if diagnosis.IsOS && len(diagnosis.DefinedInEnvs) == 0 {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Field:    varName,
				Message:  fmt.Sprintf("variable '%s' only defined in OS environment (not in project config). Consider adding to yapi.config.yml for reproducibility", varName),
				Line:     findVarLine(text, varName),
			})
			continue
		}

		// Defined in all environments: no error
		if len(diagnosis.MissingInEnvs) == 0 {
			continue
		}

		// Missing in some environments: warn
		if len(diagnosis.DefinedInEnvs) > 0 && len(diagnosis.MissingInEnvs) > 0 {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Field:    varName,
				Message:  fmt.Sprintf("variable '%s' is missing in environment(s): %s", varName, strings.Join(diagnosis.MissingInEnvs, ", ")),
				Line:     findVarLine(text, varName),
			})
			continue
		}

		// Not found anywhere: error
		if len(diagnosis.DefinedInEnvs) == 0 && !diagnosis.IsOS {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    varName,
				Message:  fmt.Sprintf("variable '%s' is not defined in any environment or defaults", varName),
				Line:     findVarLine(text, varName),
			})
		}
	}

	return diags
}

// EnvironmentRequirement describes whether a config requires an environment
type EnvironmentRequirement struct {
	Required         bool
	MissingVariables []string
	PartialVariables map[string]string
	Message          string
}

// CheckEnvironmentRequirement analyzes a config to determine if it needs an environment.
// requestEnvFileVars is an optional set of variable names from the request's own env_files.
func CheckEnvironmentRequirement(text string, project *config.ProjectConfigV1, projectRoot string, requestEnvFileVars map[string]bool) *EnvironmentRequirement {
	varMatrix := buildVarMatrix(text, project, projectRoot, requestEnvFileVars)
	if varMatrix == nil {
		return &EnvironmentRequirement{Required: false}
	}

	var missingVars []string
	partialVars := make(map[string]string)
	var anyEnvSpecificVar bool

	for varName, diagnosis := range varMatrix {
		if diagnosis.IsOS || diagnosis.IsInDefaults {
			continue
		}

		if len(diagnosis.DefinedInEnvs) == 0 {
			missingVars = append(missingVars, varName)
			continue
		}

		anyEnvSpecificVar = true
		sort.Strings(diagnosis.DefinedInEnvs)
		partialVars[varName] = strings.Join(diagnosis.DefinedInEnvs, ", ")
	}

	if !anyEnvSpecificVar && len(missingVars) == 0 {
		return &EnvironmentRequirement{Required: false}
	}

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
		varList := make([]string, 0, len(partialVars))
		for varName := range partialVars {
			varList = append(varList, varName)
		}
		sort.Strings(varList)
		for _, varName := range varList {
			msg.WriteString(fmt.Sprintf("  - %s (available in: %s)\n", varName, partialVars[varName]))
		}
	}

	if project != nil {
		availableEnvs := project.ListEnvironments()
		sort.Strings(availableEnvs)
		msg.WriteString("\nUse --env <name> to select an environment.\n")
		msg.WriteString(fmt.Sprintf("Available environments: %s", strings.Join(availableEnvs, ", ")))
	}

	return &EnvironmentRequirement{
		Required:         true,
		MissingVariables: missingVars,
		PartialVariables: partialVars,
		Message:          msg.String(),
	}
}

// extractEnvVarNames extracts all unique environment variable names from the text.
func extractEnvVarNames(text string) map[string]bool {
	result := make(map[string]bool)
	refs := FindEnvVarRefs(text)

	for _, ref := range refs {
		if strings.Contains(ref.Name, ".") {
			continue
		}
		if isJQBuiltin(ref.Name) {
			continue
		}
		result[ref.Name] = true
	}

	return result
}

// findVarLine finds the line number where a variable is first referenced.
func findVarLine(text, varName string) int {
	lines := strings.Split(text, "\n")
	patterns := []string{
		fmt.Sprintf("${%s}", varName),
		fmt.Sprintf("$%s", varName),
	}

	for lineNum, line := range lines {
		for _, pattern := range patterns {
			if strings.Contains(line, pattern) && !strings.Contains(line, pattern+".") {
				return lineNum
			}
		}
	}

	return -1
}

// BuildProjectResolver creates a Resolver that uses project environment variables.
func BuildProjectResolver(projectVars map[string]string) vars.Resolver {
	return func(key string) (string, error) {
		if val, ok := os.LookupEnv(key); ok {
			return val, nil
		}
		if val, ok := projectVars[key]; ok {
			return val, nil
		}
		return "", nil
	}
}
