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
func ImportPostmanEnvironment(filePath string) (map[string]string, error) {
	data, err := os.ReadFile(filePath) // #nosec G304 -- filePath is validated user-provided file path
	if err != nil {
		return nil, fmt.Errorf("failed to read environment file: %w", err)
	}

	var env PostmanEnvironment
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, fmt.Errorf("failed to parse Postman environment: %w", err)
	}

	envVars := make(map[string]string)
	for _, v := range env.Values {
		if v.Enabled {
			envVars[v.Key] = v.Value
		}
	}

	return envVars, nil
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

	// Convert body
	if req.Body != nil && req.Body.Mode == "raw" && req.Body.Raw != "" {
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

// convertVariables converts Postman variable syntax {{var}} to yapi syntax ${var}
func convertVariables(s string) string {
	// Replace {{variable}} with ${variable}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	return re.ReplaceAllString(s, `${$1}`)
}

// sanitizeFileName converts a name to a safe filename
func sanitizeFileName(name string) string {
	// Replace spaces with hyphens
	name = strings.ReplaceAll(name, " ", "-")

	// Remove or replace special characters
	re := regexp.MustCompile(`[^a-zA-Z0-9\-_.]`)
	name = re.ReplaceAllString(name, "")

	// Convert to lowercase
	name = strings.ToLower(name)

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
