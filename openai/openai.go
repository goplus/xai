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

	"github.com/goplus/xai"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
)

var (
	_ xai.Provider = (*Provider)(nil)
)

// -----------------------------------------------------------------------------

type Provider struct {
	responses responses.ResponseService
}

func (p *Provider) Gen(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.GenResponse, error) {
	resp, err := p.responses.New(ctx, buildParams(params), buildOptions(opts)...)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return response{resp}, nil
}

func (p *Provider) GenStream(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) iter.Seq2[xai.GenResponse, error] {
	resp := p.responses.NewStreaming(ctx, buildParams(params), buildOptions(opts)...)
	return buildRespIter(resp)
}

// -----------------------------------------------------------------------------

const (
	Scheme = "openai"
)

// Create creates a new Provider instance based on the scheme in the given URI.
// uri should be in the format of "openai:base=xxx", where "base" is the base URL
// of the API endpoint.
// For example, "openai:base=https://api.openai.com".
func Create(ctx context.Context, uri string) (xai.Provider, error) {
	params, err := url.ParseQuery(strings.TrimPrefix(uri, Scheme+":"))
	if err != nil {
		return nil, err
	}
	var opts []option.RequestOption
	if base := params["base"]; len(base) > 0 {
		opts = append(opts, option.WithBaseURL(base[0]))
	}
	return &Provider{
		responses: responses.NewResponseService(opts...),
	}, nil
}

func init() {
	xai.RegisterCreator(Scheme, Create)
}

// -----------------------------------------------------------------------------
