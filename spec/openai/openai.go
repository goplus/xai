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
	"iter"
	"net/url"
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/option"
)

// -----------------------------------------------------------------------------

// Service implements xai.Service using OpenAI APIs.
// It supports both v3 (Responses API) and v1 (Chat Completions API).
type Service struct {
	provider provider
	tools    tools
}

func (p *Service) Features() xai.Feature {
	return p.provider.Features()
}

func (p *Service) Gen(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.GenResponse, error) {
	req := buildParams(params)
	return p.provider.Gen(ctx, req, buildOptions(opts))
}

func (p *Service) GenStream(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) iter.Seq2[xai.GenResponse, error] {
	req := buildParams(params)
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
	_ = query // reserved for future use
	return &Service{
		provider: newV3Provider(opts),
		tools:    make(tools),
	}, nil
}

// NewV1 creates a new Service instance using v1 Chat Completions API.
// uri should be in the format of "openai-v1:base=service_base_url&key=api_key".
func NewV1(ctx context.Context, uri string) (xai.Service, error) {
	query, opts, err := parseURI(uri, SchemeV1)
	if err != nil {
		return nil, err
	}
	_ = query // reserved for future use
	return &Service{
		provider: newV1Provider(opts),
		tools:    make(tools),
	}, nil
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

func init() {
	xai.Register(Scheme, New)
	xai.Register(SchemeV1, NewV1)
}

// -----------------------------------------------------------------------------
