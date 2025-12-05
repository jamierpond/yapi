package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"yapi.run/cli/internal/domain"
)

// graphqlPayload represents the standard GraphQL JSON envelope
type graphqlPayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLTransport returns a transport function for GraphQL requests.
func GraphQLTransport(client HTTPClient) TransportFunc {
	httpFn := HTTPTransport(client)

	return func(ctx context.Context, req *domain.Request) (*domain.Response, error) {
		// Construct the GraphQL payload
		payload := graphqlPayload{
			Query: req.Metadata["graphql_query"],
		}
		if vars, ok := req.Metadata["graphql_variables"]; ok && vars != "" {
			if err := json.Unmarshal([]byte(vars), &payload.Variables); err != nil {
				return nil, fmt.Errorf("failed to unmarshal graphql variables: %w", err)
			}
		}

		// Marshal to JSON
		jsonBytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal graphql payload: %w", err)
		}

		// Create a new request for HTTP execution
		httpReq := &domain.Request{
			URL:     req.URL,
			Method:  "POST",
			Headers: req.Headers,
			Body:    strings.NewReader(string(jsonBytes)),
		}
		if httpReq.Headers == nil {
			httpReq.Headers = make(map[string]string)
		}
		httpReq.Headers["Content-Type"] = "application/json"

		return httpFn(ctx, httpReq)
	}
}
