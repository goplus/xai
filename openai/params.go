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
	"github.com/goplus/xai"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

// -----------------------------------------------------------------------------

type params struct {
	params responses.ResponseNewParams
	sys    responses.ResponseInputMessageContentListParam
	msgs   xai.MessageBuilder
}

func (p *params) System(v xai.TextBuilder) xai.ParamBuilder {
	p.sys = buildTexts(v)
	return p
}

func (p *params) Messages(v xai.MessageBuilder) xai.ParamBuilder {
	// we will merge system prompt and messages into input param in buildParams
	// so we just store the messages here
	p.msgs = v
	return p
}

func (p *params) MaxTokens(v int64) xai.ParamBuilder {
	p.params.MaxOutputTokens = param.NewOpt(v)
	return p
}

func (p *params) Model(model xai.Model) xai.ParamBuilder {
	p.params.Model = shared.ResponsesModel(model) // TODO(xsw): validate model
	return p
}

func (p *params) Container(v string) xai.ParamBuilder {
	return p
}

func (p *params) InferenceGeo(v string) xai.ParamBuilder {
	return p
}

func (p *params) Temperature(v float64) xai.ParamBuilder {
	p.params.Temperature = param.NewOpt(v)
	return p
}

func (p *params) TopK(v int64) xai.ParamBuilder {
	// TODO(xsw): openai does not support top_k, use top_p instead
	return p
}

func (p *params) TopP(v float64) xai.ParamBuilder {
	p.params.TopP = param.NewOpt(v)
	return p
}

func (p *Provider) Params() xai.ParamBuilder {
	return &params{}
}

func buildParams(in xai.ParamBuilder) responses.ResponseNewParams {
	p := in.(*params)
	// TODO(xsw): check param values
	// Merge system prompt and messages into input param
	var sys responses.ResponseInputItemUnionParam
	if len(p.sys) > 0 {
		sys = responses.ResponseInputItemParamOfMessage(p.sys, responses.EasyInputMessageRoleSystem)
	}
	p.params.Input = buildMessages(p.msgs, sys)
	return p.params
}

// -----------------------------------------------------------------------------
