package executor_test

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"testing"

	"yapi.run/cli/internal/config"
	"yapi.run/cli/internal/executor"
)

func TestTCPTransport_Echo(t *testing.T) {
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
	yaml := fmt.Sprintf(`
yapi: v1
url: tcp://%s
method: tcp
data: |
  %s
read_timeout: 1
close_after_send: true`, l.Addr().String(), expected)
	res, err := config.LoadFromString(yaml)
	if err != nil {
		t.Fatalf("LoadFromString failed: %v", err)
	}
	req := res.Request

	result, err := executor.TCPTransport(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != expected {
		t.Errorf("Expected response %q, got %q", expected, string(body))
	}
	wg.Wait()
}

func TestTCPTransport_HexEncoding(t *testing.T) {
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

	yaml := fmt.Sprintf(`
yapi: v1
url: tcp://%s
method: tcp
data: "%s"
encoding: hex
read_timeout: 1
close_after_send: true`, l.Addr().String(), hexData)
	res, err := config.LoadFromString(yaml)
	if err != nil {
		t.Fatalf("LoadFromString failed: %v", err)
	}
	req := res.Request

	result, err := executor.TCPTransport(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != expected {
		t.Errorf("Expected response %q, got %q", expected, string(body))
	}
	wg.Wait()
}

func TestTCPTransport_Base64Encoding(t *testing.T) {
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

	yaml := fmt.Sprintf(`
yapi: v1
url: tcp://%s
method: tcp
data: "%s"
encoding: base64
read_timeout: 1
close_after_send: true`, l.Addr().String(), base64Data)
	res, err := config.LoadFromString(yaml)
	if err != nil {
		t.Fatalf("LoadFromString failed: %v", err)
	}
	req := res.Request

	result, err := executor.TCPTransport(context.Background(), req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}
	if string(body) != expected {
		t.Errorf("Expected response %q, got %q", expected, string(body))
	}
	wg.Wait()
}
