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
	"time"

	"github.com/goplus/xai"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

// -----------------------------------------------------------------------------

type Service struct {
	responses responses.ResponseService
	tools     tools
}

func (p *Service) Features() xai.Feature {
	return xai.FeatureGen | xai.FeatureGenStream | xai.FeatureOperation
}

func (p *Service) Gen(ctx context.Context, gp xai.GenParams) (xai.GenResponse, error) {
	params, opts := buildParams(gp)
	resp, err := p.responses.New(ctx, params, opts...)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return response{resp}, nil
}

func (p *Service) GenStream(ctx context.Context, gp xai.GenParams) iter.Seq2[xai.GenResponse, error] {
	params, opts := buildParams(gp)
	resp := p.responses.NewStreaming(ctx, params, opts...)
	return buildRespIter(resp)
}

// -----------------------------------------------------------------------------

const (
	Scheme = "openai"
)

// New creates a new Service instance based on the scheme in the given URI.
// uri should be in the format of "openai:base=service_base_url&key=api_key".
//
// `base` is the base URL of the API endpoint.
// `timeout` is the request timeout duration (e.g., "30s").
// `key` is the API key for authentication.
// `org` is the organization ID to use for the API requests.
// `project` is the project ID to use for the API requests.
// `webhook_secret` is the secret for validating webhook requests.
//
// For example, "openai:base=https://api.openai.com/v1/&key=your_api_key".
func New(ctx context.Context, uri string) (xai.Service, error) {
	params, err := url.ParseQuery(strings.TrimPrefix(uri, Scheme+":"))
	if err != nil {
		return nil, err
	}
	// Remove calls to openai.DefaultClientOptions because we don't suggest users
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
	if org := params["org"]; len(org) > 0 {
		opts = append(opts, option.WithOrganization(org[0]))
	}
	if proj := params["project"]; len(proj) > 0 {
		opts = append(opts, option.WithProject(proj[0]))
	}
	if webhookSec := params["webhook_secret"]; len(webhookSec) > 0 {
		opts = append(opts, option.WithWebhookSecret(webhookSec[0]))
	}
	return &Service{
		responses: responses.NewResponseService(opts...),
		tools:     make(tools),
	}, nil
}

func init() {
	xai.Register(Scheme, New)
}

// -----------------------------------------------------------------------------
