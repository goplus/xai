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
	"errors"
	"iter"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/ssestream"
	"github.com/goplus/xai"
)

// -----------------------------------------------------------------------------

type contentBlock struct {
	content *anthropic.BetaContentBlockUnion
}

func (p contentBlock) AsThinking() (ret xai.Thinking, ok bool) {
	switch p.content.Type {
	case "thinking":
		u := p.content.AsThinking()
		ret.Signature = u.Signature
		ret.Text = u.Thinking
		ret.Underlying = &u
	case "redacted_thinking":
		u := p.content.AsRedactedThinking()
		ret.Signature = u.Data
		ret.Redacted = true
		ret.Underlying = &u
	default:
		return
	}
	ok = true
	return
}

func (p contentBlock) AsToolUse() (ret xai.ToolUse, ok bool) {
	switch p.content.Type {
	case "tool_use":
		u := p.content.AsToolUse()
		ret.ID = u.ID
		ret.Name = u.Name
		ret.Input = u.Input
		ret.Underlying = &u
	case "server_tool_use":
		u := p.content.AsServerToolUse()
		ret.ID = u.ID
		ret.Name = "std/" + string(u.Name)
		ret.Input = u.Input
		ret.Underlying = &u
	case "mcp_tool_use":
		panic("todo")
	default:
		return
	}
	ok = true
	return
}

func (p contentBlock) AsToolResult() (ret xai.ToolResult, ok bool) {
	switch p.content.Type {
	case "web_search_tool_result":
		u := p.content.AsWebSearchToolResult()
		ret.ID = u.ToolUseID
		ret.Name = xai.ToolWebSearch
		if u.Content.ErrorCode != "" {
			ret.Result = errors.New(string(u.Content.ErrorCode))
			ret.IsError = true
		} else {
			items := u.Content.OfBetaWebSearchResultBlockArray
			result := make([]xai.WebSearchResultItem, len(items))
			for i, item := range items {
				result[i] = xai.WebSearchResultItem{
					Title:   item.Title,
					URL:     item.URL,
					PageAge: item.PageAge,
				}
			}
			ret.Result = &xai.WebSearchResult{
				Result:     result,
				Underlying: &u,
			}
		}
		ret.Underlying = &u
	case "web_fetch_tool_result", "code_execution_tool_result",
		"bash_code_execution_tool_result", "text_editor_code_execution_tool_result",
		"tool_search_tool_result", "mcp_tool_result":
		panic("todo")
	default:
		return
	}
	ok = true
	return
}

func (p contentBlock) Text() string {
	return p.content.Text
}

func (p contentBlock) Underlying() any {
	return p.content
}

// -----------------------------------------------------------------------------

type response struct {
	msg *anthropic.BetaMessage
}

func (p response) StopReason() xai.StopReason {
	reason := p.msg.StopReason
	if reason == anthropic.BetaStopReasonToolUse {
		// NOTE(xsw): treat tool use as end turn, since the tool response will
		// be included in the content.
		reason = anthropic.BetaStopReasonEndTurn
	}
	return xai.StopReason(reason)
}

func (p response) Parts() int {
	return len(p.msg.Content)
}

func (p response) Part(i int) xai.Part {
	return contentBlock{&p.msg.Content[i]}
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
	content := make([]anthropic.BetaContentBlockParamUnion, len(p.msg.Content))
	for i, c := range p.msg.Content {
		content[i] = c.ToParam()
	}
	return &msgBuilder{content: content, role: anthropic.BetaMessageParamRoleAssistant}
}

// -----------------------------------------------------------------------------

func buildRespIter(stream *ssestream.Stream[anthropic.BetaRawMessageStreamEventUnion]) iter.Seq2[xai.GenResponse, error] {
	panic("todo")
}

// -----------------------------------------------------------------------------
