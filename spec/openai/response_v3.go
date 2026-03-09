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
	"errors"
	"fmt"
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/responses"
)

// -----------------------------------------------------------------------------
// Shared helpers
// -----------------------------------------------------------------------------

// response_v3.go adapts OpenAI Responses API (v3) payloads into xai interfaces
// and also hosts shared helper functions used by both v3 and v1 wrappers.
//
// High-level usage:
//   1. Gen: provider_v3.Gen returns *v3Response.
//   2. Read candidate parts using Text()/AsThinking()/AsToolUse()/AsToolResult().
//   3. Append candidate.ToMsg() to your history for multi-turn loops.
//   4. Stream: provider_v3.GenStream yields *v3StreamChunk text deltas.
//
// Design goals:
//   - Stable part traversal across mixed output types.
//   - Lossless access to original SDK objects via Underlying().
//   - Defensive JSON handling for tool inputs/outputs.

// panicIndex formats a uniform panic for out-of-range index access.
func panicIndex(name string, i, n int) {
	panic(fmt.Sprintf("%s: index %d out of range [0,%d)", name, i, n))
}

// coalesce returns the first non-empty string.
func coalesce(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// rawJSONOrString parses a JSON-looking string into json.RawMessage; otherwise it
// wraps the original text as map[string]any{"raw": <text>}.
//
// This lets downstream code keep structured payloads structured while preserving
// non-JSON provider outputs safely.
func rawJSONOrString(in string) any {
	s := strings.TrimSpace(in)
	if s == "" {
		return json.RawMessage("{}")
	}
	buf := []byte(s)
	if json.Valid(buf) {
		return json.RawMessage(append([]byte(nil), buf...))
	}
	return map[string]any{"raw": in}
}

// marshalToolInput serializes any tool input into JSON. It preserves valid JSON
// bytes/raw messages and wraps invalid byte/string payloads in {"raw": ...}.
func marshalToolInput(in any) json.RawMessage {
	switch v := in.(type) {
	case json.RawMessage:
		if len(v) == 0 {
			return json.RawMessage("{}")
		}
		if json.Valid(v) {
			return append(json.RawMessage(nil), v...)
		}
		b, _ := json.Marshal(map[string]any{"raw": string(v)})
		return b
	case []byte:
		if len(v) == 0 {
			return json.RawMessage("{}")
		}
		if json.Valid(v) {
			return append(json.RawMessage(nil), v...)
		}
		b, _ := json.Marshal(map[string]any{"raw": string(v)})
		return b
	case string:
		if raw, ok := rawJSONOrString(v).(json.RawMessage); ok {
			return raw
		}
		b, _ := json.Marshal(rawJSONOrString(v))
		return b
	default:
		b, err := json.Marshal(in)
		if err == nil {
			return b
		}
		b, _ = json.Marshal(map[string]any{"raw": fmt.Sprint(in)})
		return b
	}
}

// newAssistantMsg creates an internal assistant-role msgBuilder.
func newAssistantMsg() *msgBuilder {
	return &msgBuilder{msg: &message{Role: "assistant"}}
}

// newAssistantTextMsg creates an assistant message containing one text block.
func newAssistantTextMsg(text string) xai.MsgBuilder {
	mb := newAssistantMsg()
	if text != "" {
		mb.msg.Contents = append(mb.msg.Contents, &content{Type: contentText, Text: text})
	}
	return mb
}

// appendTextContent appends a non-empty text block to message content.
func appendTextContent(msg *message, text string) {
	if text == "" {
		return
	}
	msg.Contents = append(msg.Contents, &content{Type: contentText, Text: text})
}

// appendThinkingContent appends a non-empty thinking block.
func appendThinkingContent(msg *message, v xai.Thinking) {
	if v.Text == "" && v.Signature == "" {
		return
	}
	msg.Contents = append(msg.Contents, &content{
		Type: contentThinking,
		Thinking: &thinkingContent{
			Text:      v.Text,
			Signature: v.Signature,
			Redacted:  v.Redacted,
		},
	})
}

// appendToolUseContent appends a tool-use block, normalizing input to JSON.
func appendToolUseContent(msg *message, v xai.ToolUse) {
	msg.Contents = append(msg.Contents, &content{
		Type: contentToolUse,
		ToolUse: &toolUseContent{
			ID:    v.ID,
			Name:  v.Name,
			Input: marshalToolInput(v.Input),
		},
	})
}

// appendCompactionContent appends compaction data when provided.
func appendCompactionContent(msg *message, v xai.Compaction) {
	if v.Data == "" {
		return
	}
	msg.Contents = append(msg.Contents, &content{Type: contentCompaction, Compaction: v.Data})
}

// v1StopReason maps v1 finish_reason to xai.StopReason.
func v1StopReason(reason string) xai.StopReason {
	switch reason {
	case "stop":
		return xai.EndTurn
	case "length":
		return xai.StopMaxTokens
	case "content_filter":
		return xai.Refusal
	case "tool_calls", "function_call":
		return xai.PauseTurn
	}
	return xai.Unspecified
}

// buildToolNameFromType converts a v3 output item type into an internal std/*
// tool name fallback.
func buildToolNameFromType(typ string) string {
	name := strings.TrimSuffix(typ, "_call")
	if name == typ {
		name = strings.TrimSuffix(typ, "_output")
	}
	if name == "" {
		name = "tool"
	}
	return "std/" + name
}

// v3ToolName resolves the internal tool name, preferring explicit names and
// mapping known built-ins to xai standard constants.
func v3ToolName(typ, explicit string) string {
	if explicit != "" {
		return explicit
	}
	switch typ {
	case "web_search_call":
		return xai.ToolWebSearch
	case "code_interpreter_call":
		return xai.ToolCodeExecution
	case "local_shell_call", "shell_call", "shell_call_output":
		return xai.ToolBashCodeExecution
	}
	return buildToolNameFromType(typ)
}

// v3OutputMessageText concatenates textual message content parts in order.
//
// Both output_text and refusal parts are included; empty segments are skipped.
// Multiple segments are joined by newline to avoid accidental token gluing.
func v3OutputMessageText(contents []responses.ResponseOutputMessageContentUnion) string {
	if len(contents) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, item := range contents {
		var seg string
		switch item.Type {
		case "output_text":
			seg = item.Text
		case "refusal":
			seg = item.Refusal
		}
		if seg == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(seg)
	}
	return sb.String()
}

// v3ReasoningText extracts reasoning text from full reasoning content; if empty,
// it falls back to reasoning summary text.
func v3ReasoningText(u responses.ResponseReasoningItem) string {
	var sb strings.Builder
	for _, item := range u.Content {
		if item.Text == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(item.Text)
	}
	if sb.Len() > 0 {
		return sb.String()
	}
	for _, item := range u.Summary {
		if item.Text == "" {
			continue
		}
		if sb.Len() > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(item.Text)
	}
	return sb.String()
}

// isV3HostToolCallType returns true for tool call items that require a host/tool
// executor turn before generation can continue.
func isV3HostToolCallType(typ string) bool {
	switch typ {
	case "function_call", "custom_tool_call", "computer_call", "code_interpreter_call",
		"local_shell_call", "shell_call", "apply_patch_call", "mcp_approval_request":
		return true
	}
	return false
}

// hasV3Compaction checks whether output contains a compaction instruction.
func hasV3Compaction(items []responses.ResponseOutputItemUnion) bool {
	for _, item := range items {
		if item.Type == "compaction" {
			return true
		}
	}
	return false
}

// hasV3HostToolUse checks whether output contains host-managed tool requests.
func hasV3HostToolUse(items []responses.ResponseOutputItemUnion) bool {
	for _, item := range items {
		if isV3HostToolCallType(item.Type) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// V3 Response (Responses API)
// -----------------------------------------------------------------------------

// v3ContentBlock wraps one Responses API output item and implements xai.Part.
//
// Usage:
//
//	part := resp.At(0).Part(i)
//	if tu, ok := part.AsToolUse(); ok { ... }
//	if txt := part.Text(); txt != "" { ... }
type v3ContentBlock struct {
	content *responses.ResponseOutputItemUnion
}

// AsThinking maps `reasoning` items to xai.Thinking.
//
// Signature carries encrypted_content when present. Redacted is true when
// encrypted content exists but text is absent.
func (p v3ContentBlock) AsThinking() (ret xai.Thinking, ok bool) {
	if p.content == nil || p.content.Type != "reasoning" {
		return
	}
	u := p.content.AsReasoning()
	ret.Text = v3ReasoningText(u)
	ret.Signature = u.EncryptedContent
	ret.Redacted = ret.Text == "" && ret.Signature != ""
	ret.Underlying = &u
	return ret, true
}

// AsToolUse maps tool-call-like output items to xai.ToolUse.
//
// It covers function/custom tools and built-in calls that require external
// execution before the next model turn.
func (p v3ContentBlock) AsToolUse() (ret xai.ToolUse, ok bool) {
	if p.content == nil {
		return
	}
	switch p.content.Type {
	case "function_call":
		u := p.content.AsFunctionCall()
		ret.ID = coalesce(u.CallID, u.ID)
		ret.Name = v3ToolName(p.content.Type, u.Name)
		ret.Input = rawJSONOrString(u.Arguments)
		ret.Underlying = &u
	case "custom_tool_call":
		u := p.content.AsCustomToolCall()
		ret.ID = coalesce(u.CallID, u.ID)
		ret.Name = v3ToolName(p.content.Type, u.Name)
		ret.Input = rawJSONOrString(u.Input)
		ret.Underlying = &u
	case "computer_call":
		u := p.content.AsComputerCall()
		ret.ID = coalesce(u.CallID, u.ID)
		ret.Name = v3ToolName(p.content.Type, "")
		ret.Input = map[string]any{
			"action":                u.Action,
			"pending_safety_checks": u.PendingSafetyChecks,
		}
		ret.Underlying = &u
	case "code_interpreter_call":
		u := p.content.AsCodeInterpreterCall()
		ret.ID = u.ID
		ret.Name = v3ToolName(p.content.Type, "")
		ret.Input = map[string]any{
			"code":         u.Code,
			"container_id": u.ContainerID,
		}
		ret.Underlying = &u
	case "local_shell_call":
		u := p.content.AsLocalShellCall()
		ret.ID = coalesce(u.CallID, u.ID)
		ret.Name = v3ToolName(p.content.Type, "")
		ret.Input = map[string]any{"action": u.Action}
		ret.Underlying = &u
	case "shell_call":
		u := p.content.AsShellCall()
		ret.ID = coalesce(u.CallID, u.ID)
		ret.Name = v3ToolName(p.content.Type, "")
		ret.Input = map[string]any{
			"action":      u.Action,
			"environment": u.Environment,
		}
		ret.Underlying = &u
	case "apply_patch_call":
		u := p.content.AsApplyPatchCall()
		ret.ID = coalesce(u.CallID, u.ID)
		ret.Name = v3ToolName(p.content.Type, "")
		ret.Input = map[string]any{"operation": u.Operation}
		ret.Underlying = &u
	case "mcp_approval_request":
		u := p.content.AsMcpApprovalRequest()
		ret.ID = u.ID
		ret.Name = v3ToolName(p.content.Type, u.Name)
		ret.Input = map[string]any{
			"arguments":    rawJSONOrString(u.Arguments),
			"server_label": u.ServerLabel,
		}
		ret.Underlying = &u
	default:
		return
	}
	return ret, true
}

// AsToolResult maps tool-result-like output items to xai.ToolResult.
//
// This includes web/file search results, shell/apply-patch outputs, and MCP
// result envelopes.
func (p v3ContentBlock) AsToolResult() (ret xai.ToolResult, ok bool) {
	if p.content == nil {
		return
	}
	switch p.content.Type {
	case "web_search_call":
		u := p.content.AsWebSearchCall()
		ret.ID = u.ID
		ret.Name = xai.ToolWebSearch
		ret.Result = &u
		ret.Underlying = &u
	case "file_search_call":
		u := p.content.AsFileSearchCall()
		ret.ID = u.ID
		ret.Name = v3ToolName(p.content.Type, "")
		ret.Result = &u
		ret.Underlying = &u
	case "code_interpreter_call":
		u := p.content.AsCodeInterpreterCall()
		ret.ID = u.ID
		ret.Name = xai.ToolCodeExecution
		ret.Result = &u
		ret.IsError = string(u.Status) == "failed"
		ret.Underlying = &u
	case "shell_call_output":
		u := p.content.AsShellCallOutput()
		ret.ID = u.CallID
		ret.Name = v3ToolName(p.content.Type, "")
		ret.Result = &u
		ret.Underlying = &u
	case "apply_patch_call_output":
		u := p.content.AsApplyPatchCallOutput()
		ret.ID = u.CallID
		ret.Name = v3ToolName(p.content.Type, "")
		if u.Output != "" {
			ret.Result = u.Output
		} else {
			ret.Result = &u
		}
		ret.IsError = string(u.Status) == "failed"
		ret.Underlying = &u
	case "mcp_call":
		u := p.content.AsMcpCall()
		ret.ID = u.ID
		ret.Name = v3ToolName(p.content.Type, u.Name)
		if u.Error != "" {
			ret.IsError = true
			ret.Result = errors.New(u.Error)
		} else if u.Output != "" {
			ret.Result = rawJSONOrString(u.Output)
		} else {
			ret.Result = &u
		}
		ret.Underlying = &u
	case "mcp_list_tools":
		u := p.content.AsMcpListTools()
		ret.ID = u.ID
		ret.Name = v3ToolName(p.content.Type, "")
		if u.Error != "" {
			ret.IsError = true
			ret.Result = errors.New(u.Error)
		} else {
			ret.Result = &u
		}
		ret.Underlying = &u
	default:
		return
	}
	return ret, true
}

// AsBlob maps image_generation_call result (base64) to xai.Blob.
func (p v3ContentBlock) AsBlob() (ret xai.Blob, ok bool) {
	if p.content == nil || p.content.Type != "image_generation_call" {
		return
	}
	u := p.content.AsImageGenerationCall()
	if u.Result == "" {
		return
	}
	ret.MIME = "image/png"
	ret.BlobData = xai.BlobFromBase64(u.Result)
	return ret, true
}

// AsCompaction maps compaction items to xai.Compaction.
func (p v3ContentBlock) AsCompaction() (ret xai.Compaction, ok bool) {
	if p.content == nil || p.content.Type != "compaction" {
		return
	}
	u := p.content.AsCompaction()
	ret.Data = u.EncryptedContent
	return ret, true
}

// Text extracts output message text/refusal content from `message` items.
func (p v3ContentBlock) Text() string {
	if p.content == nil || p.content.Type != "message" {
		return ""
	}
	return v3OutputMessageText(p.content.Content)
}

// Underlying returns the raw SDK union pointer.
func (p v3ContentBlock) Underlying() any {
	return p.content
}

// v3ContentBlockFromText creates a synthetic message part used by streaming
// chunks where only text delta is available.
func v3ContentBlockFromText(text string) v3ContentBlock {
	return v3ContentBlock{content: &responses.ResponseOutputItemUnion{
		Type: "message",
		Content: []responses.ResponseOutputMessageContentUnion{
			{Type: "output_text", Text: text},
		},
	}}
}

// v3Response wraps a full Responses API response and implements both
// xai.GenResponse and xai.Candidate for a single output candidate.
type v3Response struct {
	msg *responses.Response
}

// StopReason computes a high-level stop reason with precedence:
//  1. compaction request
//  2. host tool call pause
//  3. response status / incomplete details
func (p *v3Response) StopReason() xai.StopReason {
	if p.msg == nil {
		return xai.Unspecified
	}
	if hasV3Compaction(p.msg.Output) {
		return xai.StopCompaction
	}
	if hasV3HostToolUse(p.msg.Output) {
		return xai.PauseTurn
	}
	switch string(p.msg.Status) {
	case "completed":
		return xai.EndTurn
	case "incomplete":
		switch p.msg.IncompleteDetails.Reason {
		case "max_output_tokens":
			return xai.StopMaxTokens
		case "content_filter":
			return xai.Refusal
		}
	case "failed":
		if p.msg.Error.Message != "" {
			return xai.Refusal
		}
	}
	return xai.Unspecified
}

// Parts returns number of output items in the response.
func (p *v3Response) Parts() int {
	if p.msg == nil {
		return 0
	}
	return len(p.msg.Output)
}

// Part returns output item i as xai.Part.
func (p *v3Response) Part(i int) xai.Part {
	n := p.Parts()
	if i < 0 || i >= n {
		panicIndex("v3Response.Part", i, n)
	}
	return v3ContentBlock{content: &p.msg.Output[i]}
}

// Len returns number of candidates (always 1 for v3 wrapper).
func (p *v3Response) Len() int {
	if p.msg == nil {
		return 0
	}
	return 1
}

// At returns candidate i (only index 0 is valid).
func (p *v3Response) At(i int) xai.Candidate {
	if i != 0 || p.Len() == 0 {
		panicIndex("v3Response.At", i, p.Len())
	}
	return p
}

// ToMsg converts response output into an assistant message that can be appended
// to conversation history.
//
// Usage:
//
//	history = append(history, resp.At(0).ToMsg())
func (p *v3Response) ToMsg() xai.MsgBuilder {
	mb := newAssistantMsg()
	if p.msg == nil {
		return mb
	}
	for _, item := range p.msg.Output {
		switch item.Type {
		case "message":
			appendTextContent(mb.msg, v3OutputMessageText(item.Content))
		case "reasoning":
			reasoning := item.AsReasoning()
			appendThinkingContent(mb.msg, xai.Thinking{
				Text:       v3ReasoningText(reasoning),
				Signature:  reasoning.EncryptedContent,
				Redacted:   reasoning.EncryptedContent != "" && len(reasoning.Content) == 0,
				Underlying: &reasoning,
			})
		case "function_call":
			call := item.AsFunctionCall()
			appendToolUseContent(mb.msg, xai.ToolUse{
				ID:         coalesce(call.CallID, call.ID),
				Name:       v3ToolName(item.Type, call.Name),
				Input:      rawJSONOrString(call.Arguments),
				Underlying: &call,
			})
		case "custom_tool_call":
			call := item.AsCustomToolCall()
			appendToolUseContent(mb.msg, xai.ToolUse{
				ID:         coalesce(call.CallID, call.ID),
				Name:       v3ToolName(item.Type, call.Name),
				Input:      rawJSONOrString(call.Input),
				Underlying: &call,
			})
		case "compaction":
			compaction := item.AsCompaction()
			appendCompactionContent(mb.msg, xai.Compaction{Data: compaction.EncryptedContent})
		case "image_generation_call":
			img := item.AsImageGenerationCall()
			if img.Result != "" {
				mb.msg.Contents = append(mb.msg.Contents, &content{
					Type:     contentImageURL,
					ImageURL: "data:image/png;base64," + img.Result,
				})
			}
		}
	}
	return mb
}

// v3StreamChunk is the minimal wrapper used for Responses API stream deltas.
//
// Each chunk contains exactly one text part and does not carry stop reason.
type v3StreamChunk struct {
	text string
}

func (p *v3StreamChunk) StopReason() xai.StopReason { return xai.Unspecified }
func (p *v3StreamChunk) Parts() int                 { return 1 }
func (p *v3StreamChunk) Part(i int) xai.Part {
	if i != 0 {
		panicIndex("v3StreamChunk.Part", i, 1)
	}
	return v3ContentBlockFromText(p.text)
}
func (p *v3StreamChunk) Len() int { return 1 }
func (p *v3StreamChunk) At(i int) xai.Candidate {
	if i != 0 {
		panicIndex("v3StreamChunk.At", i, 1)
	}
	return p
}

// ToMsg converts the delta text into an assistant message.
func (p *v3StreamChunk) ToMsg() xai.MsgBuilder { return newAssistantTextMsg(p.text) }

// -----------------------------------------------------------------------------
// Shared
// -----------------------------------------------------------------------------

// streamError represents an error event carried through stream iterator output.
type streamError struct{ msg string }

func (e *streamError) Error() string { return e.msg }

// -----------------------------------------------------------------------------
