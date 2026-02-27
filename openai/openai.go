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

	"github.com/goplus/xai"
	"github.com/openai/openai-go/v3/responses"
)

var (
	_ xai.Provider = (*Provider)(nil)
)

// -----------------------------------------------------------------------------

type Provider struct {
	responses responses.ResponseService
}

func (p *Provider) Gen(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.Message, error) {
	resp, err := p.responses.New(ctx, buildParams(params), buildOptions(opts)...)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return message{resp}, nil
}

func (p *Provider) GenStream(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) iter.Seq2[xai.Message, error] {
	resp := p.responses.NewStreaming(ctx, buildParams(params), buildOptions(opts)...)
	return buildMsgIter(resp)
}

// -----------------------------------------------------------------------------
