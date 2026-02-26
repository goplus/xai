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
	resp, err := p.models.GenerateContent(ctx, model, contents, config)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return message{resp}, nil
}

func (p *Provider) GenStreaming(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) xai.StreamMessage {
	model, contents, config := buildParams(params)
	resp := p.models.GenerateContentStream(ctx, model, contents, config)
	return resp // TODO(xsw): translate msg
}

// -----------------------------------------------------------------------------
