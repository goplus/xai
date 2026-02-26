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

package xai

import "context"

// -----------------------------------------------------------------------------

/*
// Role is the type of a message role, which can be system, assistant, user or tool.
type Role string

const (
	// RoleSystem is the role of a system, means the message is a system message.
	RoleSystem Role = "system"

	// RoleAssistant is the role of an assistant, means the message is returned by AI.
	RoleAssistant Role = "assistant"

	// RoleUser is the role of a user, means the message is a user message.
	RoleUser Role = "user"

	// RoleTool is the role of a tool, means the message is a tool call output.
	RoleTool Role = "tool"
)
*/

// -----------------------------------------------------------------------------

type OptionBuilder interface {
	WithBaseURL(base string) OptionBuilder
}

// -----------------------------------------------------------------------------

type ParamBuilder interface {
	// System prompt.
	//
	// A system prompt is a way of providing context and instructions to AI, such
	// as specifying a particular goal or role.
	System(TextBuilder) ParamBuilder

	// Input messages.
	//
	// AI models are trained to operate on alternating `user` and `assistant`
	// conversational turns. When creating a new `Message`, you specify the prior
	// conversational turns with the `messages` parameter, and the model then generates
	// the next `Message` in the conversation. Consecutive `user` or `assistant` turns
	// in your request will be combined into a single turn.
	//
	// Each input message must be an object with a `role` and `content`. You can
	// specify a single `user`-role message, or you can include multiple `user` and
	// `assistant` messages.
	//
	// If the final message uses the `assistant` role, the response content will
	// continue immediately from the content in that message. This can be used to
	// constrain part of the model's response.
	//
	// Example with a single `user` message:
	//
	// ```json
	// [{ "role": "user", "content": "Hello, AI" }]
	// ```
	//
	// Example with multiple conversational turns:
	//
	// ```json
	// [
	//
	//	{ "role": "user", "content": "Hello there." },
	//	{ "role": "assistant", "content": "Hi, I'm AI. How can I help you?" },
	//	{ "role": "user", "content": "Can you explain LLMs in plain English?" }
	//
	// ]
	// ```
	//
	// Example with a partially-filled response from AI:
	//
	// ```json
	// [
	//
	//	{
	//	  "role": "user",
	//	  "content": "What's the Greek name for Sun? (A) Sol (B) Helios (C) Sun"
	//	},
	//	{ "role": "assistant", "content": "The best answer is (" }
	//
	// ]
	// ```
	//
	// Each input message `content` may be either a single `string` or an array of
	// content blocks, where each block has a specific `type`. Using a `string` for
	// `content` is shorthand for an array of one content block of type `"text"`. The
	// following input messages are equivalent:
	//
	// ```json
	// { "role": "user", "content": "Hello, AI" }
	// ```
	//
	// ```json
	// { "role": "user", "content": [{ "type": "text", "text": "Hello, AI" }] }
	// ```
	//
	// Note that if you want to include a system prompt, you can use the
	// top-level `system` parameter — there is no `"system"` role for input messages in
	// the Messages API.
	//
	// There is a limit of 100,000 messages in a single request.
	Messages(MessageBuilder) ParamBuilder

	// The maximum number of tokens to generate before stopping.
	//
	// Note that our models may stop _before_ reaching this maximum. This parameter
	// only specifies the absolute maximum number of tokens to generate.
	//
	// Different models have different maximum values for this parameter.
	MaxTokens(int64) ParamBuilder

	// The model that will complete your prompt.
	Model(model Model) ParamBuilder

	// Container identifier for reuse across requests.
	Container(string) ParamBuilder

	// Specifies the geographic region for inference processing. If not specified, the
	// workspace's `default_inference_geo` is used.
	InferenceGeo(string) ParamBuilder

	// Amount of randomness injected into the response.
	//
	// Defaults to `1.0`. Ranges from `0.0` to `1.0`. Use `temperature` closer to `0.0`
	// for analytical / multiple choice, and closer to `1.0` for creative and
	// generative tasks.
	//
	// Note that even with `temperature` of `0.0`, the results will not be fully
	// deterministic.
	Temperature(float64) ParamBuilder

	// Only sample from the top K options for each subsequent token.
	//
	// Used to remove "long tail" low probability responses.
	// [Learn more technical details here](https://towardsdatascience.com/how-to-sample-from-language-models-682bceb97277).
	//
	// Recommended for advanced use cases only. You usually only need to use
	// `temperature`.
	TopK(int64) ParamBuilder

	// Use nucleus sampling.
	//
	// In nucleus sampling, we compute the cumulative distribution over all the options
	// for each subsequent token in decreasing probability order and cut it off once it
	// reaches a particular probability specified by `top_p`. You should either alter
	// `temperature` or `top_p`, but not both.
	//
	// Recommended for advanced use cases only. You usually only need to use
	// `temperature`.
	TopP(float64) ParamBuilder
}

// -----------------------------------------------------------------------------

type Provider interface {
	Options() OptionBuilder
	Params() ParamBuilder
	Messages() MessageBuilder
	Contents() ContentBuilder
	Images() ImageBuilder
	Docs() DocumentBuilder
	Texts() TextBuilder
	Tools() ToolBuilder

	// Send a structured list of input messages with text and/or image content, and the
	// model will generate the next message in the conversation.
	//
	// The Gen API can be used for either single queries or stateless multi-turn
	// conversations.
	//
	// Note: If you choose to set a timeout for this request, we recommend 10 minutes.
	Gen(ctx context.Context, params ParamBuilder, opts OptionBuilder) (Message, error)

	// Send a structured list of input messages with text and/or image content, and the
	// model will generate the next message in the conversation.
	//
	// The GenStreaming API can be used for either single queries or stateless multi-turn
	// conversations.
	//
	// Note: If you choose to set a timeout for this request, we recommend 10 minutes.
	GenStreaming(ctx context.Context, params ParamBuilder, opts OptionBuilder) StreamMessage
}

// -----------------------------------------------------------------------------
