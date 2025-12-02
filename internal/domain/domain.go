package domain

import (
	"io"
	"time"
)

type Request struct {
	URL      string
	Method   string
	Headers  map[string]string
	Body     io.Reader // Streamable body
	Metadata map[string]string
}

type Response struct {
	StatusCode int
	Headers    map[string]string
	Body       io.ReadCloser // Streamable response
	Duration   time.Duration
}
