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
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/goplus/xai"
)

// -----------------------------------------------------------------------------

type params struct {
	params anthropic.MessageNewParams
}

func (p *params) System(v xai.TextBuilder) xai.ParamBuilder {
	p.params.System = buildTexts(v)
	return p
}

func (p *params) Messages(v xai.MessageBuilder) xai.ParamBuilder {
	p.params.Messages = buildMessages(v)
	return p
}

func (p *params) MaxTokens(v int64) xai.ParamBuilder {
	p.params.MaxTokens = v
	return p
}

func (p *params) Model(model xai.Model) xai.ParamBuilder {
	p.params.Model = anthropic.Model(model) // TODO(xsw): check model
	return p
}

func (p *params) Container(v string) xai.ParamBuilder {
	p.params.Container = param.NewOpt(v)
	return p
}

func (p *params) InferenceGeo(v string) xai.ParamBuilder {
	p.params.InferenceGeo = param.NewOpt(v)
	return p
}

func (p *params) Temperature(v float64) xai.ParamBuilder {
	p.params.Temperature = param.NewOpt(v)
	return p
}

func (p *params) TopK(v int64) xai.ParamBuilder {
	p.params.TopK = param.NewOpt(v)
	return p
}

func (p *params) TopP(v float64) xai.ParamBuilder {
	p.params.TopP = param.NewOpt(v)
	return p
}

func (p *Provider) Params() xai.ParamBuilder {
	return &params{}
}

func buildParams(in xai.ParamBuilder) anthropic.MessageNewParams {
	p := in.(*params)
	// TODO(xsw): check param values
	return p.params
}

// -----------------------------------------------------------------------------
