// Package domain defines core request and response types.
package domain

import (
	"io"
	"time"
)

// Request represents an outgoing API request.
type Request struct {
	URL      string
	Method   string
	Headers  map[string]string
	Body     io.Reader // Streamable body
	Metadata map[string]string
}

// Response represents the result of an API request.
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       io.ReadCloser // Streamable response
	Duration   time.Duration
}
