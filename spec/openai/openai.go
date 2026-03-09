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

package openai

import (
	"context"
	"fmt"
	"iter"
	"net/http"
	"net/url"
	"strings"
	"sync"

	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/option"
)

// -----------------------------------------------------------------------------

const defaultAPIBaseURL = "https://api.openai.com/v1/"

type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Service implements xai.Service using OpenAI APIs.
// It supports both v3 (Responses API) and v1 (Chat Completions API).
type Service struct {
	provider       provider
	tools          tools
	baseURL        string
	apiKey         string
	httpClient     httpDoer
	httpClientOnce sync.Once
}

func (p *Service) Features() xai.Feature {
	if p.provider == nil {
		return xai.FeatureOperation
	}
	return p.provider.Features() | xai.FeatureOperation
}

func (p *Service) Gen(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.GenResponse, error) {
	if p.provider == nil {
		return nil, xai.ErrNotSupported
	}
	req := buildParams(params)
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("openai: at least one message is required")
	}
	return p.provider.Gen(ctx, req, buildOptions(opts))
}

func (p *Service) GenStream(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) iter.Seq2[xai.GenResponse, error] {
	if p.provider == nil {
		return func(yield func(xai.GenResponse, error) bool) {
			yield(nil, xai.ErrNotSupported)
		}
	}
	req := buildParams(params)
	if len(req.Messages) == 0 {
		return func(yield func(xai.GenResponse, error) bool) {
			yield(nil, fmt.Errorf("openai: at least one message is required"))
		}
	}
	return func(yield func(xai.GenResponse, error) bool) {
		for resp, err := range p.provider.GenStream(ctx, req, buildOptions(opts)) {
			if !yield(resp, err) {
				return
			}
		}
	}
}

// -----------------------------------------------------------------------------

const (
	// Scheme is the URI scheme for v3 Responses API.
	Scheme = "openai"
	// SchemeV1 is the URI scheme for v1 Chat Completions API.
	SchemeV1 = "openai-v1"
)

// New creates a new Service instance using v3 Responses API.
// uri should be in the format of "openai:base=service_base_url&key=api_key".
//
// `base` is the base URL of the API endpoint.
// `key` is the API key for authentication.
// `org` is the organization ID to use for the API requests.
// `project` is the project ID to use for the API requests.
// `webhook_secret` is the secret for validating webhook requests.
//
// For example, "openai:base=https://api.openai.com/v1/&key=your_api_key".
func New(ctx context.Context, uri string) (xai.Service, error) {
	query, opts, err := parseURI(uri, Scheme)
	if err != nil {
		return nil, err
	}
	return newService(newV3Provider(opts), query), nil
}

// NewV1 creates a new Service instance using v1 Chat Completions API.
// uri should be in the format of "openai-v1:base=service_base_url&key=api_key".
// When base and key are provided, uses extended response parsing to support
// APIs that return an "images" field (e.g. Qiniu gemini-2.5-flash-image).
func NewV1(ctx context.Context, uri string) (xai.Service, error) {
	query, opts, err := parseURI(uri, SchemeV1)
	if err != nil {
		return nil, err
	}
	base := normalizeAPIBaseURL(queryFirst(query, "base"))
	key := queryFirst(query, "key")
	return newService(newV1Provider(opts, base, key), query), nil
}

// NewVideoOnly creates a Service that only supports video operations (e.g. Sora).
// Gen and GenStream return xai.ErrNotSupported.
// uri should be in the format of "openai-v1:base=service_base_url&key=api_key".
func NewVideoOnly(ctx context.Context, uri string) (xai.Service, error) {
	query, _, err := parseURI(uri, SchemeV1)
	if err != nil {
		return nil, err
	}
	return newService(newVideoOnlyProvider(), query), nil
}

func parseURI(uri, scheme string) (url.Values, []option.RequestOption, error) {
	query, err := url.ParseQuery(strings.TrimPrefix(uri, scheme+":"))
	if err != nil {
		return nil, nil, err
	}
	opts := []option.RequestOption{option.WithEnvironmentProduction()}
	if base := query["base"]; len(base) > 0 {
		opts = append(opts, option.WithBaseURL(base[0]))
	}
	if key := query["key"]; len(key) > 0 {
		opts = append(opts, option.WithAPIKey(key[0]))
	}
	if org := query["org"]; len(org) > 0 {
		opts = append(opts, option.WithOrganization(org[0]))
	}
	if proj := query["project"]; len(proj) > 0 {
		opts = append(opts, option.WithProject(proj[0]))
	}
	if webhookSec := query["webhook_secret"]; len(webhookSec) > 0 {
		opts = append(opts, option.WithWebhookSecret(webhookSec[0]))
	}
	return query, opts, nil
}

func newService(p provider, query url.Values) *Service {
	return &Service{
		provider: p,
		tools:    make(tools),
		baseURL:  normalizeAPIBaseURL(queryFirst(query, "base")),
		apiKey:   queryFirst(query, "key"),
	}
}

func queryFirst(query url.Values, key string) string {
	if v := query.Get(key); v != "" {
		return v
	}
	return ""
}

func normalizeAPIBaseURL(base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = defaultAPIBaseURL
	}
	return strings.TrimSuffix(base, "/") + "/"
}

func init() {
	xai.Register(Scheme, New)
	xai.Register(SchemeV1, NewV1)
}

// -----------------------------------------------------------------------------
