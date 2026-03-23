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

package geno

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

// -----------------------------------------------------------------------------

// Client is a client for making HTTP requests to the Geno API. It provides methods
// for configuring the base URL, timeout, and other settings for the HTTP requests.
type Client struct {
	client  http.Client
	baseURL *url.URL
}

// NewClient creates a new Client instance with the given http.Client. If client is nil,
// it will use http.DefaultClient by default.
//
// We encourage users create a authenticated http.Client and pass it to NewClient, so that
// the authentication logic is decoupled from the client implementation. For example, users
// can use the oauth2 package to create an authenticated http.Client like this:
//
//	import "golang.org/x/oauth2"
//
//	ctx := context.Background()
//	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "your-token"})
//	tc := oauth2.NewClient(ctx, ts)
//	client := geno.NewClient(tc)
//
// This way, the client will automatically include the access token in the Authorization
// header of each request.
func NewClient(client *http.Client) *Client {
	if client == nil {
		client = http.DefaultClient
	}
	return &Client{client: *client}
}

// BaseURL sets the base URL for the client.
func (p *Client) BaseURL(baseURL string) *Client {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	p.baseURL = u
	return p
}

// Timeout sets the timeout for each request attempt. This should be smaller than the
// timeout defined in the context, which spans all retries.
func (p *Client) Timeout(timeout time.Duration) *Client {
	p.client.Timeout = timeout
	return p
}

// NewRequest creates a new Request with the given method and path. The path will
// be resolved against the client's base URL if it is set.
func (p *Client) NewRequest(method, path string) (*Request, error) {
	req, err := http.NewRequestWithContext(context.Background(), method, path, nil)
	if err != nil {
		return nil, err
	}
	return &Request{Request: *req, c: p}, nil
}

// -----------------------------------------------------------------------------

// HTTPOptions represents options for configuring an HTTP request, such as the
// base URL and timeout. These options can be set on a Request to override the
// client's settings.
type HTTPOptions struct {
	baseURL *url.URL
	timeout *time.Duration
}

// BaseURL sets the base URL for the request. It will override the client's base
// URL if set.
func (p *HTTPOptions) BaseURL(baseURL string) *HTTPOptions {
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}
	p.baseURL = u
	return p
}

// Timeout sets the timeout for the request. It will override the client's timeout
// if set.
func (p *HTTPOptions) Timeout(timeout time.Duration) *HTTPOptions {
	p.timeout = &timeout
	return p
}

// -----------------------------------------------------------------------------

// Request represents an HTTP request to the Geno API. It provides methods for
// configuring the request, such as setting the base URL and host header.
type Request struct {
	http.Request
	c *Client
}

// Json sets the request body to the JSON encoding of v and sets the Content-Type
// header to application/json.
func (p *Request) Json(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	p.Header.Set("Content-Type", "application/json")
	p.Request.ContentLength = int64(len(b))
	p.Request.Body = io.NopCloser(bytes.NewReader(b))
	p.Request.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(b)), nil
	}
	return nil
}

// Do sends the HTTP request and returns the response.
// options can be nil, in which case the client's settings will be used.
func (p *Request) Do(ctx context.Context, options *HTTPOptions) (*http.Response, error) {
	var baseURL *url.URL
	var timeout *time.Duration
	if options != nil {
		baseURL = options.baseURL
		timeout = options.timeout
	}
	req := p.Request.WithContext(ctx)
	client := &p.c.client
	if timeout != nil {
		// don't modify the client's timeout directly, as it may be shared by
		// multiple requests. Instead, create a copy of the client with the new
		// timeout.
		clientCopy := *client
		client = &clientCopy
		client.Timeout = *timeout
	}
	if baseURL == nil {
		baseURL = p.c.baseURL
	}
	if baseURL != nil {
		req.URL = baseURL.ResolveReference(req.URL)
		if req.Host == "" {
			req.Host = req.URL.Host
		}
	}
	return client.Do(req)
}

// -----------------------------------------------------------------------------
