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
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/openai"
)

const (
	// DefaultBaseURL is the default Qiniu API base URL (domestic).
	DefaultBaseURL = "https://api.qnaigc.com/v1/"
	// OverseasBaseURL is the overseas API base URL.
	OverseasBaseURL = "https://openai.sufy.com/v1/"
)

// ClientOption configures the Qiniu OpenAI-compatible service.
type ClientOption func(*clientConfig)

type clientConfig struct {
	baseURL string
}

// WithBaseURL sets the base URL for API requests.
func WithBaseURL(u string) ClientOption {
	return func(c *clientConfig) {
		c.baseURL = u
	}
}

// NewService creates an OpenAI-compatible Service backed by Qiniu API (api.qnaigc.com).
// It uses provider_v1 (Chat Completions API) for Gen/GenStream and supports
// Sora video operations via Service.Operation(..., xai.GenVideo).
// The apiKey can be passed directly or read from QINIU_API_KEY env when empty.
// Panics on error; use NewServiceWithError to handle errors.
func NewService(apiKey string, opts ...ClientOption) *openai.Service {
	svc, err := NewServiceWithError(apiKey, opts...)
	if err != nil {
		panic("qiniu: " + err.Error())
	}
	return svc
}

// NewServiceWithError creates an OpenAI-compatible Service backed by Qiniu API.
// Returns an error if the API key is missing or service creation fails.
func NewServiceWithError(apiKey string, opts ...ClientOption) (*openai.Service, error) {
	if apiKey == "" {
		apiKey = os.Getenv("QINIU_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required (set QINIU_API_KEY or pass apiKey)")
	}
	cfg := &clientConfig{baseURL: DefaultBaseURL}
	for _, opt := range opts {
		opt(cfg)
	}
	uri := openai.SchemeV1 + ":base=" + cfg.baseURL + "&key=" + url.QueryEscape(apiKey)
	svc, err := openai.NewV1(context.Background(), uri)
	if err != nil {
		return nil, err
	}
	return svc.(*openai.Service), nil
}

// Register registers the Qiniu-backed OpenAI service with xai under scheme "qiniu".
// After calling Register(apiKey), xai.New(ctx, "qiniu:") returns the Qiniu service.
// apiKey can be empty to use QINIU_API_KEY from environment.
func Register(apiKey string, opts ...ClientOption) {
	svc, err := NewServiceWithError(apiKey, opts...)
	if err != nil {
		panic("qiniu: " + err.Error())
	}
	xai.Register("qiniu", func(ctx context.Context, uri string) (xai.Service, error) {
		// Allow override via uri: qiniu:key=xxx or qiniu:base=xxx&key=xxx
		query := strings.TrimPrefix(uri, "qiniu:")
		if query == "" {
			return svc, nil
		}
		params, err := url.ParseQuery(query)
		if err != nil {
			return nil, err
		}
		key := apiKey
		if k := params["key"]; len(k) > 0 {
			key = k[0]
		}
		if key == "" {
			key = os.Getenv("QINIU_API_KEY")
		}
		base := DefaultBaseURL
		if b := params["base"]; len(b) > 0 {
			base = strings.TrimSuffix(b[0], "/") + "/"
		}
		opts := []ClientOption{}
		if base != DefaultBaseURL {
			opts = append(opts, WithBaseURL(base))
		}
		return NewServiceWithError(key, opts...)
	})
}
