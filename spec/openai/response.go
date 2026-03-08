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
	"strings"
	"unsafe"

	xai "github.com/goplus/xai/spec"
	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/responses"
)

// -----------------------------------------------------------------------------
// V3 Response (Responses API)
// -----------------------------------------------------------------------------

type v3ContentBlock struct {
	content *responses.ResponseOutputItemUnion
}

func (p v3ContentBlock) AsThinking() (ret xai.Thinking, ok bool) {
	switch p.content.Type {
	case "reasoning":
		u := p.content.AsReasoning()
		ret.Underlying = &u
	default:
		return
	}
	ok = true
	panic("todo")
}

func (p v3ContentBlock) AsToolUse() (ret xai.ToolUse, ok bool) {
	switch p.content.Type {
	case "function_call":
		u := p.content.AsFunctionCall()
		ret.ID = u.ID
		ret.Name = u.Name
		ret.Input = rawMessage(u.Arguments)
		ret.Underlying = &u
	case "file_search_call", "web_search_call", "computer_call", "code_interpreter_call",
		"local_shell_call", "shell_call", "apply_patch_call", "mcp_call", "custom_tool_call":
		panic("todo")
	default:
		return
	}
	ok = true
	return
}

func (p v3ContentBlock) AsToolResult() (ret xai.ToolResult, ok bool) {
	panic("todo")
}

func (p v3ContentBlock) AsBlob() (ret xai.Blob, ok bool) {
	panic("todo")
}

func (p v3ContentBlock) AsCompaction() (ret xai.Compaction, ok bool) {
	switch p.content.Type {
	case "compaction":
		u := p.content.AsCompaction()
		ret.Data = u.EncryptedContent
	default:
		return
	}
	ok = true
	return
}

func (p v3ContentBlock) Text() string {
	if len(p.content.Content) == 0 {
		return ""
	}
	var outputText strings.Builder
	for _, content := range p.content.Content {
		if content.Type == "output_text" {
			outputText.WriteString(content.Text)
		}
	}
	return outputText.String()
}

func (p v3ContentBlock) Underlying() any {
	return p.content
}

func v3ContentBlockFromText(text string) v3ContentBlock {
	return v3ContentBlock{&responses.ResponseOutputItemUnion{
		Type: "message",
		Content: []responses.ResponseOutputMessageContentUnion{
			{Type: "output_text", Text: text},
		},
	}}
}

func rawMessage(msg string) json.RawMessage {
	b := unsafe.Slice(unsafe.StringData(msg), len(msg))
	return json.RawMessage(b)
}

// -----------------------------------------------------------------------------

type v3Response struct {
	msg *responses.Response
}

func (p *v3Response) StopReason() xai.StopReason {
	switch p.msg.Status {
	case responses.ResponseStatusCompleted:
		return xai.EndTurn
	case responses.ResponseStatusIncomplete:
		switch p.msg.IncompleteDetails.Reason {
		case "max_output_tokens":
			return xai.StopMaxTokens
		case "content_filter":
			return xai.Refusal
		}
	default:
		panic("todo")
	}
	return xai.Unspecified
}

func (p *v3Response) Parts() int {
	return len(p.msg.Output)
}

func (p *v3Response) Part(i int) xai.Part {
	return v3ContentBlock{&p.msg.Output[i]}
}

func (p *v3Response) Len() int {
	return 1
}

func (p *v3Response) At(i int) xai.Candidate {
	if i != 0 {
		panic("v3Response.At: index out of range")
	}
	return p
}

func (p *v3Response) ToMsg() xai.MsgBuilder {
	panic("todo")
}

// -----------------------------------------------------------------------------

type v3StreamChunk struct {
	text string
}

func (p *v3StreamChunk) StopReason() xai.StopReason { return xai.Unspecified }
func (p *v3StreamChunk) Parts() int                 { return 1 }
func (p *v3StreamChunk) Part(i int) xai.Part        { return v3ContentBlockFromText(p.text) }
func (p *v3StreamChunk) Len() int                   { return 1 }
func (p *v3StreamChunk) At(i int) xai.Candidate {
	if i != 0 {
		panic("v3StreamChunk.At: index out of range")
	}
	return p
}
func (p *v3StreamChunk) ToMsg() xai.MsgBuilder { panic("todo") }

// -----------------------------------------------------------------------------
// V1 Response (Chat Completions API)
// -----------------------------------------------------------------------------

type v1ContentBlock struct {
	text     string
	toolCall *openai.ChatCompletionMessageToolCallUnion
}

func (p v1ContentBlock) AsThinking() (ret xai.Thinking, ok bool) {
	return
}

func (p v1ContentBlock) AsToolUse() (ret xai.ToolUse, ok bool) {
	if p.toolCall == nil {
		return
	}
	ret.ID = p.toolCall.ID
	ret.Name = p.toolCall.Function.Name
	ret.Input = rawMessage(p.toolCall.Function.Arguments)
	ret.Underlying = p.toolCall
	ok = true
	return
}

func (p v1ContentBlock) AsToolResult() (ret xai.ToolResult, ok bool) {
	return
}

func (p v1ContentBlock) AsBlob() (ret xai.Blob, ok bool) {
	return
}

func (p v1ContentBlock) AsCompaction() (ret xai.Compaction, ok bool) {
	return
}

func (p v1ContentBlock) Text() string {
	return p.text
}

func (p v1ContentBlock) Underlying() any {
	if p.toolCall != nil {
		return p.toolCall
	}
	return p.text
}

// -----------------------------------------------------------------------------

type v1Response struct {
	msg *openai.ChatCompletion
}

func (p *v1Response) StopReason() xai.StopReason {
	if len(p.msg.Choices) == 0 {
		return xai.Unspecified
	}
	switch p.msg.Choices[0].FinishReason {
	case "stop":
		return xai.EndTurn
	case "length":
		return xai.StopMaxTokens
	case "content_filter":
		return xai.Refusal
	case "tool_calls":
		return xai.PauseTurn
	}
	return xai.Unspecified
}

func (p *v1Response) Parts() int {
	if len(p.msg.Choices) == 0 {
		return 0
	}
	choice := &p.msg.Choices[0]
	n := 0
	if choice.Message.Content != "" {
		n++
	}
	n += len(choice.Message.ToolCalls)
	return n
}

func (p *v1Response) Part(i int) xai.Part {
	choice := &p.msg.Choices[0]
	if choice.Message.Content != "" {
		if i == 0 {
			return v1ContentBlock{text: choice.Message.Content}
		}
		i--
	}
	return v1ContentBlock{toolCall: &choice.Message.ToolCalls[i]}
}

func (p *v1Response) Len() int {
	return len(p.msg.Choices)
}

func (p *v1Response) At(i int) xai.Candidate {
	return &v1ChoiceResponse{choice: &p.msg.Choices[i]}
}

func (p *v1Response) ToMsg() xai.MsgBuilder {
	panic("todo")
}

// -----------------------------------------------------------------------------

type v1ChoiceResponse struct {
	choice *openai.ChatCompletionChoice
}

func (p *v1ChoiceResponse) StopReason() xai.StopReason {
	switch p.choice.FinishReason {
	case "stop":
		return xai.EndTurn
	case "length":
		return xai.StopMaxTokens
	case "content_filter":
		return xai.Refusal
	case "tool_calls":
		return xai.PauseTurn
	}
	return xai.Unspecified
}

func (p *v1ChoiceResponse) Parts() int {
	n := 0
	if p.choice.Message.Content != "" {
		n++
	}
	n += len(p.choice.Message.ToolCalls)
	return n
}

func (p *v1ChoiceResponse) Part(i int) xai.Part {
	if p.choice.Message.Content != "" {
		if i == 0 {
			return v1ContentBlock{text: p.choice.Message.Content}
		}
		i--
	}
	return v1ContentBlock{toolCall: &p.choice.Message.ToolCalls[i]}
}

func (p *v1ChoiceResponse) ToMsg() xai.MsgBuilder {
	panic("todo")
}

// -----------------------------------------------------------------------------

type v1StreamChunk struct {
	text string
}

func (p *v1StreamChunk) StopReason() xai.StopReason { return xai.Unspecified }
func (p *v1StreamChunk) Parts() int                 { return 1 }
func (p *v1StreamChunk) Part(i int) xai.Part        { return v1ContentBlock{text: p.text} }
func (p *v1StreamChunk) Len() int                   { return 1 }
func (p *v1StreamChunk) At(i int) xai.Candidate {
	if i != 0 {
		panic("v1StreamChunk.At: index out of range")
	}
	return p
}
func (p *v1StreamChunk) ToMsg() xai.MsgBuilder { panic("todo") }

// -----------------------------------------------------------------------------
// Shared
// -----------------------------------------------------------------------------

type streamError struct{ msg string }

func (e *streamError) Error() string { return e.msg }

// -----------------------------------------------------------------------------
