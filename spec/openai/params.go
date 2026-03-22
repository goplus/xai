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
	"reflect"
	"time"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/util"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
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
	params  responses.ResponseNewParams
	pparams *util.Params[adapter]
	opts    []option.RequestOption
	sys     responses.ResponseInputMessageContentListParam
	msgs    []xai.MsgBuilder
	ctx     context.Context
}

/*
// Background bool
// Whether to run the model response in the background.
// [Learn more](https://platform.openai.com/docs/guides/background).

// Instructions string
// A system (or developer) message inserted into the model's context.
//
// When using along with `previous_response_id`, the instructions from a previous
// response will not be carried over to the next response. This makes it simple to
// swap out system (or developer) messages in new responses.

// MaxToolCalls int
// The maximum number of total calls to built-in tools that can be processed in a
// response. This maximum number applies across all built-in tool calls, not per
// individual tool. Any further attempts to call a tool by the model will be
// ignored.

// ParallelToolCalls bool
// Whether to allow the model to run tool calls in parallel.

// PreviousResponseID string
// The unique ID of the previous response to the model. Use this to create
// multi-turn conversations. Learn more about
// [conversation state](https://platform.openai.com/docs/guides/conversation-state).
// Cannot be used in conjunction with `conversation`.

// Store bool
// Whether to store the generated model response for later retrieval via API.

// TopLogprobs int
// An integer between 0 and 20 specifying the number of most likely tokens to
// return at each token position, each with an associated log probability.

// PromptCacheKey string
// Used by OpenAI to cache responses for similar requests to optimize your cache
// hit rates. Replaces the `user` field.
// [Learn more](https://platform.openai.com/docs/guides/prompt-caching).

// SafetyIdentifier string
// A stable identifier used to help detect users of your application that may be
// violating OpenAI's usage policies. The IDs should be a string that uniquely
// identifies each user, with a maximum length of 64 characters. We recommend
// hashing their username or email address, in order to avoid sending us any
// identifying information.
// [Learn more](https://platform.openai.com/docs/guides/safety-best-practices#safety-identifiers).

// User string
// This field is being replaced by `safety_identifier` and `prompt_cache_key`. Use
// `prompt_cache_key` instead to maintain caching optimizations. A stable
// identifier for your end-users. Used to boost cache hit rates by better bucketing
// similar requests and to help OpenAI detect and prevent abuse.
// [Learn more](https://platform.openai.com/docs/guides/safety-best-practices#safety-identifiers).

// ContextManagement []ResponseNewParamsContextManagement
// Context management configuration for this request.

// Conversation ResponseNewParamsConversationUnion
// The conversation that this response belongs to. Items from this conversation are
// prepended to `input_items` for this response request. Input items and output
// items from this response are automatically added to this conversation after this
// response completes.

// Include []ResponseIncludable
// Specify additional output data to include in the model response. Currently
// supported values are:
//
//   - `web_search_call.action.sources`: Include the sources of the web search tool
//     call.
//   - `code_interpreter_call.outputs`: Includes the outputs of python code execution
//     in code interpreter tool call items.
//   - `computer_call_output.output.image_url`: Include image urls from the computer
//     call output.
//   - `file_search_call.results`: Include the search results of the file search tool
//     call.
//   - `message.input_image.image_url`: Include image urls from the input message.
//   - `message.output_text.logprobs`: Include logprobs with assistant messages.
//   - `reasoning.encrypted_content`: Includes an encrypted version of reasoning
//     tokens in reasoning item outputs. This enables reasoning items to be used in
//     multi-turn conversations when using the Responses API statelessly (like when
//     the `store` parameter is set to `false`, or when an organization is enrolled
//     in the zero data retention program).

// Metadata shared.Metadata
// Set of 16 key-value pairs that can be attached to an object. This can be useful
// for storing additional information about the object in a structured format, and
// querying for objects via API or the dashboard.
//
// Keys are strings with a maximum length of 64 characters. Values are strings with
// a maximum length of 512 characters.

// Prompt ResponsePromptParam
// Reference to a prompt template and its variables.
// [Learn more](https://platform.openai.com/docs/guides/text?api-mode=responses#reusable-prompts).

// PromptCacheRetention ResponseNewParamsPromptCacheRetention
// The retention policy for the prompt cache. Set to `24h` to enable extended
// prompt caching, which keeps cached prefixes active for longer, up to a maximum
// of 24 hours.
// [Learn more](https://platform.openai.com/docs/guides/prompt-caching#prompt-cache-retention).
//
// Any of "in-memory", "24h".

// ServiceTier ResponseNewParamsServiceTier
// Specifies the processing type used for serving the request.
//
//   - If set to 'auto', then the request will be processed with the service tier
//     configured in the Project settings. Unless otherwise configured, the Project
//     will use 'default'.
//   - If set to 'default', then the request will be processed with the standard
//     pricing and performance for the selected model.
//   - If set to '[flex](https://platform.openai.com/docs/guides/flex-processing)' or
//     '[priority](https://openai.com/api-priority-processing/)', then the request
//     will be processed with the corresponding service tier.
//   - When not set, the default behavior is 'auto'.
//
// When the `service_tier` parameter is set, the response body will include the
// `service_tier` value based on the processing mode actually used to serve the
// request. This response value may be different from the value set in the
// parameter.
//
// Any of "auto", "default", "flex", "scale", "priority".

// StreamOptions ResponseNewParamsStreamOptions
// Options for streaming responses. Only set this when you set `stream: true`.

// Truncation ResponseNewParamsTruncation
// The truncation strategy to use for the model response.
//
//   - `auto`: If the input to this Response exceeds the model's context window size,
//     the model will truncate the response to fit the context window by dropping
//     items from the beginning of the conversation.
//   - `disabled` (default): If the input size will exceed the context window size
//     for a model, the request will fail with a 400 error.
//
// Any of "auto", "disabled".

// Reasoning shared.ReasoningParam
// **gpt-5 and o-series models only**
//
// Configuration options for
// [reasoning models](https://platform.openai.com/docs/guides/reasoning).

// Text ResponseTextConfigParam
// Configuration options for a text response from the model. Can be plain text or
// structured JSON data. Learn more:
//
// - [Text inputs and outputs](https://platform.openai.com/docs/guides/text)
// - [Structured Outputs](https://platform.openai.com/docs/guides/structured-outputs)

// ToolChoice ResponseNewParamsToolChoiceUnion
// How the model should select which tool (or tools) to use when generating a
// response. See the `tools` parameter to see how to specify which tools the model
// can call.
*/
func (p *params) Set(name string, val any) xai.GenParams {
	if p.pparams == nil {
		p.pparams = util.NewParams[adapter](&p.params)
	}
	p.pparams.Set(name, val)
	return p
}

func (p *params) System(texts ...string) xai.GenParams {
	content := make(responses.ResponseInputMessageContentListParam, len(texts))
	for i, text := range texts {
		content[i] = responses.ResponseInputContentParamOfInputText(text)
	}
	p.sys = content
	return p
}

func (p *params) Messages(msgs ...xai.MsgBuilder) xai.GenParams {
	// we will merge system prompt and messages into input param in buildParams
	// so we just store the messages here
	p.msgs = msgs
	return p
}

func (p *params) Tools(tools ...xai.ToolBase) xai.GenParams {
	p.params.Tools = buildTools(tools)
	return p
}

func (p *params) Model(model xai.Model) xai.GenParams {
	p.params.Model = shared.ResponsesModel(model) // TODO(xsw): validate model
	return p
}

func (p *params) MaxOutputTokens(v int64) xai.GenParams {
	p.params.MaxOutputTokens = param.NewOpt(v)
	return p
}

func (p *params) Compact(maxInputTokens int64) xai.GenParams {
	panic("todo")
}

func (p *params) Temperature(v float64) xai.GenParams {
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

func (p *params) Ctx(ctx context.Context) xai.GenParams {
	p.ctx = ctx
	return p
}

func (p *Service) GenParams() xai.GenParams {
	return &params{}
}

func buildParams(in xai.GenParams) (context.Context, responses.ResponseNewParams, []option.RequestOption) {
	p := in.(*params)
	// TODO(xsw): check param values
	ctx := p.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	// Merge system prompt and messages into input param
	var sys responses.ResponseInputItemUnionParam
	if len(p.sys) > 0 {
		sys = responses.ResponseInputItemParamOfMessage(p.sys, responses.EasyInputMessageRoleSystem)
	}
	p.params.Input = buildMessages(p.msgs, sys)
	return ctx, p.params, p.opts
}

// -----------------------------------------------------------------------------
