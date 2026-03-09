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
	"encoding/json"

	xai "github.com/goplus/xai/spec"
	openai "github.com/openai/openai-go/v3"
)

// -----------------------------------------------------------------------------
// V1 Response (Chat Completions API)
// -----------------------------------------------------------------------------

// response_v1.go adapts OpenAI Chat Completions responses (a.k.a. v1 API) into
// xai.GenResponse / xai.Candidate / xai.Part.
//
// Usage:
//   - Non-streaming: provider_v1.Gen returns *v1Response.
//   - Streaming: provider_v1.GenStream emits *v1StreamChunk.
//   - Agent loop: consume candidate parts with Text()/AsToolUse()/AsBlob(), then
//     call candidate.ToMsg() to append assistant output back into conversation
//     history for the next turn.
//
// Compatibility notes:
//   - Supports both modern tool_calls and legacy function_call payloads.
//   - Supports provider extensions that return image blobs as extra parts
//     (v1ResponseWithImages).

// v1ContentBlock represents one logical output part in a v1 assistant message.
// Exactly one field is expected to be populated:
//   - text: plain assistant text
//   - toolCall: modern tool_call entry
//   - legacyFunctionCall: deprecated function_call fallback
//
// It implements xai.Part.
type v1ContentBlock struct {
	text               string
	toolCall           *openai.ChatCompletionMessageToolCallUnion
	legacyFunctionCall *v1LegacyFunctionCall
}

// AsThinking always returns false for v1 chat completions because this API does
// not expose structured thinking blocks in assistant output.
func (p v1ContentBlock) AsThinking() (ret xai.Thinking, ok bool) {
	return
}

// AsToolUse maps either tool_calls or legacy function_call to xai.ToolUse.
//
// The Input field is normalized as:
//   - json.RawMessage when the source string is valid JSON
//   - map[string]any{"raw": <source>} when the source is non-JSON text
func (p v1ContentBlock) AsToolUse() (ret xai.ToolUse, ok bool) {
	if p.toolCall != nil {
		switch p.toolCall.Type {
		case "custom":
			u := p.toolCall.AsCustom()
			ret.ID = u.ID
			ret.Name = u.Custom.Name
			ret.Input = rawJSONOrString(u.Custom.Input)
			ret.Underlying = &u
		default:
			u := p.toolCall.AsFunction()
			ret.ID = u.ID
			ret.Name = u.Function.Name
			ret.Input = rawJSONOrString(u.Function.Arguments)
			ret.Underlying = &u
		}
		return ret, true
	}
	if p.legacyFunctionCall != nil {
		ret.Name = p.legacyFunctionCall.Name
		ret.Input = rawJSONOrString(p.legacyFunctionCall.Arguments)
		ret.Underlying = p.legacyFunctionCall
		return ret, true
	}
	return
}

// AsToolResult always returns false for v1 assistant output blocks.
// Tool results are represented as tool-role input messages, not assistant parts.
func (p v1ContentBlock) AsToolResult() (ret xai.ToolResult, ok bool) {
	return
}

// AsBlob always returns false for normal v1 content blocks.
// Blob blocks use v1ContentBlockBlob.
func (p v1ContentBlock) AsBlob() (ret xai.Blob, ok bool) {
	return
}

// AsCompaction always returns false for v1 chat completions.
func (p v1ContentBlock) AsCompaction() (ret xai.Compaction, ok bool) {
	return
}

// Text returns plain text when this block is text.
func (p v1ContentBlock) Text() string {
	return p.text
}

// Underlying returns the original SDK object (tool call / legacy function call)
// when available, otherwise returns the text string.
func (p v1ContentBlock) Underlying() any {
	if p.toolCall != nil {
		return p.toolCall
	}
	if p.legacyFunctionCall != nil {
		return p.legacyFunctionCall
	}
	return p.text
}

// v1ContentBlockBlob represents an image/blob part produced by provider-specific
// extensions (e.g. Qiniu image generation responses).
//
// It implements xai.Part with only AsBlob returning true.
type v1ContentBlockBlob struct {
	blob xai.Blob
}

func (p v1ContentBlockBlob) AsThinking() (ret xai.Thinking, ok bool)     { return }
func (p v1ContentBlockBlob) AsToolUse() (ret xai.ToolUse, ok bool)       { return }
func (p v1ContentBlockBlob) AsToolResult() (ret xai.ToolResult, ok bool) { return }
func (p v1ContentBlockBlob) AsCompaction() (ret xai.Compaction, ok bool) { return }
func (p v1ContentBlockBlob) Text() string                                { return "" }
func (p v1ContentBlockBlob) Underlying() any                             { return p.blob }
func (p v1ContentBlockBlob) AsBlob() (ret xai.Blob, ok bool)             { return p.blob, true }

// chatCompletionMessageRaw is a local parser for provider-extended message JSON
// that may include `images`.
//
// This is intentionally separate from SDK types so we can tolerate non-standard
// fields while still preserving compatibility.
type chatCompletionMessageRaw struct {
	Content string `json:"content"`
	Images  []struct {
		Type     string `json:"type"`
		ImageURL struct {
			URL string `json:"url"`
		} `json:"image_url"`
	} `json:"images"`
	ToolCalls []openai.ChatCompletionMessageToolCallUnion `json:"tool_calls"`
}

// v1LegacyFunctionCall models deprecated function_call payloads.
//
// We parse this from raw JSON to avoid directly depending on deprecated SDK
// fields and to keep backward compatibility with older OpenAI-compatible
// gateways.
type v1LegacyFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// parseV1LegacyFunctionCall extracts function_call from raw message JSON.
//
// Returns nil if:
//   - raw is empty
//   - JSON parsing fails
//   - function_call is missing or name is empty
func parseV1LegacyFunctionCall(raw string) *v1LegacyFunctionCall {
	if raw == "" {
		return nil
	}
	var msg struct {
		FunctionCall *v1LegacyFunctionCall `json:"function_call"`
	}
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		return nil
	}
	if msg.FunctionCall == nil || msg.FunctionCall.Name == "" {
		return nil
	}
	return msg.FunctionCall
}

// v1Candidate represents one choice in chat completion results.
//
// Part order is deterministic:
//  1. text (if present)
//  2. image blobs (if any)
//  3. tool calls (if any)
//  4. legacy function_call (if present)
//
// This ordering allows callers to iterate parts once and render/extract outputs
// in stable sequence.
type v1Candidate struct {
	finishReason       string
	text               string
	toolCalls          []openai.ChatCompletionMessageToolCallUnion
	legacyFunctionCall *v1LegacyFunctionCall
	images             []xai.Blob
}

// newV1Candidate constructs a normalized candidate from an SDK choice and
// optional provider-extended data.
//
// If raw is provided, it can override text and tool_calls when the upstream
// provider emits non-standard message envelopes.
func newV1Candidate(choice *openai.ChatCompletionChoice, images []xai.Blob, raw *chatCompletionMessageRaw) *v1Candidate {
	if choice == nil {
		return &v1Candidate{}
	}
	text := choice.Message.Content
	if raw != nil && raw.Content != "" {
		text = raw.Content
	}
	toolCalls := choice.Message.ToolCalls
	if raw != nil && len(raw.ToolCalls) > 0 {
		toolCalls = raw.ToolCalls
	}
	candidate := &v1Candidate{
		finishReason: choice.FinishReason,
		text:         text,
		toolCalls:    toolCalls,
		images:       images,
	}
	if len(toolCalls) == 0 {
		candidate.legacyFunctionCall = parseV1LegacyFunctionCall(choice.Message.RawJSON())
	}
	return candidate
}

// StopReason maps OpenAI finish_reason to xai.StopReason.
func (p *v1Candidate) StopReason() xai.StopReason {
	return v1StopReason(p.finishReason)
}

// Parts returns the number of normalized parts in this candidate.
func (p *v1Candidate) Parts() int {
	n := 0
	if p.text != "" {
		n++
	}
	n += len(p.images)
	n += len(p.toolCalls)
	if p.legacyFunctionCall != nil {
		n++
	}
	return n
}

// Part returns the i-th part according to the stable ordering documented on
// v1Candidate. It panics on out-of-range indices.
func (p *v1Candidate) Part(i int) xai.Part {
	n := p.Parts()
	if i < 0 || i >= n {
		panicIndex("v1Candidate.Part", i, n)
	}
	if p.text != "" {
		if i == 0 {
			return v1ContentBlock{text: p.text}
		}
		i--
	}
	if i < len(p.images) {
		return v1ContentBlockBlob{blob: p.images[i]}
	}
	i -= len(p.images)
	if i < len(p.toolCalls) {
		return v1ContentBlock{toolCall: &p.toolCalls[i]}
	}
	i -= len(p.toolCalls)
	if p.legacyFunctionCall != nil && i == 0 {
		return v1ContentBlock{legacyFunctionCall: p.legacyFunctionCall}
	}
	panicIndex("v1Candidate.Part", i, n)
	return nil
}

// ToMsg converts the candidate into an assistant MsgBuilder for conversation
// history replay.
//
// Usage:
//
//	history = append(history, resp.At(0).ToMsg())
func (p *v1Candidate) ToMsg() xai.MsgBuilder {
	mb := newAssistantMsg()
	appendTextContent(mb.msg, p.text)
	for i := range p.toolCalls {
		toolUse, ok := (v1ContentBlock{toolCall: &p.toolCalls[i]}).AsToolUse()
		if ok {
			appendToolUseContent(mb.msg, toolUse)
		}
	}
	if p.legacyFunctionCall != nil {
		toolUse, ok := (v1ContentBlock{legacyFunctionCall: p.legacyFunctionCall}).AsToolUse()
		if ok {
			appendToolUseContent(mb.msg, toolUse)
		}
	}
	return mb
}

// v1Response is the normal non-streaming v1 wrapper.
//
// It implements both xai.GenResponse and xai.Candidate for convenience:
//   - Len/At expose all choices.
//   - StopReason/Parts/Part/ToMsg delegate to the first choice.
type v1Response struct {
	msg *openai.ChatCompletion
}

// Len returns number of choices.
func (p *v1Response) Len() int {
	if p.msg == nil {
		return 0
	}
	return len(p.msg.Choices)
}

// At returns a normalized candidate for choice i.
func (p *v1Response) At(i int) xai.Candidate {
	n := p.Len()
	if i < 0 || i >= n {
		panicIndex("v1Response.At", i, n)
	}
	return newV1Candidate(&p.msg.Choices[i], nil, nil)
}

func (p *v1Response) firstCandidate() *v1Candidate {
	if p.Len() == 0 {
		return &v1Candidate{}
	}
	return newV1Candidate(&p.msg.Choices[0], nil, nil)
}

// StopReason proxies the first candidate stop reason.
func (p *v1Response) StopReason() xai.StopReason {
	return p.firstCandidate().StopReason()
}

// Parts proxies the first candidate part count.
func (p *v1Response) Parts() int {
	return p.firstCandidate().Parts()
}

// Part proxies the first candidate part lookup.
func (p *v1Response) Part(i int) xai.Part {
	return p.firstCandidate().Part(i)
}

// ToMsg proxies the first candidate message conversion.
func (p *v1Response) ToMsg() xai.MsgBuilder {
	return p.firstCandidate().ToMsg()
}

// v1ResponseWithImages is a v1 response wrapper that additionally carries
// per-choice image blobs and provider raw message payloads.
//
// This is used by providers that return non-standard `images` fields while still
// preserving the xai.GenResponse surface.
type v1ResponseWithImages struct {
	msg    *openai.ChatCompletion
	images [][]xai.Blob // per-choice images
	raw    []*chatCompletionMessageRaw
}

// Len returns number of choices.
func (p *v1ResponseWithImages) Len() int {
	if p.msg == nil {
		return 0
	}
	return len(p.msg.Choices)
}

func (p *v1ResponseWithImages) candidateAt(i int) *v1Candidate {
	choice := &p.msg.Choices[i]
	var imgs []xai.Blob
	if i < len(p.images) {
		imgs = p.images[i]
	}
	var raw *chatCompletionMessageRaw
	if i < len(p.raw) {
		raw = p.raw[i]
	}
	return newV1Candidate(choice, imgs, raw)
}

// At returns candidate i with merged standard + extended fields.
func (p *v1ResponseWithImages) At(i int) xai.Candidate {
	n := p.Len()
	if i < 0 || i >= n {
		panicIndex("v1ResponseWithImages.At", i, n)
	}
	return p.candidateAt(i)
}

// ToMsg returns a replayable assistant message from the first choice.
func (p *v1ResponseWithImages) ToMsg() xai.MsgBuilder {
	if p.Len() == 0 {
		return newAssistantMsg()
	}
	return p.candidateAt(0).ToMsg()
}

// v1StreamChunk is the minimal candidate wrapper used by v1 streaming deltas.
//
// Each chunk exposes exactly one text part.
type v1StreamChunk struct {
	text string
}

func (p *v1StreamChunk) StopReason() xai.StopReason { return xai.Unspecified }
func (p *v1StreamChunk) Parts() int                 { return 1 }
func (p *v1StreamChunk) Part(i int) xai.Part {
	if i != 0 {
		panicIndex("v1StreamChunk.Part", i, 1)
	}
	return v1ContentBlock{text: p.text}
}
func (p *v1StreamChunk) Len() int { return 1 }
func (p *v1StreamChunk) At(i int) xai.Candidate {
	if i != 0 {
		panicIndex("v1StreamChunk.At", i, 1)
	}
	return p
}

// ToMsg converts a streaming text delta into an assistant message.
//
// Usage:
//
//	for resp, _ := range svc.GenStream(...) {
//	    history = append(history, resp.At(0).ToMsg())
//	}
func (p *v1StreamChunk) ToMsg() xai.MsgBuilder { return newAssistantTextMsg(p.text) }

// -----------------------------------------------------------------------------
