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

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/goplus/xai"
)

var (
	_ xai.Provider = (*Provider)(nil)
)

// -----------------------------------------------------------------------------

type Provider struct {
	cli anthropic.Client
}

func (p *Provider) Chat(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) (xai.Message, error) {
	resp, err := p.cli.Messages.New(ctx, buildParams(params), buildOptions(opts)...)
	if err != nil {
		return nil, err // TODO(xsw): translate error
	}
	return resp, nil // TODO(xsw): translate msg
}

func (p *Provider) ChatStreaming(ctx context.Context, params xai.ParamBuilder, opts xai.OptionBuilder) xai.StreamMessage {
	resp := p.cli.Messages.NewStreaming(ctx, buildParams(params), buildOptions(opts)...)
	return resp // TODO(xsw): translate msg
}

// -----------------------------------------------------------------------------
