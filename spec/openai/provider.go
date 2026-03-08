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
	"encoding/json"
	"iter"

	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/option"
)

// -----------------------------------------------------------------------------

// provider is the internal interface for different OpenAI API versions.
type provider interface {
	Gen(ctx context.Context, req *genRequest, opts []option.RequestOption) (genResponse, error)
	GenStream(ctx context.Context, req *genRequest, opts []option.RequestOption) iter.Seq2[genResponse, error]
	Features() xai.Feature
}

// genResponse is the internal response interface.
type genResponse interface {
	xai.GenResponse
}

// -----------------------------------------------------------------------------
// Internal request types
// -----------------------------------------------------------------------------

// genRequest is the internal unified request format.
type genRequest struct {
	Model           string
	System          []textContent
	Messages        []*message
	Tools           []*toolDef
	MaxOutputTokens int64
	Temperature     float64
	TopP            float64
	HasTemperature  bool
	HasTopP         bool
}

// textContent represents a text content item.
type textContent struct {
	Text string
}

// message is the internal unified message format.
type message struct {
	Role     string // "user", "assistant", "system", "tool"
	Contents []*content
}

// contentType defines the type of content.
type contentType int

const (
	contentText contentType = iota
	contentImageURL
	contentImageFile
	contentDocURL
	contentDocFile
	contentToolUse
	contentToolResult
	contentThinking
	contentCompaction
)

// content is the internal unified content format.
type content struct {
	Type        contentType
	Text        string
	ImageURL    string
	ImageDetail string
	FileID      string
	FileURL     string
	FileMIME    string
	ToolUse     *toolUseContent
	ToolResult  *toolResultContent
	Thinking    *thinkingContent
	Compaction  string
}

// toolUseContent represents a tool use request.
type toolUseContent struct {
	ID    string
	Name  string
	Input json.RawMessage
}

// toolResultContent represents a tool result.
type toolResultContent struct {
	ID      string
	Name    string
	Result  json.RawMessage
	IsError bool
}

// thinkingContent represents thinking/reasoning content.
type thinkingContent struct {
	Text      string
	Signature string
	Redacted  bool
}

// toolDef is the internal tool definition.
type toolDef struct {
	Name        string
	Description string
	IsWebSearch bool
}

// -----------------------------------------------------------------------------
