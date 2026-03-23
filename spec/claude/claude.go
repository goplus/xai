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

package claude

import (
	"context"
	"iter"
	"net/url"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/goplus/xai"
)

// -----------------------------------------------------------------------------

type Service struct {
	messages anthropic.BetaMessageService
	tools    tools
}

func (p *Service) Features() xai.Feature {
	return xai.FeatureGen | xai.FeatureGenStream
}

// -----------------------------------------------------------------------------

func (p *Service) Gen(gp xai.GenParams) (xai.GenResponse, error) {
	ctx, params, opts := buildParams(gp)
	resp, err := p.messages.New(ctx, params, opts...)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return response{resp}, nil
}

func (p *Service) GenStream(gp xai.GenParams) iter.Seq2[xai.GenResponse, error] {
	ctx, params, opts := buildParams(gp)
	resp := p.messages.NewStreaming(ctx, params, opts...)
	return buildRespIter(resp)
}

// -----------------------------------------------------------------------------

func (p *Service) Actions(model xai.Model) []xai.Action {
	// claude doesn't support any actions for now.
	return nil
}

func (p *Service) Operation(model xai.Model, action xai.Action) (op xai.Operation, err error) {
	return nil, xai.ErrNotFound
}

// -----------------------------------------------------------------------------

const (
	Scheme = "claude"
)

// New creates a new Service instance based on the scheme in the given URI.
// uri should be in the format of "claude:base=service_base_url&key=api_key".
//
// `base` is the base URL of the API endpoint.
// `timeout` is the request timeout duration (e.g., "30s").
// `key` is the API key for authentication (don't use both `key` and `token`).
// `token` is the authentication token for the API requests.
//
// For example, "claude:base=https://api.anthropic.com/&key=your_api_key".
func New(ctx context.Context, uri string) (xai.Service, error) {
	params, err := url.ParseQuery(strings.TrimPrefix(uri, Scheme+":"))
	if err != nil {
		return nil, err
	}
	// Remove calls to anthropic.DefaultClientOptions because we don't suggest users
	// to set environment variables for API key and base URL. Instead, they should
	// provide these parameters directly in the URI.
	opts := []option.RequestOption{option.WithEnvironmentProduction()}
	if base := params["base"]; len(base) > 0 {
		opts = append(opts, option.WithBaseURL(base[0]))
	}
	if timeout := params["timeout"]; len(timeout) > 0 {
		d, err := time.ParseDuration(timeout[0])
		if err != nil {
			return nil, err
		}
		opts = append(opts, option.WithRequestTimeout(d))
	}
	if key := params["key"]; len(key) > 0 {
		opts = append(opts, option.WithAPIKey(key[0]))
	}
	if token := params["token"]; len(token) > 0 {
		opts = append(opts, option.WithAuthToken(token[0]))
	}
	return &Service{
		messages: anthropic.NewBetaMessageService(opts...),
		tools:    make(tools),
	}, nil
}

func init() {
	xai.Register(Scheme, New)
}

// -----------------------------------------------------------------------------
