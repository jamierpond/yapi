package executor

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"yapi.run/cli/internal/config"
)

// TCPExecutor handles TCP requests.
type TCPExecutor struct{}

// NewTCPExecutor creates a new TCPExecutor.
func NewTCPExecutor() *TCPExecutor {
	return &TCPExecutor{}
}

// Execute performs a TCP request based on the provided YapiConfig.
func (e *TCPExecutor) Execute(cfg *config.YapiConfig) (string, error) {
	// Extract host and port from URL
	target := strings.TrimPrefix(cfg.URL, "tcp://")
	if !strings.Contains(target, ":") {
		return "", fmt.Errorf("TCP URL must be in format tcp://host:port, got %s", cfg.URL)
	}

	// Prepare data to send
	var sendData []byte
	if cfg.Data != "" {
		sendData = []byte(cfg.Data)
	} else if cfg.Body != nil {
		b, err := json.Marshal(cfg.Body)
		if err != nil {
			return "", fmt.Errorf("failed to marshal request body for TCP: %w", err)
		}
		sendData = b
	}

	// Handle encoding
	switch cfg.Encoding {
	case "hex":
		decoded, err := hex.DecodeString(string(sendData))
		if err != nil {
			return "", fmt.Errorf("failed to decode hex data: %w", err)
		}
		sendData = decoded
	case "base64":
		decoded, err := base64.StdEncoding.DecodeString(string(sendData))
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 data: %w", err)
		}
		sendData = decoded
	case "text", "": // Default is text
		// No special decoding needed
	default:
		return "", fmt.Errorf("unsupported TCP encoding: %s", cfg.Encoding)
	}

	// Establish connection with a dial timeout (e.g., 5 seconds for connection setup)
	conn, err := net.DialTimeout("tcp", target, 5*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to dial TCP target %s: %w", target, err)
	}
	defer conn.Close()

	// Write data if present
	if len(sendData) > 0 {
		_, err := conn.Write(sendData)
		if err != nil {
			return "", fmt.Errorf("failed to write data to TCP connection: %w", err)
		}
		if cfg.CloseAfterSend {
			// Explicitly close the write half of the connection
			// This signals to the server that no more data will be sent from client side
			if tcpConn, ok := conn.(*net.TCPConn); ok {
				_ = tcpConn.CloseWrite() // Ignore error as we still want to read response
			}
		}
	}

	// Set read deadline
	if cfg.ReadTimeout > 0 {
		_ = conn.SetReadDeadline(time.Now().Add(time.Duration(cfg.ReadTimeout) * time.Second))
	}

	// Read response
	respBuf := bytes.NewBuffer(nil)
	_, err = io.Copy(respBuf, conn)
	// io.Copy reads until EOF or error, which naturally handles server closing connection or timeout
	// We don't need to explicitly handle close_after_send if io.Copy completes correctly.
	// If server keeps connection open indefinitely, read_timeout will eventually kick in.

	if err != nil && !strings.Contains(err.Error(), "timeout") && err != io.EOF {
		return "", fmt.Errorf("failed to read from TCP connection: %w", err)
	}

	return respBuf.String(), nil
}
