package importer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"yapi.run/cli/internal/config"
)

// ImportResult represents the result of importing a collection
type ImportResult struct {
	Files       map[string]config.ConfigV1 // relative path -> config
	Environment map[string]string          // environment variables
}

// EnvironmentImportResult represents the result of importing a Postman environment
type EnvironmentImportResult struct {
	Name             string            // Environment name from Postman
	ConfigVars       map[string]string // Non-secret vars for yapi.config.yml
	SecretVars       map[string]string // Secret vars for .env file
	SecretWarnings   []string          // Warnings about detected secrets
	UndefinedSecrets []string          // Current-only secrets (no initial value)
}

// ImportPostmanCollection imports a Postman collection from a JSON file
func ImportPostmanCollection(filePath string) (*ImportResult, error) {
	data, err := os.ReadFile(filePath) // #nosec G304 -- filePath is validated user-provided file path
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var collection PostmanCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse Postman collection: %w", err)
	}

	result := &ImportResult{
		Files:       make(map[string]config.ConfigV1),
		Environment: make(map[string]string),
	}

	// Convert all items in the collection
	convertItems(collection.Item, "", result)

	return result, nil
}

// ImportPostmanEnvironment imports a Postman environment file
// Returns structured data separating config vars from secrets
func ImportPostmanEnvironment(filePath string) (*EnvironmentImportResult, error) {
	data, err := os.ReadFile(filePath) // #nosec G304 -- filePath is validated user-provided file path
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file: %w", err)
	}

	var env PostmanEnvironment
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("failed to parse Postman environment: %w", err)
	}

	result := &EnvironmentImportResult{
		Name:             env.Name,
		ConfigVars:       make(map[string]string),
		SecretVars:       make(map[string]string),
		SecretWarnings:   []string{},
		UndefinedSecrets: []string{},
	}

	for _, v := range env.Values {
		if !v.Enabled {
			continue
		}

		// Determine the effective value
		// If only "value" (current) is set, it's likely a secret
		// If "initial" is set, it's intended to be shared
		hasInitial := v.Initial != ""
		hasCurrent := v.Value != ""

		effectiveValue := v.Initial
		if effectiveValue == "" {
			effectiveValue = v.Value
		}

		isSecret := looksLikeSecret(v.Key, effectiveValue)

		// Classify the variable
		// Priority: explicit secret detection over initial/current distinction
		if isSecret {
			if hasInitial {
				// Secret in initial value - warn user
				result.SecretWarnings = append(result.SecretWarnings,
					fmt.Sprintf("Variable '%s' looks like a secret but is in 'initial' value (will be shared)", v.Key))
			}
			result.SecretVars[v.Key] = effectiveValue
		} else if !hasInitial && hasCurrent {
			// Current-only variable that doesn't look like a secret
			// This is likely a config var that wasn't properly exported
			// Add to config vars but note it came from current
			result.ConfigVars[v.Key] = effectiveValue
		} else {
			// Regular config variable with initial value
			result.ConfigVars[v.Key] = effectiveValue
		}
	}

	return result, nil
}

// looksLikeSecret detects if a variable name or value looks like a secret
func looksLikeSecret(key, value string) bool {
	lowerKey := strings.ToLower(key)

	// Exact matches for common secret variable names
	exactMatches := []string{
		"jwt", "token", "apikey", "api_key", "secret", "password",
		"passwd", "pwd", "auth", "bearer", "session", "cookie",
	}
	for _, exact := range exactMatches {
		if lowerKey == exact {
			return true
		}
	}

	// Keywords that indicate secrets when part of the name
	// Use word boundaries to avoid false positives
	secretKeywords := []string{
		"_token", "_key", "_secret", "_password", "_pwd", "_auth",
		"_credential", "_private", "_session", "_cookie", "_bearer",
		"token_", "key_", "secret_", "password_", "pwd_", "auth_",
		"credential_", "private_", "session_", "cookie_", "bearer_",
	}

	for _, keyword := range secretKeywords {
		if strings.Contains(lowerKey, keyword) {
			return true
		}
	}

	// Check for high-entropy values (likely tokens/keys)
	// Only check if value is reasonably long and looks random
	if len(value) > 32 && hasHighEntropy(value) {
		return true
	}

	return false
}

// hasHighEntropy checks if a string has high entropy (random-looking)
func hasHighEntropy(s string) bool {
	if len(s) == 0 {
		return false
	}

	// Simple entropy check: count unique characters
	// High-entropy strings have many unique characters
	uniqueChars := make(map[rune]bool)
	for _, c := range s {
		uniqueChars[c] = true
	}

	// If more than 70% of characters are unique, it's likely high-entropy
	uniqueRatio := float64(len(uniqueChars)) / float64(len(s))
	return uniqueRatio > 0.7 && len(s) > 20
}

// convertItems recursively converts Postman items to yapi configs
func convertItems(items []PostmanItem, basePath string, result *ImportResult) {
	for _, item := range items {
		// If this item has a request, convert it
		if item.Request != nil {
			cfg := convertRequest(item.Name, item.Request)

			// Generate file path
			fileName := sanitizeFileName(item.Name) + ".yapi.yml"
			filePath := filepath.Join(basePath, fileName)

			result.Files[filePath] = cfg
		}

		// If this item has sub-items (folder), recurse
		if len(item.Item) > 0 {
			folderPath := filepath.Join(basePath, sanitizeFileName(item.Name))
			convertItems(item.Item, folderPath, result)
		}
	}
}

// convertRequest converts a single Postman request to a yapi ConfigV1
func convertRequest(name string, req *PostmanRequest) config.ConfigV1 {
	cfg := config.ConfigV1{
		Yapi:   "v1",
		Method: strings.ToUpper(req.Method),
		URL:    convertURL(req.URL),
	}

	// Convert query parameters
	queryParams := extractQueryParams(req.URL)
	if len(queryParams) > 0 {
		cfg.Query = queryParams
	}

	// Convert headers
	if len(req.Header) > 0 {
		cfg.Headers = make(map[string]string)
		for _, h := range req.Header {
			if !h.Disabled {
				cfg.Headers[h.Key] = convertVariables(h.Value)
			}
		}
	}

	// Convert body based on mode
	if req.Body != nil {
		switch req.Body.Mode {
		case "raw":
			if req.Body.Raw != "" {
				rawBody := convertVariables(req.Body.Raw)

				// Determine if it's JSON
				isJSON := false
				if req.Body.Options != nil && req.Body.Options.Raw != nil {
					isJSON = req.Body.Options.Raw.Language == "json"
				}

				// Try to parse as JSON and use body field for better yapi experience
				if isJSON || isJSONString(rawBody) {
					var bodyData any
					if err := json.Unmarshal([]byte(rawBody), &bodyData); err == nil {
						// Successfully parsed - use body field if it's an object
						if bodyMap, ok := bodyData.(map[string]any); ok {
							cfg.Body = bodyMap
						} else {
							// Arrays or other types - use json field
							cfg.JSON = rawBody
						}
					} else {
						// Failed to parse - fall back to json field
						cfg.JSON = rawBody
					}

					// Set content type if not already set
					if cfg.Headers == nil {
						cfg.Headers = make(map[string]string)
					}
					if _, hasContentType := cfg.Headers["Content-Type"]; !hasContentType {
						cfg.ContentType = "application/json"
					}
				} else {
					// Not JSON - use json field for raw text
					cfg.JSON = rawBody
				}
			}

		case "urlencoded":
			if len(req.Body.URLEncoded) > 0 {
				cfg.Form = make(map[string]string)
				for _, field := range req.Body.URLEncoded {
					if !field.Disabled && field.Key != "" {
						cfg.Form[field.Key] = convertVariables(field.Value)
					}
				}
				cfg.ContentType = "application/x-www-form-urlencoded"
			}

		case "formdata":
			if len(req.Body.FormData) > 0 {
				cfg.Form = make(map[string]string)
				for _, field := range req.Body.FormData {
					// Only handle text fields for now (skip file uploads)
					if !field.Disabled && field.Key != "" && field.Type != "file" {
						cfg.Form[field.Key] = convertVariables(field.Value)
					}
				}
				// Only set content type if we actually have form fields
				if len(cfg.Form) > 0 {
					cfg.ContentType = "multipart/form-data"
				}
			}
		}
	}

	return cfg
}

// convertURL converts a Postman URL to a string, replacing variables
// Note: Query parameters are stripped and should be extracted separately via extractQueryParams
func convertURL(url PostmanURL) string {
	var baseURL string

	if url.Raw != "" {
		baseURL = convertVariables(url.Raw)
		// Strip query parameters from raw URL
		if idx := strings.Index(baseURL, "?"); idx != -1 {
			baseURL = baseURL[:idx]
		}
		return baseURL
	}

	// Construct from parts if raw is not available
	var urlStr strings.Builder

	if url.Protocol != "" {
		urlStr.WriteString(url.Protocol)
		urlStr.WriteString("://")
	}

	if len(url.Host) > 0 {
		urlStr.WriteString(convertVariables(strings.Join(url.Host, ".")))
	}

	if len(url.Path) > 0 {
		urlStr.WriteString("/")
		urlStr.WriteString(strings.Join(url.Path, "/"))
	}

	return urlStr.String()
}

// extractQueryParams extracts query parameters from a Postman URL
func extractQueryParams(url PostmanURL) map[string]string {
	params := make(map[string]string)

	// First, extract from the Query array if present
	for _, q := range url.Query {
		if !q.Disabled && q.Key != "" {
			params[q.Key] = convertVariables(q.Value)
		}
	}

	// If no Query array, try to parse from raw URL
	if len(params) == 0 && url.Raw != "" {
		if idx := strings.Index(url.Raw, "?"); idx != -1 {
			queryString := url.Raw[idx+1:]
			// Parse query string manually
			pairs := strings.Split(queryString, "&")
			for _, pair := range pairs {
				if pair == "" {
					continue
				}
				kv := strings.SplitN(pair, "=", 2)
				if len(kv) >= 1 {
					key := kv[0]
					value := ""
					if len(kv) == 2 {
						value = kv[1]
					}
					params[key] = convertVariables(value)
				}
			}
		}
	}

	return params
}

// convertVariables converts Postman variable syntax to yapi syntax
// {{variable}} -> ${variable}
// Postman dynamic variables ({{$guid}}, {{$timestamp}}, etc.) are converted
// but will need manual handling as yapi doesn't have built-in dynamic variables
func convertVariables(s string) string {
	// Replace {{variable}} with ${variable}
	// This includes Postman dynamic variables like:
	// - {{$guid}} -> ${$guid} (will need manual UUID generation)
	// - {{$timestamp}} -> ${$timestamp} (will need manual timestamp)
	// - {{$randomInt}} -> ${$randomInt} (will need manual random number)
	// - {{$isoTimestamp}} -> ${$isoTimestamp} (will need manual ISO timestamp)
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	return re.ReplaceAllString(s, `${$1}`)
}

// sanitizeFileName converts a name to a safe filename
// Prevents directory traversal attacks by removing dots and path separators
func sanitizeFileName(name string) string {
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove dots, slashes, and other special characters to prevent directory traversal
	// Only allow alphanumeric, hyphens, and underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_]`)
	name = re.ReplaceAllString(name, "")

	// Convert to lowercase
	name = strings.ToLower(name)

	// Ensure we only use the base name (additional safety against path injection)
	name = filepath.Base(name)

	// Prevent empty filenames
	if name == "" || name == "." || name == ".." {
		name = "unnamed"
	}

	// Limit length
	if len(name) > 200 {
		name = name[:200]
	}

	return name
}

// isJSONString checks if a string looks like JSON
func isJSONString(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}")) ||
		(strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"))
}
