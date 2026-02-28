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

package gemini

import (
	"context"
	"iter"
	"net/url"
	"strings"

	"github.com/goplus/xai"
	"google.golang.org/genai"
)

var (
	_ xai.Provider = (*Provider)(nil)
)

// -----------------------------------------------------------------------------

type Provider struct {
	models genai.Models
}

func (p *Provider) Gen(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.Message, error) {
	model, contents, config := buildParams(params)
	buildOptions(config, opts)
	resp, err := p.models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return message{resp}, nil
}

func (p *Provider) GenStream(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) iter.Seq2[xai.Message, error] {
	model, contents, config := buildParams(params)
	buildOptions(config, opts)
	iter := p.models.GenerateContentStream(ctx, model, contents, config)
	return func(yield func(xai.Message, error) bool) {
		iter(func(resp *genai.GenerateContentResponse, err error) bool {
			return yield(message{resp}, err)
		})
	}
}

// -----------------------------------------------------------------------------

const (
	Scheme = "gemini"
)

// Create creates a new Provider instance based on the scheme in the given URI.
// uri should be in the format of "gemini:?base=xxx", where "base" is the base URL
// of the API endpoint.
// For example, "gemini:?base=https://generativelanguage.googleapis.com/".
func Create(ctx context.Context, uri string) (xai.Provider, error) {
	params, err := url.ParseQuery(strings.TrimPrefix(uri, Scheme+":"))
	if err != nil {
		return nil, err
	}
	var conf genai.ClientConfig
	if base := params["base"]; len(base) > 0 {
		conf.HTTPOptions.BaseURL = base[0]
	}
	cli, err := genai.NewClient(ctx, &conf)
	if err != nil {
		return nil, err
	}
	return &Provider{
		models: *cli.Models,
	}, nil
}

func init() {
	xai.RegisterCreator(Scheme, Create)
}

// -----------------------------------------------------------------------------
