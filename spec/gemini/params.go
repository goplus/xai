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
	"time"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/util"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

type genParams struct {
	model    string
	contents []*genai.Content
	config   genai.GenerateContentConfig
	pconfig  *util.Params[adapter]
	ctx      context.Context
}

/*
// TopK float
// Optional. For each token selection step, the “top_k“ tokens with the
// highest probabilities are sampled. Then tokens are further filtered based
// on “top_p“ with the final token selected using temperature sampling. Use
// a lower number for less random responses and a higher number for more
// random responses.

// CandidateCount int
// Optional. Number of response variations to return.
// If empty, the system will choose a default value (currently 1).

// StopSequences []string
// Optional. List of strings that tells the model to stop generating text if one
// of the strings is encountered in the response.

// ResponseLogprobs bool
// Optional. Whether to return the log probabilities of the tokens that were
// chosen by the model at each step.

// Logprobs int
// Optional. Number of top candidate tokens to return the log probabilities for
// at each generation step.

// PresencePenalty float
// Optional. Positive values penalize tokens that already appear in the
// generated text, increasing the probability of generating more diverse
// content.

// FrequencyPenalty float
// Optional. Positive values penalize tokens that repeatedly appear in the
// generated text, increasing the probability of generating more diverse
// content.

// Seed int
// Optional. When “seed“ is fixed to a specific number, the model makes a best
// effort to provide the same response for repeated requests. By default, a
// random number is used.

// RoutingConfig *GenerationConfigRoutingConfig
// Optional. Configuration for model router requests.

// ModelSelectionConfig *ModelSelectionConfig
// Optional. Configuration for model selection.

// SafetySettings []*SafetySetting
// Optional. Safety settings in the request to block unsafe content in the
// response.

// ToolConfig *ToolConfig
// Optional. Associates model output to a specific function call.

// Labels map[string]string
// Optional. Labels with user-defined metadata to break down billed charges.

// CachedContent string
// Optional. Resource name of a context cache that can be used in subsequent
// requests.

// ResponseModalities []string
// Optional. The requested modalities of the response. Represents the set of
// modalities that the model can return.

// MediaResolution MediaResolution
// Optional. If specified, the media resolution specified will be used.

// SpeechConfig *SpeechConfig
// Optional. The speech generation configuration.

// AudioTimestamp bool
// Optional. If enabled, audio timestamp will be included in the request to the
// model.

// ThinkingConfig *ThinkingConfig
// Optional. The thinking features configuration.

// ImageConfig *ImageConfig
// Optional. The image generation configuration.

// EnableEnhancedCivicAnswers bool
// Optional. Enables enhanced civic answers. It may not be available for all
// models. This field is not supported in Vertex AI.

// ModelArmorConfig *ModelArmorConfig
// Optional. Settings for prompt and response sanitization using the Model Armor
// service. If supplied, safety_settings must not be supplied.
*/
func (p *genParams) Set(name string, val any) xai.GenParams {
	if p.pconfig == nil {
		p.pconfig = util.NewParams[adapter](&p.config)
	}
	p.pconfig.Set(name, val)
	return p
}

func (p *genParams) System(texts ...string) xai.GenParams {
	parts := make([]*genai.Part, len(texts))
	for i, text := range texts {
		parts[i] = genai.NewPartFromText(text)
	}
	p.config.SystemInstruction = &genai.Content{
		Parts: parts,
	}
	return p
}

func (p *genParams) Messages(msgs ...xai.MsgBuilder) xai.GenParams {
	p.contents = buildMessages(msgs)
	return p
}

func (p *genParams) Tools(tools ...xai.ToolBase) xai.GenParams {
	p.config.Tools = buildTools(tools)
	return p
}

func (p *genParams) Model(model xai.Model) xai.GenParams {
	p.model = string(model) // TODO(xsw): validate model
	return p
}

func (p *genParams) MaxOutputTokens(v int64) xai.GenParams {
	p.config.MaxOutputTokens = int32(v)
	return p
}

func (p *genParams) Compact(maxInputTokens int64) xai.GenParams {
	// gemini does not support compaction, so we just ignore this parameter for now.
	return p
}

func (p *genParams) Temperature(v float64) xai.GenParams {
	p.config.Temperature = genai.Ptr(float32(v))
	return p
}

func (p *genParams) TopP(v float64) xai.GenParams {
	p.config.TopP = genai.Ptr(float32(v)) // TODO(xsw): validate top_p
	return p
}

func (p *genParams) BaseURL(base string) xai.GenParams {
	if p.config.HTTPOptions == nil {
		p.config.HTTPOptions = &genai.HTTPOptions{}
	}
	p.config.HTTPOptions.BaseURL = base
	return p
}

func (p *genParams) Timeout(timeout time.Duration) xai.GenParams {
	if p.config.HTTPOptions == nil {
		p.config.HTTPOptions = &genai.HTTPOptions{}
	}
	p.config.HTTPOptions.Timeout = &timeout
	return p
}

func (p *genParams) Ctx(ctx context.Context) xai.GenParams {
	p.ctx = ctx
	return p
}

func (p *Service) GenParams() xai.GenParams {
	return &genParams{}
}

func buildGenParams(in xai.GenParams) (context.Context, string, []*genai.Content, *genai.GenerateContentConfig) {
	p := in.(*genParams)
	ctx := p.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return ctx, p.model, p.contents, &p.config
}

// -----------------------------------------------------------------------------
