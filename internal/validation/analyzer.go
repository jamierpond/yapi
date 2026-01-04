// Package validation provides config analysis and diagnostics.
package validation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/utils"
	"yapi.run/cli/internal/vars"
)

// extractLineFromError attempts to extract a line number from YAML error messages.
// YAML errors often look like "line 22: cannot unmarshal..." - returns 0-indexed line or -1 if not found.
func extractLineFromError(errMsg string) int {
	re := regexp.MustCompile(`line (\d+):`)
	matches := re.FindStringSubmatch(errMsg)
	if len(matches) >= 2 {
		if lineNum, err := strconv.Atoi(matches[1]); err == nil {
			return lineNum - 1 // Convert to 0-indexed
		}
	}
	return -1
}

// Diagnostic is the canonical diagnostic type that both CLI and LSP use.
type Diagnostic struct {
	Severity Severity
	Field    string // "url", "method", "graphql", "jq_filter", etc
	Message  string // human readable message

	// Optional position info. LSP uses it, CLI may ignore.
	Line int // 0-based, -1 if unknown
	Col  int // 0-based, -1 if unknown
}

// Analysis is the shared result type from analyzing a config.
type Analysis struct {
	Request     *domain.Request
	Diagnostics []Diagnostic
	Warnings    []string           // parsed-level warnings like missing yapi: v1
	Chain       []config.ChainStep // Chain steps if this is a chain config
	Base        *config.ConfigV1   // Base config for chain merging
	Expect      config.Expectation // Expectations for single request validation
}

// AnalyzeOptions contains options for analyzing a config.
type AnalyzeOptions struct {
	StrictEnv bool // If true, error on missing env files and disable OS env fallback
}

// HasErrors returns true if there are any error-level diagnostics.
func (a *Analysis) HasErrors() bool {
	for _, d := range a.Diagnostics {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

// JSONOutput is the JSON-serializable output for validation results.
type JSONOutput struct {
	Valid       bool             `json:"valid"`
	Diagnostics []JSONDiagnostic `json:"diagnostics"`
	Warnings    []string         `json:"warnings"`
}

// JSONDiagnostic is a JSON-serializable diagnostic.
type JSONDiagnostic struct {
	Severity string `json:"severity"`
	Field    string `json:"field,omitempty"`
	Message  string `json:"message"`
	Line     int    `json:"line"`
	Col      int    `json:"col"`
}

// ToJSON converts the analysis to a JSON-serializable output.
func (a *Analysis) ToJSON() JSONOutput {
	diags := make([]JSONDiagnostic, 0, len(a.Diagnostics))
	for _, d := range a.Diagnostics {
		diags = append(diags, JSONDiagnostic{
			Severity: d.Severity.String(),
			Field:    d.Field,
			Message:  d.Message,
			Line:     d.Line,
			Col:      d.Col,
		})
	}

	warnings := a.Warnings
	if warnings == nil {
		warnings = []string{}
	}

	return JSONOutput{
		Valid:       !a.HasErrors(),
		Diagnostics: diags,
		Warnings:    warnings,
	}
}

// AnalyzeConfigString is the single entrypoint for analyzing YAML config.
// Both CLI and LSP should call this function.
func AnalyzeConfigString(text string) (*Analysis, error) {
	return AnalyzeConfigStringWithProject(text, nil, "")
}

// AnalyzeConfigStringWithProject analyzes a YAML config with optional project context.
// If project is provided, performs cross-environment variable validation, uses project
// variables from the default environment for resolution, and applies environment defaults.
func AnalyzeConfigStringWithProject(text string, project *config.ProjectConfigV1, projectRoot string) (*Analysis, error) {
	return AnalyzeConfigStringWithProjectAndPath(text, "", project, projectRoot)
}

// AnalyzeConfigStringWithProjectAndPath analyzes a YAML config with optional project context
// and config file path for resolving relative env_files.
func AnalyzeConfigStringWithProjectAndPath(text string, configPath string, project *config.ProjectConfigV1, projectRoot string) (*Analysis, error) {
	return AnalyzeConfigStringWithProjectAndPathAndOptions(text, configPath, project, projectRoot, AnalyzeOptions{})
}

// AnalyzeConfigStringWithProjectAndPathAndOptions analyzes a YAML config with project context and options.
func AnalyzeConfigStringWithProjectAndPathAndOptions(text string, configPath string, project *config.ProjectConfigV1, projectRoot string, opts AnalyzeOptions) (*Analysis, error) {
	var parseRes *config.ParseResult
	var err error

	// If project config is available, use project variables and defaults
	if project != nil {
		// Get the default environment (or first available environment)
		envName := project.DefaultEnvironment
		if envName == "" && len(project.Environments) > 0 {
			// If no default, use the first environment alphabetically for consistency
			envNames := project.ListEnvironments()
			if len(envNames) > 0 {
				envName = envNames[0]
			}
		}

		// Get the environment to extract defaults
		var envDefaults *config.ConfigV1
		if env, ok := project.Environments[envName]; ok {
			// Extract the embedded ConfigV1 from the environment
			envDefaults = &env.ConfigV1
		}

		// Resolve environment variables from project config
		envVars, resolveErr := project.ResolveEnvFiles(projectRoot, envName)
		if resolveErr == nil {
			// Build project-aware resolver
			resolver := BuildProjectResolver(envVars)
			parseRes, err = config.LoadFromStringWithOptions(text, config.LoadOptions{
				ConfigPath: configPath,
				Resolver:   resolver,
				Defaults:   envDefaults,
				StrictEnv:  opts.StrictEnv,
			})
		} else {
			// Fall back to parsing with just defaults if we can't resolve env vars
			parseRes, err = config.LoadFromStringWithOptions(text, config.LoadOptions{
				ConfigPath: configPath,
				Defaults:   envDefaults,
				StrictEnv:  opts.StrictEnv,
			})
		}
	} else {
		// No project config - use path for env_files resolution
		parseRes, err = config.LoadFromStringWithOptions(text, config.LoadOptions{
			ConfigPath: configPath,
			StrictEnv:  opts.StrictEnv,
		})
	}

	if err != nil {
		line := extractLineFromError(err.Error())
		diag := Diagnostic{
			Severity: SeverityError,
			Field:    "",
			Message:  fmt.Sprintf("invalid YAML: %v", err),
			Line:     line,
			Col:      0,
		}
		return &Analysis{Diagnostics: []Diagnostic{diag}}, nil
	}
	return analyzeParsed(text, parseRes, project, projectRoot, configPath), nil
}

// AnalyzeConfigFile loads a file and analyzes it.
// If path is "-", reads from stdin.
func AnalyzeConfigFile(path string) (*Analysis, error) {
	return AnalyzeConfigFileWithOptions(path, AnalyzeOptions{})
}

// AnalyzeConfigFileWithOptions analyzes a config file with custom options.
func AnalyzeConfigFileWithOptions(path string, opts AnalyzeOptions) (*Analysis, error) {
	data, err := utils.ReadInput(path)
	if err != nil {
		diag := Diagnostic{
			Severity: SeverityError,
			Field:    "",
			Message:  fmt.Sprintf("failed to read config: %v", err),
			Line:     0,
			Col:      0,
		}
		return &Analysis{Diagnostics: []Diagnostic{diag}}, nil
	}

	// Use path for resolving relative env_files (unless reading from stdin)
	configPath := ""
	if path != "-" {
		configPath = path
	}

	parseRes, err := config.LoadFromStringWithOptions(string(data), config.LoadOptions{
		ConfigPath: configPath,
		StrictEnv:  opts.StrictEnv,
	})
	if err != nil {
		diag := Diagnostic{
			Severity: SeverityError,
			Field:    "",
			Message:  fmt.Sprintf("failed to load config: %v", err),
			Line:     0,
			Col:      0,
		}
		return &Analysis{Diagnostics: []Diagnostic{diag}}, nil
	}

	return analyzeParsed(string(data), parseRes, nil, "", configPath), nil
}

// analyzeParsed is the common analysis path for both string and file inputs.
// configPath is used for resolving relative env_files paths.
func analyzeParsed(text string, parseRes *config.ParseResult, project *config.ProjectConfigV1, projectRoot string, configPath string) *Analysis {
	var diags []Diagnostic

	// Extract env file variable names from the config for validation
	var envFileVarNames map[string]bool
	if parseRes.Base != nil && len(parseRes.Base.EnvFiles) > 0 {
		envFileVarNames = extractEnvFileVarNames(text)
	}

	// Validate env_files entries exist (with proper line/col positions)
	// These diagnostics supersede the warnings from the loader since they include line numbers
	var envFileDiags []Diagnostic
	if project != nil {
		envFileDiags = ValidateEnvFilesExistFromProject(text, project, projectRoot, "")
	} else if configPath != "" {
		envFileDiags = ValidateEnvFilesExist(text, configPath)
	}
	diags = append(diags, envFileDiags...)

	// Filter out env file warnings from parseRes.Warnings since we have diagnostics with line numbers
	if len(envFileDiags) > 0 {
		filteredWarnings := make([]string, 0, len(parseRes.Warnings))
		for _, w := range parseRes.Warnings {
			if !strings.Contains(w, "env file") || !strings.Contains(w, "not found") {
				filteredWarnings = append(filteredWarnings, w)
			}
		}
		parseRes.Warnings = filteredWarnings
	}

	// Chain config
	if len(parseRes.Chain) > 0 {
		diags = append(diags, validateChain(text, parseRes.Base, parseRes.Chain)...)

		// Use project-aware validation if available
		if project != nil {
			diags = append(diags, ValidateProjectVars(text, project, projectRoot)...)
		} else {
			diags = append(diags, validateEnvVarsWithEnvFiles(text, envFileVarNames)...)
		}

		return &Analysis{
			Chain:       parseRes.Chain,
			Base:        parseRes.Base,
			Diagnostics: diags,
			Warnings:    parseRes.Warnings,
		}
	}

	// Single request config
	req := parseRes.Request

	for _, iss := range ValidateRequest(req) {
		diags = append(diags, Diagnostic{
			Severity: iss.Severity,
			Field:    iss.Field,
			Message:  iss.Message,
			Line:     findFieldLine(text, iss.Field),
			Col:      0,
		})
	}

	diags = append(diags, ValidateGraphQLSyntax(text, req)...)
	diags = append(diags, ValidateJQSyntax(text, req)...)
	diags = append(diags, validateUnknownKeys(text)...)

	// Use project-aware validation if available
	if project != nil {
		diags = append(diags, ValidateProjectVars(text, project, projectRoot)...)
	} else {
		diags = append(diags, validateEnvVarsWithEnvFiles(text, envFileVarNames)...)
	}

	if len(parseRes.Expect.Assert.Body) > 0 {
		diags = append(diags, ValidateChainAssertions(text, parseRes.Expect.Assert.Body, "")...)
	}
	if len(parseRes.Expect.Assert.Headers) > 0 {
		diags = append(diags, ValidateChainAssertions(text, parseRes.Expect.Assert.Headers, "")...)
	}

	return &Analysis{
		Request:     req,
		Diagnostics: diags,
		Warnings:    parseRes.Warnings,
		Expect:      parseRes.Expect,
		Base:        parseRes.Base,
	}
}

// validateUnknownKeys checks for unknown keys in the YAML and returns warnings.
func validateUnknownKeys(text string) []Diagnostic {
	if text == "" {
		return nil
	}

	var raw map[string]any
	if err := yaml.Unmarshal([]byte(text), &raw); err != nil {
		return nil
	}

	unknownKeys := config.FindUnknownKeys(raw)
	var diags []Diagnostic
	for _, key := range unknownKeys {
		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Field:    key,
			Message:  fmt.Sprintf("unknown key '%s' will be ignored", key),
			Line:     findFieldLine(text, key),
			Col:      0,
		})
	}
	return diags
}

// findChainStepLine finds the line number where a chain step with given name starts
func findChainStepLine(text, stepName string) int {
	if text == "" || stepName == "" {
		return -1
	}
	// Look for "- name: stepName" or "name: stepName" pattern
	patterns := []string{
		fmt.Sprintf("- name: %s", stepName),
		fmt.Sprintf("-  name: %s", stepName),
		fmt.Sprintf("name: %s", stepName),
		fmt.Sprintf("- name: \"%s\"", stepName),
		fmt.Sprintf("name: \"%s\"", stepName),
		fmt.Sprintf("- name: '%s'", stepName),
		fmt.Sprintf("name: '%s'", stepName),
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		for _, pattern := range patterns {
			if strings.HasPrefix(trimmed, pattern) {
				return i
			}
		}
	}
	return -1
}

// findValueInText finds the line number where a specific value appears in text
func findValueInText(text, value string) int {
	if text == "" || value == "" {
		return -1
	}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.Contains(line, value) {
			return i
		}
	}
	return -1
}

// validateChain validates chain configuration
func validateChain(text string, base *config.ConfigV1, chain []config.ChainStep) []Diagnostic {
	var diags []Diagnostic
	definedSteps := make(map[string]bool)

	for i, step := range chain {
		stepLine := findChainStepLine(text, step.Name)

		// 1. Check name is present
		if step.Name == "" {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Message:  fmt.Sprintf("step #%d missing 'name'", i+1),
				Line:     stepLine,
				Col:      0,
			})
		} else if definedSteps[step.Name] {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    step.Name,
				Message:  fmt.Sprintf("duplicate step name '%s'", step.Name),
				Line:     stepLine,
				Col:      0,
			})
		}

		// 2. Check URL is present (either in step or in base config)
		hasURL := step.URL != "" || (base != nil && base.URL != "")
		if !hasURL {
			diags = append(diags, Diagnostic{
				Severity: SeverityError,
				Field:    step.Name,
				Message:  fmt.Sprintf("step '%s' missing 'url' (not in step or base config)", step.Name),
				Line:     stepLine,
				Col:      0,
			})
		}

		// 3. Check for references to future steps
		diags = append(diags, scanForUndefinedRefs(text, step.URL, definedSteps, step.Name, "url")...)

		// Check Headers
		for _, v := range step.Headers {
			diags = append(diags, scanForUndefinedRefs(text, v, definedSteps, step.Name, "headers")...)
		}

		// Check Body values recursively (handles nested maps like body.params.track_index)
		diags = append(diags, scanBodyForUndefinedRefs(text, step.Body, definedSteps, step.Name, "body")...)

		// Check JSON field
		if step.JSON != "" {
			diags = append(diags, scanForUndefinedRefs(text, step.JSON, definedSteps, step.Name, "json")...)
		}

		// Check Variables
		for k, v := range step.Variables {
			if s, ok := v.(string); ok {
				diags = append(diags, scanForUndefinedRefs(text, s, definedSteps, step.Name, fmt.Sprintf("variables.%s", k))...)
			}
		}

		// 4. Validate JQ assertions
		if len(step.Expect.Assert.Body) > 0 {
			diags = append(diags, ValidateChainAssertions(text, step.Expect.Assert.Body, step.Name)...)
		}
		if len(step.Expect.Assert.Headers) > 0 {
			diags = append(diags, ValidateChainAssertions(text, step.Expect.Assert.Headers, step.Name)...)
		}

		// 5. Add to defined scope
		if step.Name != "" {
			definedSteps[step.Name] = true
		}
	}
	return diags
}

// scanBodyForUndefinedRefs recursively scans a body map for undefined step references
func scanBodyForUndefinedRefs(text string, body map[string]any, definedSteps map[string]bool, currentStep, path string) []Diagnostic {
	var diags []Diagnostic
	for k, v := range body {
		fieldPath := fmt.Sprintf("%s.%s", path, k)
		switch val := v.(type) {
		case string:
			diags = append(diags, scanForUndefinedRefs(text, val, definedSteps, currentStep, fieldPath)...)
		case map[string]any:
			diags = append(diags, scanBodyForUndefinedRefs(text, val, definedSteps, currentStep, fieldPath)...)
		case []any:
			for i, item := range val {
				itemPath := fmt.Sprintf("%s[%d]", fieldPath, i)
				if s, ok := item.(string); ok {
					diags = append(diags, scanForUndefinedRefs(text, s, definedSteps, currentStep, itemPath)...)
				} else if m, ok := item.(map[string]any); ok {
					diags = append(diags, scanBodyForUndefinedRefs(text, m, definedSteps, currentStep, itemPath)...)
				}
			}
		}
	}
	return diags
}

// scanForUndefinedRefs checks a value string for references to undefined steps
func scanForUndefinedRefs(text, value string, definedSteps map[string]bool, currentStep, fieldName string) []Diagnostic {
	var diags []Diagnostic
	matches := vars.Expansion.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		var key string
		if strings.HasPrefix(match[0], "${") {
			key = match[1]
		} else {
			key = match[2]
		}

		// Only check chain references (containing dot)
		if strings.Contains(key, ".") {
			parts := strings.Split(key, ".")
			refStep := parts[0]

			if !definedSteps[refStep] {
				msg := fmt.Sprintf("step '%s' references '%s' before it is defined", currentStep, refStep)
				if refStep == currentStep {
					msg = fmt.Sprintf("step '%s' cannot reference itself", currentStep)
				}

				// Find the actual line where this reference appears
				line := findValueInText(text, match[0])

				diags = append(diags, Diagnostic{
					Severity: SeverityError,
					Field:    fmt.Sprintf("%s.%s", currentStep, fieldName),
					Message:  msg,
					Line:     line,
					Col:      0,
				})
			}
		}
	}
	return diags
}

// EnvVarInfo holds information about an env var reference for hover/diagnostics
type EnvVarInfo struct {
	Name       string
	Value      string // Empty if not defined
	IsDefined  bool
	Line       int
	Col        int
	StartIndex int
	EndIndex   int
}

// isJQBuiltin returns true if the variable name is a known JQ built-in variable.
// JQ built-ins in yapi include: _headers, _body, _request, _response
func isJQBuiltin(varName string) bool {
	switch varName {
	case "_headers", "_body", "_request", "_response":
		return true
	}
	return false
}

// FindEnvVarRefs finds all environment variable references in text
func FindEnvVarRefs(text string) []EnvVarInfo {
	var refs []EnvVarInfo
	lines := strings.Split(text, "\n")

	// Track if we're inside a graphql block (which uses $var syntax for GraphQL variables)
	inGraphQLBlock := false
	graphqlIndent := 0

	for lineNum, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip YAML comments
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Strip inline comments (everything after # that's not in a string)
		// Simple heuristic: find # that's not inside quotes
		lineWithoutComment := line
		if idx := strings.Index(line, "#"); idx != -1 {
			// Check if the # is inside quotes by counting quotes before it
			beforeHash := line[:idx]
			doubleQuotes := strings.Count(beforeHash, "\"") - strings.Count(beforeHash, "\\\"")
			singleQuotes := strings.Count(beforeHash, "'") - strings.Count(beforeHash, "\\'")
			// If even number of quotes before #, it's not inside a string
			if doubleQuotes%2 == 0 && singleQuotes%2 == 0 {
				lineWithoutComment = line[:idx]
			}
		}

		// Check for graphql: field start
		if strings.HasPrefix(trimmed, "graphql:") {
			inGraphQLBlock = true
			// Find the indentation of the graphql key
			graphqlIndent = len(line) - len(strings.TrimLeft(line, " \t"))
			continue
		}

		// If we're in a graphql block, check if we've exited it
		if inGraphQLBlock {
			// Empty lines stay in block
			if trimmed == "" {
				continue
			}
			// Calculate current line's indentation
			currentIndent := len(line) - len(strings.TrimLeft(line, " \t"))
			// If current indentation is <= graphql key's indentation and line has content,
			// we've exited the block (unless it's a continuation like |)
			if currentIndent <= graphqlIndent && !strings.HasPrefix(trimmed, "|") && !strings.HasPrefix(trimmed, ">") {
				inGraphQLBlock = false
			} else {
				// Still in graphql block - skip $var matching (GraphQL variables)
				continue
			}
		}

		matches := vars.EnvOnly.FindAllStringSubmatchIndex(lineWithoutComment, -1)
		for _, match := range matches {
			// match[0:2] = full match, match[2:4] = ${VAR} capture
			fullStart, fullEnd := match[0], match[1]
			fullMatch := lineWithoutComment[fullStart:fullEnd]

			// Extract variable name from ${VAR}
			var varName string
			if match[2] != -1 {
				varName = lineWithoutComment[match[2]:match[3]]
			}

			if varName == "" {
				continue
			}

			// Check if it's actually an env var (not a chain ref)
			// Chain refs have dots like ${step.field}
			if strings.Contains(fullMatch, ".") {
				continue
			}

			// Skip known JQ built-in variables (_headers, _body, etc.)
			// but allow user environment variables that happen to start with underscore
			if isJQBuiltin(varName) {
				continue
			}

			value := os.Getenv(varName)
			refs = append(refs, EnvVarInfo{
				Name:       varName,
				Value:      value,
				IsDefined:  value != "",
				Line:       lineNum,
				Col:        fullStart,
				StartIndex: fullStart,
				EndIndex:   fullEnd,
			})
		}
	}
	return refs
}

// RedactValue redacts a value for display, showing only first/last chars
func RedactValue(value string) string {
	if value == "" {
		return "(empty)"
	}
	if len(value) <= 4 {
		return strings.Repeat("*", len(value))
	}
	return value[:2] + strings.Repeat("*", len(value)-4) + value[len(value)-2:]
}

// validateEnvVars checks for undefined environment variables and returns warnings
func validateEnvVars(text string) []Diagnostic {
	return validateEnvVarsWithEnvFiles(text, nil)
}

// validateEnvVarsWithEnvFiles checks for undefined environment variables,
// considering variables that will be loaded from env_files.
func validateEnvVarsWithEnvFiles(text string, envFileVarNames map[string]bool) []Diagnostic {
	var diags []Diagnostic

	refs := FindEnvVarRefs(text)
	for _, ref := range refs {
		// Skip if variable is defined in OS env
		if ref.IsDefined {
			continue
		}

		// Skip if variable will be loaded from env_files
		if envFileVarNames != nil && envFileVarNames[ref.Name] {
			continue
		}

		diags = append(diags, Diagnostic{
			Severity: SeverityWarning,
			Field:    ref.Name,
			Message:  fmt.Sprintf("environment variable '%s' is not defined", ref.Name),
			Line:     ref.Line,
			Col:      ref.Col,
		})
	}

	return diags
}

// extractEnvFileVarNames extracts variable names referenced in the config that could be
// satisfied by env_files. This is a heuristic - we mark all referenced variables as
// potentially coming from env_files when env_files is present.
func extractEnvFileVarNames(text string) map[string]bool {
	result := make(map[string]bool)
	refs := FindEnvVarRefs(text)
	for _, ref := range refs {
		result[ref.Name] = true
	}
	return result
}

// EnvFileInfo holds information about an env_files entry for diagnostics
type EnvFileInfo struct {
	Path   string // The path as written in the config
	Line   int    // 0-indexed line number
	Col    int    // 0-indexed column number
	Exists bool   // Whether the resolved file exists
}

// FindEnvFilesInConfig parses the YAML to find env_files entries with their positions
func FindEnvFilesInConfig(text string) []EnvFileInfo {
	var result []EnvFileInfo

	var root yaml.Node
	if err := yaml.Unmarshal([]byte(text), &root); err != nil {
		return result
	}

	// root is a DocumentNode, get the content
	if len(root.Content) == 0 {
		return result
	}

	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return result
	}

	// Find env_files key
	for i := 0; i < len(doc.Content)-1; i += 2 {
		keyNode := doc.Content[i]
		valueNode := doc.Content[i+1]

		if keyNode.Value == "env_files" && valueNode.Kind == yaml.SequenceNode {
			// Process each env file entry
			for _, itemNode := range valueNode.Content {
				if itemNode.Kind == yaml.ScalarNode {
					result = append(result, EnvFileInfo{
						Path: itemNode.Value,
						Line: itemNode.Line - 1, // YAML uses 1-indexed lines
						Col:  itemNode.Column - 1,
					})
				}
			}
			break
		}
	}

	return result
}

// ValidateEnvFilesExist checks that env_files entries exist and returns diagnostics
// configPath is the path to the config file for resolving relative paths
func ValidateEnvFilesExist(text string, configPath string) []Diagnostic {
	var diags []Diagnostic

	envFiles := FindEnvFilesInConfig(text)
	if len(envFiles) == 0 {
		return diags
	}

	// Determine base directory for resolving relative paths
	baseDir := "."
	if configPath != "" {
		baseDir = filepath.Dir(configPath)
	}

	for _, ef := range envFiles {
		// Resolve the path
		filePath := ef.Path
		if !filepath.IsAbs(ef.Path) {
			filePath = filepath.Join(baseDir, ef.Path)
		}

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Field:    "env_files",
				Message:  fmt.Sprintf("env file '%s' not found", ef.Path),
				Line:     ef.Line,
				Col:      ef.Col,
			})
		}
	}

	return diags
}

// ValidateEnvFilesExistFromProject checks env_files existence for project-based configs
func ValidateEnvFilesExistFromProject(text string, project *config.ProjectConfigV1, projectRoot string, envName string) []Diagnostic {
	var diags []Diagnostic

	if project == nil {
		return diags
	}

	// Get environment-specific env_files
	var envFiles []string
	if envName == "" {
		envName = project.DefaultEnvironment
	}
	if env, ok := project.Environments[envName]; ok {
		envFiles = env.EnvFiles
	}

	// Parse YAML to find env_files entries with positions
	envFileInfos := FindEnvFilesInConfig(text)

	// Check project-level env_files
	for _, envFile := range envFiles {
		filePath := envFile
		if !filepath.IsAbs(envFile) {
			filePath = filepath.Join(projectRoot, envFile)
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			// Find the line in the project config - for now, use line 0
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Field:    "env_files",
				Message:  fmt.Sprintf("env file '%s' not found (from project environment '%s')", envFile, envName),
				Line:     0,
				Col:      0,
			})
		}
	}

	// Also check config-level env_files
	for _, ef := range envFileInfos {
		filePath := ef.Path
		if !filepath.IsAbs(ef.Path) {
			filePath = filepath.Join(projectRoot, ef.Path)
		}

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			diags = append(diags, Diagnostic{
				Severity: SeverityWarning,
				Field:    "env_files",
				Message:  fmt.Sprintf("env file '%s' not found", ef.Path),
				Line:     ef.Line,
				Col:      ef.Col,
			})
		}
	}

	return diags
}
