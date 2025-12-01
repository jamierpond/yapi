package executor_test

import (
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	"cli/internal/config"
	"cli/internal/executor"
)

func TestTCPExecutor_Execute_Echo(t *testing.T) {
	expected := "Hello from yapi!\n"

	// Mock TCP server
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		// Read data from client
		received, err := io.ReadAll(conn)
		if err != nil {
			t.Errorf("Server failed to read from client: %v", err)
			return
		}
		if string(received) != expected {
			t.Errorf("Server expected %q, got %q", expected, string(received))
		}

		// Echo data back
		_, err = conn.Write(received)
		if err != nil {
			t.Errorf("Server failed to write to client: %v", err)
		}
	}()

	// Client configuration
	cfg := &config.YapiConfig{
		URL:            fmt.Sprintf("tcp://%s", l.Addr().String()),
		Method:         "tcp",
		Data:           expected,
		Encoding:       "text",
		ReadTimeout:    1,    // Short timeout for testing
		CloseAfterSend: true, // Should close write half
	}

	exec := executor.NewTCPExecutor()
	result, err := exec.Execute(cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected response %q, got %q", expected, result)
	}
	wg.Wait()
}

func TestTCPExecutor_Execute_HexEncoding(t *testing.T) {
	hexData := "48656c6c6f"
	expected := "Hello"

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		received, err := io.ReadAll(conn)
		if err != nil {
			t.Errorf("Server failed to read from client: %v", err)
			return
		}
		if string(received) != expected {
			t.Errorf("Server expected %q, got %q", expected, string(received))
		}
		_, err = conn.Write(received)
		if err != nil {
			t.Errorf("Server failed to write to client: %v", err)
		}
	}()

	cfg := &config.YapiConfig{
		URL:            fmt.Sprintf("tcp://%s", l.Addr().String()),
		Method:         "tcp",
		Data:           hexData,
		Encoding:       "hex",
		ReadTimeout:    1,
		CloseAfterSend: true,
	}

	exec := executor.NewTCPExecutor()
	result, err := exec.Execute(cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected response %q, got %q", expected, result)
	}
	wg.Wait()
}

func TestTCPExecutor_Execute_Base64Encoding(t *testing.T) {
	base64Data := "SGVsbG8="
	expected := "Hello"

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	defer l.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		received, err := io.ReadAll(conn)
		if err != nil {
			t.Errorf("Server failed to read from client: %v", err)
			return
		}
		if string(received) != expected {
			t.Errorf("Server expected %q, got %q", expected, string(received))
		}
		_, err = conn.Write(received)
		if err != nil {
			t.Errorf("Server failed to write to client: %v", err)
		}
	}()

	cfg := &config.YapiConfig{
		URL:            fmt.Sprintf("tcp://%s", l.Addr().String()),
		Method:         "tcp",
		Data:           base64Data,
		Encoding:       "base64",
		ReadTimeout:    1,
		CloseAfterSend: true,
	}

	exec := executor.NewTCPExecutor()
	result, err := exec.Execute(cfg)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result != expected {
		t.Errorf("Expected response %q, got %q", expected, result)
	}
	wg.Wait()
}
