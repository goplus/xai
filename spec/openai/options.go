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
	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/option"
)

// -----------------------------------------------------------------------------

type options struct {
	opts []option.RequestOption
}

func (p *options) WithBaseURL(base string) xai.OptionBuilder {
	p.opts = append(p.opts, option.WithBaseURL(base))
	return p
}

func (p *Service) Options() xai.OptionBuilder {
	return &options{}
}

// WithThinking returns an OptionBuilder with thinking enabled or disabled.
// Pass svc.Options() as the first argument. Only effective for OpenAI-compatible services.
// Use thinking-enabled models like deepseek-v3.2-251201 for best results.
func WithThinking(ob xai.OptionBuilder, enabled bool) xai.OptionBuilder {
	if p, ok := ob.(*options); ok {
		return p.withThinking(enabled)
	}
	return ob
}

func (p *options) withThinking(enabled bool) *options {
	typ := "disabled"
	if enabled {
		typ = "enabled"
	}
	p.opts = append(p.opts, option.WithJSONSet("thinking", map[string]string{"type": typ}))
	return p
}

func buildOptions(opts xai.OptionBuilder) (ret []option.RequestOption) {
	if p, ok := opts.(*options); ok {
		ret = p.opts
	}
	return
}

// -----------------------------------------------------------------------------
