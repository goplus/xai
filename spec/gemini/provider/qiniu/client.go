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
	"strings"
	"time"
)

const (
	DefaultHTTPTimeout = 180 * time.Second

	DefaultMaxRetries     = 3
	DefaultBaseRetryDelay = 1 * time.Second
)

// HTTPDoer is the minimal interface required by this provider for HTTP calls.
type HTTPDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type client struct {
	token          string
	base           string
	http           HTTPDoer
	maxRetries     int
	baseRetryDelay time.Duration
	debugLog       bool
	logger         *log.Logger
}

type clientOptions struct {
	maxRetries     int
	baseRetryDelay time.Duration
	debugLog       bool
	logger         *log.Logger
}

func newClient(token, baseURL string, doer HTTPDoer, opts *clientOptions) *client {
	if doer == nil {
		doer = &http.Client{Timeout: DefaultHTTPTimeout}
	}
	c := &client{
		token: token,
		base:  normalizeBaseURL(baseURL),
		http:  doer,
	}
	if opts != nil {
		c.maxRetries = opts.maxRetries
		c.baseRetryDelay = opts.baseRetryDelay
		c.debugLog = opts.debugLog
		c.logger = opts.logger
	}
	if c.logger == nil {
		c.logger = log.Default()
	}
	return c
}

func (c *client) postJSON(ctx context.Context, path string, payload any, out any) error {
	return c.postJSONAt(ctx, c.base, path, payload, out)
}

func (c *client) postJSONAt(ctx context.Context, baseURL, path string, payload any, out any) error {
	data, err := c.doWithRetry(ctx, http.MethodPost, baseURL, path, payload)
	if err != nil {
		return err
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(data, out)
}

func (c *client) postStreamAt(ctx context.Context, baseURL, path string, payload any) (io.ReadCloser, error) {
	resp, err := c.doStreamWithRetry(ctx, baseURL, path, payload)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

// isRetryableStatus returns true if the HTTP status code indicates a retryable error.
func isRetryableStatus(statusCode int) bool {
	return statusCode == 429 || statusCode == 500 || statusCode == 502 || statusCode == 503 || statusCode == 504
}

func (c *client) logDebug(format string, args ...any) {
	if c.debugLog && c.logger != nil {
		c.logger.Printf("[qiniu] "+format, args...)
	}
}

// buildCurlCommand builds an equivalent curl command for debugging.
func (c *client) buildCurlCommand(method, url string, body []byte) string {
	var cmd bytes.Buffer
	cmd.WriteString("curl -X ")
	cmd.WriteString(method)
	cmd.WriteString(fmt.Sprintf(" -H 'Authorization: Bearer %s'", c.token))
	cmd.WriteString(" -H 'Content-Type: application/json'")
	cmd.WriteString(" -H 'Accept: application/json'")
	if len(body) > 0 {
		cmd.WriteString(fmt.Sprintf(" -d '%s'", string(body)))
	}
	cmd.WriteString(fmt.Sprintf(" '%s'", url))
	return cmd.String()
}

func (c *client) doWithRetry(ctx context.Context, method, baseURL, path string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")
	curlCmd := c.buildCurlCommand(method, url, body)
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

		resp, err := c.do(ctx, method, url, body)
		if err != nil {
			lastErr = fmt.Errorf("qiniu: request failed: %w", err)
			c.logDebug("Network error (attempt %d/%d): %v", attempt+1, maxAttempts, err)
			continue
		}

		data, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("qiniu: failed to read response body: %w", err)
			continue
		}

		requestID := resp.Header.Get("X-Request-Id")
		c.logDebug("Response status: %d, request_id: %s", resp.StatusCode, requestID)
		if c.debugLog && len(data) > 0 {
			c.logDebug("Response body: %s", string(data))
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return data, nil
		}

		apiErr := c.parseAPIError(data, resp.StatusCode, requestID)
		if isRetryableStatus(resp.StatusCode) && attempt < c.maxRetries {
			lastErr = apiErr
			c.logDebug("Retryable status %d (attempt %d/%d)", resp.StatusCode, attempt+1, maxAttempts)
			continue
		}
		return nil, apiErr
	}
	return nil, lastErr
}

func (c *client) doStreamWithRetry(ctx context.Context, baseURL, path string, payload any) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	url := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")
	curlCmd := c.buildCurlCommand(http.MethodPost, url, body)
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

		resp, err := c.do(ctx, http.MethodPost, url, body)
		if err != nil {
			lastErr = fmt.Errorf("qiniu: request failed: %w", err)
			c.logDebug("Network error (attempt %d/%d): %v", attempt+1, maxAttempts, err)
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return resp, nil
		}

		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		requestID := resp.Header.Get("X-Request-Id")
		apiErr := c.parseAPIError(data, resp.StatusCode, requestID)
		if isRetryableStatus(resp.StatusCode) && attempt < c.maxRetries {
			lastErr = apiErr
			c.logDebug("Retryable status %d (attempt %d/%d)", resp.StatusCode, attempt+1, maxAttempts)
			continue
		}
		return nil, apiErr
	}
	return nil, lastErr
}

func (c *client) parseAPIError(respBody []byte, statusCode int, requestID string) *APIError {
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

func (c *client) do(ctx context.Context, method, url string, body []byte) (*http.Response, error) {
	var reqBody io.Reader
	if len(body) > 0 {
		reqBody = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return c.http.Do(req)
}

