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
	xai "github.com/goplus/xai/spec"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

type params struct {
	model    string
	contents []*genai.Content
	config   genai.GenerateContentConfig
}

func (p *params) System(v xai.TextBuilder) xai.ParamBuilder {
	p.config.SystemInstruction = buildTexts(v)
	return p
}

func (p *params) Messages(msgs ...xai.MsgBuilder) xai.ParamBuilder {
	p.contents = buildMessages(msgs)
	return p
}

func (p *params) Tools(tools ...xai.ToolBase) xai.ParamBuilder {
	p.config.Tools = buildTools(tools)
	return p
}

func (p *params) Model(model xai.Model) xai.ParamBuilder {
	p.model = string(model) // TODO(xsw): validate model
	return p
}

func (p *params) MaxOutputTokens(v int64) xai.ParamBuilder {
	p.config.MaxOutputTokens = int32(v)
	return p
}

func (p *params) Compact(maxInputTokens int64) xai.ParamBuilder {
	// gemini does not support compaction, so we just ignore this parameter for now.
	return p
}

func (p *params) Container(v string) xai.ParamBuilder {
	// TODO(xsw): validate container
	return p
}

func (p *params) InferenceGeo(v string) xai.ParamBuilder {
	// TODO(xsw): validate inference geo
	return p
}

func (p *params) Temperature(v float64) xai.ParamBuilder {
	p.config.Temperature = genai.Ptr(float32(v))
	return p
}

func (p *params) TopK(v int64) xai.ParamBuilder {
	p.config.TopK = genai.Ptr(float32(v)) // TODO(xsw): validate top_k
	return p
}

func (p *params) TopP(v float64) xai.ParamBuilder {
	p.config.TopP = genai.Ptr(float32(v)) // TODO(xsw): validate top_p
	return p
}

func (p *Service) Params() xai.ParamBuilder {
	return &params{}
}

func buildParams(in xai.ParamBuilder) (string, []*genai.Content, *genai.GenerateContentConfig) {
	p := in.(*params)
	return p.model, p.contents, &p.config
}

// -----------------------------------------------------------------------------
