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

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/goplus/xai"
)

var (
	_ xai.Provider = (*Provider)(nil)
)

// -----------------------------------------------------------------------------

type Provider struct {
	messages anthropic.BetaMessageService
}

func (p *Provider) Gen(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.Message, error) {
	resp, err := p.messages.New(ctx, buildParams(params), buildOptions(opts)...)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return message{resp}, nil
}

func (p *Provider) GenStream(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) iter.Seq2[xai.Message, error] {
	resp := p.messages.NewStreaming(ctx, buildParams(params), buildOptions(opts)...)
	return buildMsgIter(resp)
}

// -----------------------------------------------------------------------------

const (
	Scheme = "claude"
)

// Create creates a new Provider instance based on the scheme in the given URI.
// uri should be in the format of "claude:base=xxx", where "base" is the base URL
// of the API endpoint.
// For example, "claude:base=https://api.anthropic.com".
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
		messages: anthropic.NewBetaMessageService(opts...),
	}, nil
}

func init() {
	xai.RegisterCreator(Scheme, Create)
}

// -----------------------------------------------------------------------------
