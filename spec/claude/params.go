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
	"reflect"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/goplus/xai"
	"github.com/goplus/xai/util"
)

// -----------------------------------------------------------------------------

type adapter struct{}

func (adapter) SetBasicOpt(fld, v reflect.Value, vkind reflect.Kind) (ok bool) {
	ok = fld.Kind() == reflect.Struct // is param.Opt
	if ok {
		setBasicOpt(fld.Addr().Interface(), v, vkind)
	}
	return
}

func setBasicOpt(fld any, v reflect.Value, vkind reflect.Kind) {
	switch fld := fld.(type) {
	case *param.Opt[string]:
		*fld = param.NewOpt(v.String())
	case *param.Opt[int64]:
		*fld = param.NewOpt(v.Int())
	case *param.Opt[float64]:
		if vkind >= reflect.Int && vkind <= reflect.Int64 {
			*fld = param.NewOpt(float64(v.Int()))
		} else {
			*fld = param.NewOpt(v.Float())
		}
	case *param.Opt[bool]:
		*fld = param.NewOpt(v.Bool())
	default:
		panic("unsupported opt type")
	}
}

func (adapter) ToUnderlying(val any) any {
	panic("todo")
}

func (adapter) FromUnderlying(v any, kind reflect.Kind) any {
	panic("todo")
}

// -----------------------------------------------------------------------------

type params struct {
	params  anthropic.BetaMessageNewParams
	pparams *util.Params[adapter]
	opts    []option.RequestOption
}

/*
// InferenceGeo string
// Specifies the geographic region for inference processing. If not specified, the
// workspace's `default_inference_geo` is used.

// TopK int
// Only sample from the top K options for each subsequent token.
//
// Used to remove "long tail" low probability responses.
// [Learn more technical details here](https://towardsdatascience.com/how-to-sample-from-language-models-682bceb97277).
//
// Recommended for advanced use cases only. You usually only need to use
// `temperature`.

// Container BetaMessageNewParamsContainerUnion
// Container identifier for reuse across requests.

// Speed BetaMessageNewParamsSpeed
// The inference speed mode for this request. `"fast"` enables high
// output-tokens-per-second inference.
//
// Any of "standard", "fast".

// CacheControl BetaCacheControlEphemeralParam
// Top-level cache control automatically applies a cache_control marker to the last
// cacheable block in the request.

// ContextManagement BetaContextManagementConfigParam
// Context management configuration.
//
// This allows you to control how Claude manages context across multiple requests,
// such as whether to clear function results or not.

// MCPServers []BetaRequestMCPServerURLDefinitionParam
// MCP servers to be utilized in this request

// Metadata BetaMetadataParam
// An object describing metadata about the request.

// OutputConfig BetaOutputConfigParam
// Configuration options for the model's output, such as the output format.

// ServiceTier BetaMessageNewParamsServiceTier
// Determines whether to use priority capacity (if available) or standard capacity
// for this request.
//
// Anthropic offers different levels of service for your API requests. See
// [service-tiers](https://docs.claude.com/en/api/service-tiers) for details.
//
// Any of "auto", "standard_only".

// StopSequences []string
// Custom text sequences that will cause the model to stop generating.
//
// Our models will normally stop when they have naturally completed their turn,
// which will result in a response `stop_reason` of `"end_turn"`.
//
// If you want the model to stop generating when it encounters custom strings of
// text, you can use the `stop_sequences` parameter. If the model encounters one of
// the custom sequences, the response `stop_reason` value will be `"stop_sequence"`
// and the response `stop_sequence` value will contain the matched stop sequence.

// Thinking BetaThinkingConfigParamUnion
// Configuration for enabling Claude's extended thinking.
//
// When enabled, responses include `thinking` content blocks showing Claude's
// thinking process before the final answer. Requires a minimum budget of 1,024
// tokens and counts towards your `max_tokens` limit.
//
// See
// [extended thinking](https://docs.claude.com/en/docs/build-with-claude/extended-thinking)
// for details.

// ToolChoice BetaToolChoiceUnionParam
// How the model should use the provided tools. The model can use a specific tool,
// any available tool, decide by itself, or not use tools at all.
*/
func (p *params) Set(name string, val any) xai.GenParams {
	if p.pparams == nil {
		p.pparams = util.NewParams[adapter](&p.params)
	}
	p.pparams.Set(name, val)
	return p
}

func (p *params) System(texts ...string) xai.GenParams {
	content := make([]anthropic.BetaTextBlockParam, len(texts))
	for i, text := range texts {
		content[i].Text = text
	}
	p.params.System = content
	return p
}

func (p *params) Messages(msgs ...xai.MsgBuilder) xai.GenParams {
	p.params.Messages = buildMessages(msgs)
	return p
}

func (p *params) Tools(tools ...xai.ToolBase) xai.GenParams {
	p.params.Tools = buildTools(tools)
	return p
}

func (p *params) Model(model xai.Model) xai.GenParams {
	p.params.Model = anthropic.Model(model) // TODO(xsw): validate model
	return p
}

func (p *params) MaxOutputTokens(v int64) xai.GenParams {
	p.params.MaxTokens = v
	return p
}

func (p *params) Compact(maxInputTokens int64) xai.GenParams {
	p.params.Betas = []anthropic.AnthropicBeta{
		"compact-2026-01-12",
	}
	p.params.ContextManagement.Edits = append(p.params.ContextManagement.Edits, anthropic.BetaContextManagementConfigEditUnionParam{
		OfCompact20260112: &anthropic.BetaCompact20260112EditParam{
			Trigger: anthropic.BetaInputTokensTriggerParam{
				Value: maxInputTokens,
			},
		},
	})
	return p
}

func (p *params) Temperature(v float64) xai.GenParams {
	if v > 1 {
		v = 1 // claude does not support temperature > 1
	}
	p.params.Temperature = param.NewOpt(v)
	return p
}

func (p *params) TopP(v float64) xai.GenParams {
	p.params.TopP = param.NewOpt(v)
	return p
}

func (p *params) BaseURL(base string) xai.GenParams {
	p.opts = append(p.opts, option.WithBaseURL(base))
	return p
}

func (p *params) Timeout(timeout time.Duration) xai.GenParams {
	p.opts = append(p.opts, option.WithRequestTimeout(timeout))
	return p
}

func (p *Service) GenParams() xai.GenParams {
	return &params{}
}

func buildParams(in xai.GenParams) (anthropic.BetaMessageNewParams, []option.RequestOption) {
	p := in.(*params)
	// TODO(xsw): check param values
	return p.params, p.opts
}

// -----------------------------------------------------------------------------
