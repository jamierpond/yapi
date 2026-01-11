package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWaitForHealth_HTTPSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := WaitForHealth(context.Background(), []string{server.URL}, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForHealth_HTTPUnhealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	err := WaitForHealth(context.Background(), []string{server.URL}, 500*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("expected timeout error, got: %v", err)
	}
}

func TestWaitForHealth_HTTPBecomesHealthy(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := WaitForHealth(context.Background(), []string{server.URL}, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts < 3 {
		t.Errorf("expected at least 3 attempts, got %d", attempts)
	}
}

func TestWaitForHealth_MultipleURLs(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server2.Close()

	err := WaitForHealth(context.Background(), []string{server1.URL, server2.URL}, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWaitForHealth_EmptyURLs(t *testing.T) {
	err := WaitForHealth(context.Background(), []string{}, 5*time.Second)
	if err != nil {
		t.Fatalf("expected nil error for empty URLs, got: %v", err)
	}
}

func TestWaitForHealth_InvalidScheme(t *testing.T) {
	err := WaitForHealth(context.Background(), []string{"ftp://example.com"}, 500*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for invalid scheme")
	}
	if !strings.Contains(err.Error(), "unsupported") {
		t.Errorf("expected unsupported scheme error, got: %v", err)
	}
}

func TestWaitForHealth_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := WaitForHealth(ctx, []string{server.URL}, 10*time.Second)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestCheckHTTP_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	err := checkHTTP(context.Background(), server.URL, newHTTPClient())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckHTTP_StatusCodes(t *testing.T) {
	tests := []struct {
		status  int
		wantErr bool
	}{
		{200, false},
		{201, false},
		{204, false},
		{299, false},
		{301, true},
		{400, true},
		{500, true},
		{503, true},
	}

	client := newHTTPClient()
	for _, tt := range tests {
		t.Run(http.StatusText(tt.status), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			err := checkHTTP(context.Background(), server.URL, client)
			if (err != nil) != tt.wantErr {
				t.Errorf("status %d: expected error=%v, got error=%v", tt.status, tt.wantErr, err)
			}
		})
	}
}

func TestCheckTCP_InvalidHost(t *testing.T) {
	err := checkTCP(context.Background(), "localhost:99999")
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}
