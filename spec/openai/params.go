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
)

// -----------------------------------------------------------------------------

type params struct {
	req  genRequest
	msgs []xai.MsgBuilder
}

func (p *params) System(v xai.TextBuilder) xai.ParamBuilder {
	p.req.System = buildTexts(v)
	return p
}

func (p *params) Messages(msgs ...xai.MsgBuilder) xai.ParamBuilder {
	p.msgs = msgs
	return p
}

func (p *params) Tools(tools ...xai.ToolBase) xai.ParamBuilder {
	p.req.Tools = make([]*toolDef, len(tools))
	for i, v := range tools {
		td := &toolDef{}
		v.UnderlyingAssignTo(td)
		p.req.Tools[i] = td
	}
	return p
}

func (p *params) Model(model xai.Model) xai.ParamBuilder {
	p.req.Model = string(model)
	return p
}

func (p *params) MaxOutputTokens(v int64) xai.ParamBuilder {
	p.req.MaxOutputTokens = v
	return p
}

func (p *params) Compact(maxInputTokens int64) xai.ParamBuilder {
	return p
}

func (p *params) Container(v string) xai.ParamBuilder {
	return p
}

func (p *params) InferenceGeo(v string) xai.ParamBuilder {
	return p
}

func (p *params) Temperature(v float64) xai.ParamBuilder {
	p.req.Temperature = v
	p.req.HasTemperature = true
	return p
}

func (p *params) TopK(v int64) xai.ParamBuilder {
	return p
}

func (p *params) TopP(v float64) xai.ParamBuilder {
	p.req.TopP = v
	p.req.HasTopP = true
	return p
}

func (p *Service) Params() xai.ParamBuilder {
	return &params{}
}

func buildParams(in xai.ParamBuilder) *genRequest {
	p := in.(*params)
	// Build messages
	p.req.Messages = make([]*message, len(p.msgs))
	for i, mb := range p.msgs {
		if ext, ok := mb.(msgBuilderExt); ok {
			p.req.Messages[i] = ext.msgBuilder.getMessage()
		} else {
			p.req.Messages[i] = mb.(*msgBuilder).getMessage()
		}
	}
	return &p.req
}

// -----------------------------------------------------------------------------
