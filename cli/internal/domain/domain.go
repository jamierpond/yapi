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

// SetHeader sets a header value, initializing the Headers map if needed.
func (r *Request) SetHeader(key, value string) {
	if r.Headers == nil {
		r.Headers = make(map[string]string)
	}
	r.Headers[key] = value
}

// Response represents the result of an API request.
type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       io.ReadCloser // Streamable response
	Duration   time.Duration
}
