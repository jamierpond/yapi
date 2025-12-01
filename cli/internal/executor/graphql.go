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
func (e *GraphQLExecutor) Execute(cfg *config.YapiConfig) (*HTTPResponse, error) {
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

	// Create a copy of the config for HTTP execution
	httpCfg := *cfg
	httpCfg.Method = "POST"
	httpCfg.JSON = string(jsonBytes)
	httpCfg.ContentType = "application/json"
	// Clear GraphQL fields to avoid confusion
	httpCfg.Graphql = ""
	httpCfg.Variables = nil

	return e.httpExec.Execute(&httpCfg)
}
