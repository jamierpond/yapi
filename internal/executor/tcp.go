package executor

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"

	"yapi.run/cli/internal/domain"
)

// TCPExecutor handles TCP requests.
type TCPExecutor struct{}

// NewTCPExecutor creates a new TCPExecutor.
func NewTCPExecutor() *TCPExecutor {
	return &TCPExecutor{}
}

// Execute performs a TCP request based on the provided domain.Request.
func (e *TCPExecutor) Execute(ctx context.Context, req *domain.Request) (*domain.Response, error) {
	// Extract metadata
	data := req.Metadata["data"]
	encoding := req.Metadata["encoding"]
	readTimeout, _ := strconv.Atoi(req.Metadata["read_timeout"])
	idleTimeout, _ := strconv.Atoi(req.Metadata["idle_timeout"])
	closeAfterSend, _ := strconv.ParseBool(req.Metadata["close_after_send"])

	// Extract host and port from URL
	target := strings.TrimPrefix(req.URL, "tcp://")
	if !strings.Contains(target, ":") {
		return nil, fmt.Errorf("TCP URL must be in format tcp://host:port, got %s", req.URL)
	}

	// Prepare data to send
	var sendData []byte
	var err error
	if data != "" {
		sendData = []byte(data)
	} else if req.Body != nil {
		var buf bytes.Buffer
		if _, err = io.Copy(&buf, req.Body); err != nil {
			return nil, fmt.Errorf("failed to read request body for TCP: %w", err)
		}
		sendData = buf.Bytes()
	}

	// Handle encoding
	switch encoding {
	case "hex":
		decoded, err := hex.DecodeString(string(sendData))
		if err != nil {
			return nil, fmt.Errorf("failed to decode hex data: %w", err)
		}
		sendData = decoded
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(string(sendData))
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64 data: %w", err)
		}
		sendData = decoded
	case "text", "": // Default is text
		// No special decoding needed
	default:
		return nil, fmt.Errorf("unsupported TCP encoding: %s", encoding)
	}

	// Establish connection
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		return nil, fmt.Errorf("failed to dial TCP target %s: %w", target, err)
	}
	defer conn.Close()

	// Write data if present
	if len(sendData) > 0 {
		_, err := conn.Write(sendData)
		if err != nil {
			return nil, fmt.Errorf("failed to write data to TCP connection: %w", err)
		}
		if closeAfterSend {
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				_ = tcpConn.CloseWrite()
			}
		}
	}

	// Read response
	var respBuf bytes.Buffer

	// Set read deadline
	if readTimeout > 0 {
		conn.SetReadDeadline(time.Now().Add(time.Duration(readTimeout) * time.Second))
	} else if idleTimeout > 0 {
		// This is a simplification. A more robust solution would reset the deadline after each read.
		conn.SetReadDeadline(time.Now().Add(time.Duration(idleTimeout) * time.Millisecond))
	}

	_, err = io.Copy(&respBuf, conn)
	if err != nil {
		// Ignore timeout errors as they are expected when the server doesn't close the connection
		if netErr, ok := err.(net.Error); !ok || !netErr.Timeout() {
			return nil, fmt.Errorf("failed to read from TCP connection: %w", err)
		}
	}

	return &domain.Response{
		StatusCode: 0, // TCP has no status code
		Body:       io.NopCloser(&respBuf),
	}, nil
}
