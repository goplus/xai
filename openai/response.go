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
	"iter"
	"strings"
	"unsafe"

	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/packages/ssestream"
	"github.com/openai/openai-go/v3/responses"
)

// -----------------------------------------------------------------------------

type contentBlock struct {
	content *responses.ResponseOutputItemUnion
}

func (p contentBlock) AsThinking() (ret xai.Thinking, ok bool) {
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

func (p contentBlock) AsToolUse() (ret xai.ToolUse, ok bool) {
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

func (p contentBlock) AsToolResult() (ret xai.ToolResult, ok bool) {
	panic("todo")
}

func (p contentBlock) AsBlob() (ret xai.Blob, ok bool) {
	panic("todo")
}

func (p contentBlock) AsCompaction() (ret xai.Compaction, ok bool) {
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

func (p contentBlock) Text() string {
	if len(p.content.Content) == 0 {
		return "" // for non-text content, we return empty string as text content.
	}
	var outputText strings.Builder
	for _, content := range p.content.Content {
		if content.Type == "output_text" {
			outputText.WriteString(content.Text)
		}
	}
	return outputText.String()
}

func (p contentBlock) Underlying() any {
	return p.content
}

func rawMessage(msg string) json.RawMessage {
	b := unsafe.Slice(unsafe.StringData(msg), len(msg))
	return json.RawMessage(b)
}

// -----------------------------------------------------------------------------

type response struct {
	msg *responses.Response
}

func (p response) StopReason() xai.StopReason {
	switch p.msg.Status {
	case responses.ResponseStatusCompleted:
		return xai.EndTurn
	case responses.ResponseStatusIncomplete:
		switch p.msg.IncompleteDetails.Reason {
		case "max_output_tokens":
			return xai.MaxTokens
		case "content_filter":
			return xai.Refusal
		}
	default:
		// TODO(xsw): map other status to stop reason.
		panic("todo")
	}
	return xai.Unspecified
}

func (p response) Parts() int {
	return len(p.msg.Output)
}

func (p response) Part(i int) xai.Part {
	return contentBlock{&p.msg.Output[i]}
}

func buildPart(part xai.Part) responses.ResponseInputItemUnionParam {
	panic("todo")
}

func (p response) Len() int {
	return 1
}

func (p response) At(i int) xai.Candidate {
	if i != 0 {
		panic("response.At: index out of range")
	}
	return p
}

func (p response) ToMsg() xai.MsgBuilder {
	panic("todo")
}

// -----------------------------------------------------------------------------

func buildRespIter(stream *ssestream.Stream[responses.ResponseStreamEventUnion]) iter.Seq2[xai.GenResponse, error] {
	panic("todo")
}

// -----------------------------------------------------------------------------
