/*
 * Copyright (c) 2026 The XGo Authors (xgo.dev). All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package qiniu

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	DefaultBaseURL     = "https://api.qnaigc.com"
	OverseasBaseURL    = "https://openai.sufy.com"
	DefaultHTTPTimeout = 60 * time.Second

	DefaultMaxRetries     = 3
	DefaultBaseRetryDelay = 1 * time.Second
)

// Client is the HTTP client for Qiniu Kling API.
type Client struct {
	httpClient     *http.Client
	baseURL        string
	token          string
	maxRetries     int
	baseRetryDelay time.Duration
	debugLog       bool
	logger         *log.Logger
}

// ClientOption configures the Client.
type ClientOption func(*Client)

// WithBaseURL sets the base URL for API requests.
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithRetry enables retry with exponential backoff for retryable errors.
// maxRetries is the maximum number of retry attempts (0 means no retry).
// baseDelay is the initial delay between retries (doubles with each attempt).
func WithRetry(maxRetries int, baseDelay time.Duration) ClientOption {
	return func(c *Client) {
		c.maxRetries = maxRetries
		c.baseRetryDelay = baseDelay
	}
}

// WithDebugLog enables debug logging for HTTP requests and responses.
func WithDebugLog(enabled bool) ClientOption {
	return func(c *Client) {
		c.debugLog = enabled
	}
}

// WithLogger sets a custom logger for debug output.
func WithLogger(logger *log.Logger) ClientOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// NewClient creates a new Qiniu API client.
func NewClient(token string, opts ...ClientOption) *Client {
	c := &Client{
		httpClient:     &http.Client{Timeout: DefaultHTTPTimeout},
		baseURL:        DefaultBaseURL,
		token:          token,
		maxRetries:     0,
		baseRetryDelay: DefaultBaseRetryDelay,
		debugLog:       true,
		logger:         log.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Token returns the current token.
func (c *Client) Token() string {
	return c.token
}

// BaseURL returns the current base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Post sends a POST request with JSON body and returns the response body.
func (c *Client) Post(ctx context.Context, endpoint string, body any) ([]byte, error) {
	return c.doRequest(ctx, http.MethodPost, endpoint, body)
}

// Get sends a GET request and returns the response body.
func (c *Client) Get(ctx context.Context, endpoint string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, endpoint, nil)
}

// isRetryableStatus returns true if the HTTP status code indicates a retryable error.
func isRetryableStatus(statusCode int) bool {
	return statusCode == 429 || statusCode == 500 || statusCode == 502 || statusCode == 503 || statusCode == 504
}

func (c *Client) logDebug(format string, args ...any) {
	if c.debugLog && c.logger != nil {
		c.logger.Printf("[qiniu] "+format, args...)
	}
}

// buildCurlCommand builds an equivalent curl command for debugging.
func (c *Client) buildCurlCommand(method, url string, body []byte) string {
	var cmd bytes.Buffer
	cmd.WriteString("curl -X ")
	cmd.WriteString(method)

	cmd.WriteString(fmt.Sprintf(" -H 'Authorization: Bearer %s'", c.token))
	cmd.WriteString(" -H 'Content-Type: application/json'")

	if len(body) > 0 {
		cmd.WriteString(fmt.Sprintf(" -d '%s'", string(body)))
	}

	cmd.WriteString(fmt.Sprintf(" '%s'", url))
	return cmd.String()
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, body any) ([]byte, error) {
	url := c.baseURL + endpoint

	var reqBodyBytes []byte
	if body != nil {
		var err error
		reqBodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("qiniu: failed to marshal request body: %w", err)
		}
	}

	curlCmd := c.buildCurlCommand(method, url, reqBodyBytes)
	if c.logger != nil {
		c.logger.Printf("[qiniu] curl command:\n%s", curlCmd)
	} else {
		fmt.Println(curlCmd)
	}

	if os.Getenv("QINIU_MOCK_CURL") != "" {
		c.logDebug("Mock mode enabled, returning mock response")
		return nil, errors.New("mock mode enabled")
	}

	var lastErr error
	maxAttempts := c.maxRetries + 1

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			delay := c.baseRetryDelay * time.Duration(1<<uint(attempt-1))
			c.logDebug("Retry %d/%d after %v", attempt, c.maxRetries, delay)

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		var reqBody io.Reader
		if len(reqBodyBytes) > 0 {
			reqBody = bytes.NewReader(reqBodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("qiniu: failed to create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("qiniu: request failed: %w", err)
			c.logDebug("Network error (attempt %d/%d): %v", attempt+1, maxAttempts, err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("qiniu: failed to read response body: %w", err)
			continue
		}

		requestID := resp.Header.Get("X-Request-Id")
		c.logDebug("Response status: %d, request_id: %s", resp.StatusCode, requestID)
		if c.debugLog && len(respBody) > 0 {
			c.logDebug("Response body: %s", string(respBody))
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return respBody, nil
		}

		apiErr := c.parseAPIError(respBody, resp.StatusCode, requestID)

		if isRetryableStatus(resp.StatusCode) && attempt < c.maxRetries {
			lastErr = apiErr
			c.logDebug("Retryable status %d (attempt %d/%d)", resp.StatusCode, attempt+1, maxAttempts)
			continue
		}

		return nil, apiErr
	}

	return nil, lastErr
}

func (c *Client) parseAPIError(respBody []byte, statusCode int, requestID string) *APIError {
	var apiErr APIError
	if json.Unmarshal(respBody, &apiErr) == nil && apiErr.HasError() {
		apiErr.StatusCode = statusCode
		if apiErr.RequestID == "" {
			apiErr.RequestID = requestID
		}
		return &apiErr
	}
	return &APIError{
		Message:    fmt.Sprintf("API error (status %d): %s", statusCode, string(respBody)),
		StatusCode: statusCode,
		RequestID:  requestID,
	}
}

// APIError represents an error response from the Qiniu API.
type APIError struct {
	Code       string `json:"code,omitempty"`
	Message    string `json:"message,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	StatusCode int    `json:"-"`
	Error_     *struct {
		Code    string `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
		Type    string `json:"type,omitempty"`
	} `json:"error,omitempty"`
}

// HasError returns true if the response indicates an error.
func (e *APIError) HasError() bool {
	return e.Message != "" || e.Code != "" || e.Error_ != nil
}

// Error implements the error interface.
func (e *APIError) Error() string {
	var msg, code string

	if e.Error_ != nil {
		if e.Error_.Message != "" {
			msg = e.Error_.Message
		}
		if e.Error_.Code != "" {
			code = e.Error_.Code
		}
	}
	if msg == "" && e.Message != "" {
		msg = e.Message
	}
	if code == "" && e.Code != "" {
		code = e.Code
	}

	if msg == "" {
		msg = "unknown error"
	}

	var result string
	if code != "" {
		result = fmt.Sprintf("qiniu: [%s] %s", code, msg)
	} else {
		result = fmt.Sprintf("qiniu: %s", msg)
	}

	if e.RequestID != "" {
		result += fmt.Sprintf(" (request_id: %s)", e.RequestID)
	}
	if e.StatusCode > 0 {
		result += fmt.Sprintf(" (status: %d)", e.StatusCode)
	}

	return result
}

// GetCode returns the error code.
func (e *APIError) GetCode() string {
	if e.Error_ != nil && e.Error_.Code != "" {
		return e.Error_.Code
	}
	return e.Code
}

// GetMessage returns the error message.
func (e *APIError) GetMessage() string {
	if e.Error_ != nil && e.Error_.Message != "" {
		return e.Error_.Message
	}
	return e.Message
}
