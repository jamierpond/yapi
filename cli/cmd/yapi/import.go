package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"yapi.run/cli/internal/cli/color"
	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/importer"
)

// collectUsedVariables extracts all ${var} references from imported configs
func collectUsedVariables(files map[string]config.ConfigV1) map[string]bool {
	varPattern := regexp.MustCompile(`\$\{([^}]+)\}`)
	vars := make(map[string]bool)

	for _, cfg := range files {
		// Check URL
		for _, match := range varPattern.FindAllStringSubmatch(cfg.URL, -1) {
			if len(match) > 1 {
				vars[match[1]] = true
			}
		}

		// Check headers
		for _, v := range cfg.Headers {
			for _, match := range varPattern.FindAllStringSubmatch(v, -1) {
				if len(match) > 1 {
					vars[match[1]] = true
				}
			}
		}

		// Check query params
		for _, v := range cfg.Query {
			for _, match := range varPattern.FindAllStringSubmatch(v, -1) {
				if len(match) > 1 {
					vars[match[1]] = true
				}
			}
		}

		// Check form data
		for _, v := range cfg.Form {
			for _, match := range varPattern.FindAllStringSubmatch(v, -1) {
				if len(match) > 1 {
					vars[match[1]] = true
				}
			}
		}

		// Check JSON body
		if cfg.JSON != "" {
			for _, match := range varPattern.FindAllStringSubmatch(cfg.JSON, -1) {
				if len(match) > 1 {
					vars[match[1]] = true
				}
			}
		}
	}

	return vars
}

// variableCategories holds categorized variables from import
type variableCategories struct {
	configVars  map[string]string
	secretVars  map[string]string
	dynamicVars []string
}

// categorizeImportedVariables separates variables into config, secrets, and dynamic
func categorizeImportedVariables(envResult *importer.EnvironmentImportResult, usedVars map[string]bool) variableCategories {
	categories := variableCategories{
		configVars:  make(map[string]string),
		secretVars:  make(map[string]string),
		dynamicVars: []string{},
	}

	// Add environment variables
	if envResult != nil {
		for k, v := range envResult.ConfigVars {
			categories.configVars[k] = v
		}
		for k, v := range envResult.SecretVars {
			categories.secretVars[k] = v
		}
	}

	// Add undefined variables from collection
	for varName := range usedVars {
		// Skip if already categorized
		if _, exists := categories.configVars[varName]; exists {
			continue
		}
		if _, exists := categories.secretVars[varName]; exists {
			continue
		}

		// Check if this is a Postman dynamic variable
		if strings.HasPrefix(varName, "$") {
			categories.dynamicVars = append(categories.dynamicVars, varName)
		} else {
			categories.configVars[varName] = "" // Empty placeholder
		}
	}

	return categories
}

// writeYapiConfig generates and writes the yapi.config.yml file
func writeYapiConfig(outDir, envName string, configVars, secretVars map[string]string) error {
	yapiConfigPath := filepath.Join(outDir, "yapi.config.yml")
	var yapiConfigContent strings.Builder

	yapiConfigContent.WriteString("yapi: v1\n\n")
	yapiConfigContent.WriteString("# Imported from Postman collection\n")
	if len(secretVars) > 0 {
		yapiConfigContent.WriteString("# Secrets are in .env file - DO NOT commit .env to version control\n")
	}
	yapiConfigContent.WriteString("\n")
	yapiConfigContent.WriteString(fmt.Sprintf("default_environment: %s\n\n", envName))
	yapiConfigContent.WriteString("environments:\n")
	yapiConfigContent.WriteString(fmt.Sprintf("  %s:\n", envName))

	// Add config vars if any
	if len(configVars) > 0 {
		yapiConfigContent.WriteString("    vars:\n")
		// Sort keys for consistent output
		var keys []string
		for k := range configVars {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := configVars[k]
			if v == "" {
				yapiConfigContent.WriteString(fmt.Sprintf("      %s: \"\"\n", k))
			} else {
				yapiConfigContent.WriteString(fmt.Sprintf("      %s: %s\n", k, quoteYAMLValue(v)))
			}
		}
		yapiConfigContent.WriteString("\n")
	}

	// Add env_files reference if there are secrets
	if len(secretVars) > 0 {
		yapiConfigContent.WriteString("    env_files:\n")
		yapiConfigContent.WriteString("      - .env\n")
	}

	if err := os.WriteFile(yapiConfigPath, []byte(yapiConfigContent.String()), 0600); err != nil {
		return fmt.Errorf("failed to write yapi.config.yml: %w", err)
	}

	return nil
}

// writeEnvFile generates and writes the .env file for secrets
func writeEnvFile(outDir string, secretVars map[string]string, dynamicVars []string) error {
	if len(secretVars) == 0 && len(dynamicVars) == 0 {
		return nil
	}

	envFilePath := filepath.Join(outDir, ".env")
	var envContent strings.Builder
	envContent.WriteString("# Secrets from Postman environment\n")
	envContent.WriteString("# DO NOT commit this file to version control!\n")
	envContent.WriteString("# Add .env to your .gitignore\n\n")

	if len(secretVars) > 0 {
		// Sort keys for consistent output
		var keys []string
		for k := range secretVars {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		envContent.WriteString("# Detected secrets (fill in real values):\n")
		for _, k := range keys {
			v := secretVars[k]
			if v == "" {
				envContent.WriteString(fmt.Sprintf("%s=\n", k))
			} else {
				envContent.WriteString(fmt.Sprintf("%s=%s\n", k, v))
			}
		}
		envContent.WriteString("\n")
	}

	if len(dynamicVars) > 0 {
		sort.Strings(dynamicVars)
		envContent.WriteString("# Postman dynamic variables (require manual handling):\n")
		envContent.WriteString("# - $guid: Generate a UUID\n")
		envContent.WriteString("# - $timestamp: Current Unix timestamp\n")
		envContent.WriteString("# - $isoTimestamp: Current ISO 8601 timestamp\n")
		envContent.WriteString("# - $randomInt: Random integer\n")
		for _, varName := range dynamicVars {
			envContent.WriteString(fmt.Sprintf("# %s=\n", varName))
		}
	}

	if err := os.WriteFile(envFilePath, []byte(envContent.String()), 0600); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}

	return nil
}

// writeRequestFiles writes all imported request files
func writeRequestFiles(outDir string, files map[string]config.ConfigV1) (int, error) {
	fileCount := 0
	for relPath, cfg := range files {
		fullPath := filepath.Join(outDir, relPath)

		// Create parent directory
		if err := os.MkdirAll(filepath.Dir(fullPath), 0750); err != nil {
			return 0, fmt.Errorf("failed to create directory for %s: %w", relPath, err)
		}

		// Marshal to YAML
		yamlData, err := yaml.Marshal(cfg)
		if err != nil {
			return 0, fmt.Errorf("failed to marshal config for %s: %w", relPath, err)
		}

		// Write file
		if err := os.WriteFile(fullPath, yamlData, 0600); err != nil {
			return 0, fmt.Errorf("failed to write file %s: %w", relPath, err)
		}

		fileCount++
		fmt.Fprintf(os.Stderr, "  %s %s\n", color.Green("OK"), relPath)
	}
	return fileCount, nil
}

// sanitizeEnvName converts an environment name to a safe identifier
func sanitizeEnvName(name string) string {
	// Replace spaces and special characters with hyphens
	name = strings.ReplaceAll(name, " ", "-")
	name = regexp.MustCompile(`[^a-zA-Z0-9\-_]`).ReplaceAllString(name, "")
	name = strings.ToLower(name)
	if name == "" {
		return "imported"
	}
	return name
}

// quoteYAMLValue properly quotes a YAML value if needed
func quoteYAMLValue(value string) string {
	// If the value contains special characters, quote it
	if strings.ContainsAny(value, ":#[]{}|>*&!%@`") || strings.HasPrefix(value, " ") || strings.HasSuffix(value, " ") {
		return fmt.Sprintf("\"%s\"", strings.ReplaceAll(value, "\"", "\\\""))
	}
	// If it looks like a number or boolean, quote it to keep it as string
	if value == "true" || value == "false" || regexp.MustCompile(`^\d+$`).MatchString(value) {
		return fmt.Sprintf("\"%s\"", value)
	}
	return value
}

// importE handles the import command to convert external collections to yapi format
func importE(cmd *cobra.Command, args []string) error {
	inputPath := args[0]
	outDir, _ := cmd.Flags().GetString("output")
	envPath, _ := cmd.Flags().GetString("env")

	// Check if input file exists
	if _, err := os.Stat(inputPath); err != nil {
		return fmt.Errorf("input file not found: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\n%s\n", color.Accent("yapi import"))
	fmt.Fprintf(os.Stderr, "%s\n", color.Dim("Importing Postman collection..."))
	fmt.Fprintf(os.Stderr, "\n")

	// Import the collection
	result, err := importer.ImportPostmanCollection(inputPath)
	if err != nil {
		return fmt.Errorf("failed to import collection: %w", err)
	}

	if len(result.Files) == 0 {
		fmt.Fprintf(os.Stderr, "%s\n", color.Yellow("No requests found in collection"))
		return nil
	}

	// Import environment file if specified
	var envResult *importer.EnvironmentImportResult
	if envPath != "" {
		if _, err := os.Stat(envPath); err != nil {
			return fmt.Errorf("environment file not found: %w", err)
		}
		envResult, err = importer.ImportPostmanEnvironment(envPath)
		if err != nil {
			return fmt.Errorf("failed to import environment: %w", err)
		}

		totalVars := len(envResult.ConfigVars) + len(envResult.SecretVars)
		fmt.Fprintf(os.Stderr, "%s Imported %d variables (%d config, %d secrets)\n",
			color.Green("OK"), totalVars, len(envResult.ConfigVars), len(envResult.SecretVars))

		// Show warnings about detected secrets
		if len(envResult.SecretWarnings) > 0 {
			fmt.Fprintf(os.Stderr, "\n%s\n", color.Yellow("Warning: Security Warnings:"))
			for _, warning := range envResult.SecretWarnings {
				fmt.Fprintf(os.Stderr, "  %s\n", color.Yellow("- "+warning))
			}
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Create output directory
	if err := os.MkdirAll(outDir, 0750); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Collect and categorize all variables
	usedVars := collectUsedVariables(result.Files)
	envName := "imported"
	if envResult != nil && envResult.Name != "" {
		envName = sanitizeEnvName(envResult.Name)
	}
	categories := categorizeImportedVariables(envResult, usedVars)

	// Write configuration files
	if err := writeYapiConfig(outDir, envName, categories.configVars, categories.secretVars); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "  %s yapi.config.yml (%d config variables)\n", color.Green("OK"), len(categories.configVars))

	if err := writeEnvFile(outDir, categories.secretVars, categories.dynamicVars); err != nil {
		return err
	}
	if len(categories.secretVars) > 0 || len(categories.dynamicVars) > 0 {
		if len(categories.dynamicVars) > 0 {
			fmt.Fprintf(os.Stderr, "  %s .env (%d secrets, %d dynamic variables)\n",
				color.Green("OK"), len(categories.secretVars), len(categories.dynamicVars))
		} else {
			fmt.Fprintf(os.Stderr, "  %s .env (%d secrets)\n", color.Green("OK"), len(categories.secretVars))
		}
	}

	// Write request files
	fileCount, err := writeRequestFiles(outDir, result.Files)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "%s\n", color.Green(fmt.Sprintf("Successfully imported %d request(s) to %s", fileCount, outDir)))
	fmt.Fprintf(os.Stderr, "\n")

	return nil
}
