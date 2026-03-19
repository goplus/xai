package qiniu

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func newSilentLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

func TestClientPostSendsHeadersAndBody(t *testing.T) {
	const apiKey = "client-token"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/voice/asr" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+apiKey {
			t.Fatalf("unexpected auth header: %q", got)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Fatalf("unexpected content type: %q", got)
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	client := NewClient(apiKey,
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
		WithLogger(newSilentLogger()),
		WithDebugLog(false),
	)

	resp, err := client.Post(context.Background(), "/voice/asr", map[string]any{"hello": "world"})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	if got := string(resp); got != `{"ok":true}` {
		t.Fatalf("unexpected response: %s", got)
	}
}

func TestClientRetriesOnRetryableStatus(t *testing.T) {
	var attempts atomic.Int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := attempts.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"message":"retry later"}`))
			return
		}
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer ts.Close()

	client := NewClient("retry-token",
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
		WithRetry(2, time.Millisecond),
		WithLogger(newSilentLogger()),
		WithDebugLog(false),
	)

	resp, err := client.Get(context.Background(), "/voice/list")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got := string(resp); got != `{"ok":true}` {
		t.Fatalf("unexpected response: %s", got)
	}
	if got := attempts.Load(); got != 3 {
		t.Fatalf("expected 3 attempts, got %d", got)
	}
}

func TestClientReturnsStructuredAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", "req-audio-1")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":{"code":"bad_request","message":"invalid audio","type":"invalid_request_error"}}`))
	}))
	defer ts.Close()

	client := NewClient("err-token",
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
		WithLogger(newSilentLogger()),
		WithDebugLog(false),
	)

	_, err := client.Post(context.Background(), "/voice/asr", map[string]any{"bad": true})
	if err == nil {
		t.Fatal("expected API error")
	}
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("unexpected status code: %d", apiErr.StatusCode)
	}
	if apiErr.RequestID != "req-audio-1" {
		t.Fatalf("unexpected request id: %q", apiErr.RequestID)
	}
	if !strings.Contains(err.Error(), "invalid audio") {
		t.Fatalf("unexpected error text: %v", err)
	}
}

func TestClientMockModeShortCircuitsRequests(t *testing.T) {
	t.Setenv("QINIU_MOCK_CURL", "1")

	client := NewClient("mock-token",
		WithLogger(newSilentLogger()),
		WithDebugLog(false),
	)

	_, err := client.Get(context.Background(), "/voice/list")
	if err == nil || !strings.Contains(err.Error(), "mock mode enabled") {
		t.Fatalf("expected mock mode error, got %v", err)
	}
}
