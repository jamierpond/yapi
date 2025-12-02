package executor

import (
	"encoding/json"
	"fmt"

	"yapi.run/cli/internal/config"
)

// graphqlPayload represents the standard GraphQL JSON envelope
type graphqlPayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLExecutor handles GraphQL requests by wrapping HTTP
type GraphQLExecutor struct {
	httpExec *HTTPExecutor
}

// NewGraphQLExecutor creates a new GraphQLExecutor
func NewGraphQLExecutor() *GraphQLExecutor {
	return &GraphQLExecutor{
		httpExec: NewHTTPExecutor(),
	}
}

// Execute performs a GraphQL request
func (e *GraphQLExecutor) Execute(originalCfg *config.Config) (*HTTPResponse, error) {
	// Create a deep copy of the config to avoid mutating the original
	cfg := *originalCfg

	// Construct the GraphQL payload
	payload := graphqlPayload{
		Query:     cfg.Graphql,
		Variables: cfg.Variables,
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal graphql payload: %w", err)
	}

	// Modify the copy for HTTP execution
	cfg.Method = "POST"
	cfg.Body = string(jsonBytes)
	cfg.ContentType = "application/json"
	// Clear GraphQL fields to avoid confusion
	cfg.Graphql = ""
	cfg.Variables = nil

	return e.httpExec.Execute(&cfg)
}
