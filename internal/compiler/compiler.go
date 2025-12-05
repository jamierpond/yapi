// Package compiler transforms ConfigV1 into domain.Request via recursive interpolation and validation.
// This is the Single Source of Truth for both CLI runtime and LSP.
package compiler

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/constants"
	"yapi.run/cli/internal/domain"
	"yapi.run/cli/internal/utils"
	"yapi.run/cli/internal/vars"
)

// CompiledRequest is the result of compiling a ConfigV1.
type CompiledRequest struct {
	Request *domain.Request
	Errors  []error
}

// HasErrors returns true if compilation produced errors.
func (c *CompiledRequest) HasErrors() bool {
	return len(c.Errors) > 0
}

// Compile transforms ConfigV1 -> domain.Request via recursive interpolation + validation.
func Compile(cfg *config.ConfigV1, resolver vars.Resolver) *CompiledRequest {
	res := &CompiledRequest{}

	// 1. Recursive Interpolation
	interpolated, err := resolveConfig(cfg, resolver)
	if err != nil {
		res.Errors = append(res.Errors, err)
		return res
	}

	// 2. Canonicalize
	if interpolated.Method == "" {
		interpolated.Method = constants.MethodGET
	}
	interpolated.Method = constants.CanonicalizeMethod(interpolated.Method)

	// 3. Construct Domain Object
	req := &domain.Request{
		Method:   interpolated.Method,
		Headers:  interpolated.Headers,
		Metadata: make(map[string]string),
	}

	// 4. URL Construction
	fullURL := interpolated.URL
	if interpolated.Path != "" {
		fullURL += interpolated.Path
	}
	if len(interpolated.Query) > 0 {
		q := url.Values{}
		for k, v := range interpolated.Query {
			q.Set(k, v)
		}
		if strings.Contains(fullURL, "?") {
			fullURL += "&" + q.Encode()
		} else {
			fullURL += "?" + q.Encode()
		}
	}
	req.URL = fullURL

	// 5. Body Handling
	if interpolated.JSON != "" && interpolated.Body != nil && len(interpolated.Body) > 0 {
		res.Errors = append(res.Errors, fmt.Errorf("`body` and `json` are mutually exclusive"))
	}

	if interpolated.JSON != "" {
		req.Body = strings.NewReader(interpolated.JSON)
		req.Metadata["body_source"] = "json"
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Content-Type"] = utils.Coalesce(req.Headers["Content-Type"], "application/json")
	} else if interpolated.Body != nil {
		bodyBytes, err := json.Marshal(interpolated.Body)
		if err != nil {
			res.Errors = append(res.Errors, fmt.Errorf("invalid json in 'body' field: %w", err))
		} else {
			req.Body = strings.NewReader(string(bodyBytes))
			if req.Headers == nil {
				req.Headers = make(map[string]string)
			}
			req.Headers["Content-Type"] = utils.Coalesce(req.Headers["Content-Type"], "application/json")
		}
	}

	// Content-Type override
	if interpolated.ContentType != "" {
		if req.Headers == nil {
			req.Headers = make(map[string]string)
		}
		req.Headers["Content-Type"] = interpolated.ContentType
	}

	// 6. Protocol Detection and Validation
	transport := detectTransport(req.URL, interpolated)
	req.Metadata["transport"] = transport

	switch transport {
	case constants.TransportGRPC:
		if interpolated.Service == "" {
			res.Errors = append(res.Errors, fmt.Errorf("gRPC requires 'service'"))
		}
		if interpolated.RPC == "" {
			res.Errors = append(res.Errors, fmt.Errorf("gRPC requires 'rpc'"))
		}
		req.Metadata["service"] = interpolated.Service
		req.Metadata["rpc"] = interpolated.RPC
		req.Metadata["proto"] = interpolated.Proto
		req.Metadata["proto_path"] = interpolated.ProtoPath
		req.Metadata["insecure"] = fmt.Sprintf("%t", interpolated.Insecure)
		req.Metadata["plaintext"] = fmt.Sprintf("%t", interpolated.Plaintext)

	case constants.TransportTCP:
		if interpolated.Encoding != "" && !isValidEncoding(interpolated.Encoding) {
			res.Errors = append(res.Errors, fmt.Errorf("invalid encoding '%s'", interpolated.Encoding))
		}
		req.Metadata["data"] = interpolated.Data
		req.Metadata["encoding"] = interpolated.Encoding
		req.Metadata["read_timeout"] = fmt.Sprintf("%d", interpolated.ReadTimeout)
		req.Metadata["idle_timeout"] = fmt.Sprintf("%d", interpolated.IdleTimeout)
		req.Metadata["close_after_send"] = fmt.Sprintf("%t", interpolated.CloseAfterSend)
	}

	// JQ Filter
	if interpolated.JQFilter != "" {
		req.Metadata["jq_filter"] = interpolated.JQFilter
	}

	// GraphQL
	if interpolated.Graphql != "" {
		req.Metadata["graphql_query"] = interpolated.Graphql
		if interpolated.Variables != nil {
			varsJSON, err := json.Marshal(interpolated.Variables)
			if err != nil {
				res.Errors = append(res.Errors, fmt.Errorf("could not marshal graphql variables: %w", err))
			} else {
				req.Metadata["graphql_variables"] = string(varsJSON)
			}
		}
	}

	res.Request = req
	return res
}

// resolveConfig clones the config and walks it with the resolver.
func resolveConfig(cfg *config.ConfigV1, resolver vars.Resolver) (*config.ConfigV1, error) {
	clone := *cfg
	var err error

	if clone.URL, err = vars.ExpandString(clone.URL, resolver); err != nil {
		return nil, fmt.Errorf("url: %w", err)
	}
	if clone.Path, err = vars.ExpandString(clone.Path, resolver); err != nil {
		return nil, fmt.Errorf("path: %w", err)
	}
	if clone.JSON, err = vars.ExpandString(clone.JSON, resolver); err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}
	if clone.Data, err = vars.ExpandString(clone.Data, resolver); err != nil {
		return nil, fmt.Errorf("data: %w", err)
	}

	// Walk map[string]string fields
	if clone.Headers, err = walkStringMap(clone.Headers, resolver); err != nil {
		return nil, fmt.Errorf("headers: %w", err)
	}
	if clone.Query, err = walkStringMap(clone.Query, resolver); err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}

	// Walk deep maps
	if clone.Body != nil {
		clone.Body, err = walkDeep(clone.Body, resolver)
		if err != nil {
			return nil, fmt.Errorf("body: %w", err)
		}
	}
	if clone.Variables != nil {
		clone.Variables, err = walkDeep(clone.Variables, resolver)
		if err != nil {
			return nil, fmt.Errorf("variables: %w", err)
		}
	}

	return &clone, nil
}

// walkStringMap interpolates all values in a string map.
func walkStringMap(m map[string]string, resolver vars.Resolver) (map[string]string, error) {
	if m == nil {
		return nil, nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		expanded, err := vars.ExpandString(v, resolver)
		if err != nil {
			return nil, fmt.Errorf("key '%s': %w", k, err)
		}
		out[k] = expanded
	}
	return out, nil
}

// walkDeep recursively interpolates interface{} maps/slices.
func walkDeep(v map[string]interface{}, resolver vars.Resolver) (map[string]interface{}, error) {
	if v == nil {
		return nil, nil
	}
	out := make(map[string]interface{}, len(v))
	for k, sv := range v {
		res, err := walkValue(sv, resolver)
		if err != nil {
			return nil, fmt.Errorf("key '%s': %w", k, err)
		}
		out[k] = res
	}
	return out, nil
}

// walkValue recursively interpolates a single value.
func walkValue(v interface{}, resolver vars.Resolver) (interface{}, error) {
	switch val := v.(type) {
	case string:
		return vars.ExpandString(val, resolver)
	case map[string]interface{}:
		return walkDeep(val, resolver)
	case []interface{}:
		out := make([]interface{}, len(val))
		for i, sv := range val {
			res, err := walkValue(sv, resolver)
			if err != nil {
				return nil, err
			}
			out[i] = res
		}
		return out, nil
	default:
		return val, nil
	}
}

func detectTransport(u string, c *config.ConfigV1) string {
	urlLower := strings.ToLower(u)
	if strings.HasPrefix(urlLower, "grpc://") || strings.HasPrefix(urlLower, "grpcs://") {
		return constants.TransportGRPC
	}
	if strings.HasPrefix(urlLower, "tcp://") {
		return constants.TransportTCP
	}
	if c.Graphql != "" {
		return constants.TransportGraphQL
	}
	return constants.TransportHTTP
}

func isValidEncoding(e string) bool {
	return e == "text" || e == "hex" || e == "base64"
}
