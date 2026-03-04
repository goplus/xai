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

import (
	"encoding/json"
	"io"
)

// -----------------------------------------------------------------------------

type ImageType string

const (
	ImageJPEG ImageType = "image/jpeg"
	ImagePNG  ImageType = "image/png"
	ImageGIF  ImageType = "image/gif"
	ImageWebP ImageType = "image/webp"
)

type DocumentType string

const (
	DocPlainText DocumentType = "text/plain"
	DocPDF       DocumentType = "application/pdf"
)

type ImageData interface {
	ImageType() ImageType
}

type ImageBuilder interface {
	From(mime ImageType, displayName string, src io.Reader) (ImageData, error)
	FromLocal(mime ImageType, fileName string) (ImageData, error)
	FromBase64(mime ImageType, displayName string, base64 string) (ImageData, error)
	FromBytes(mime ImageType, displayName string, data []byte) ImageData
}

type DocumentData interface {
	DocumentType() DocumentType
}

type DocumentBuilder interface {
	From(mime DocumentType, displayName string, src io.Reader) (DocumentData, error)
	FromLocal(mime DocumentType, fileName string) (DocumentData, error)
	FromBase64(mime DocumentType, displayName string, base64 string) (DocumentData, error)
	FromBytes(mime DocumentType, displayName string, data []byte) DocumentData
	PlainText(text string) DocumentData
}

type TextBuilder interface {
	Text(text string) TextBuilder
}

type MsgBuilder interface {
	Text(text string) MsgBuilder

	Image(image ImageData) MsgBuilder
	ImageURL(mime ImageType, url string) MsgBuilder
	ImageFile(mime ImageType, fileID string) MsgBuilder

	Doc(doc DocumentData) MsgBuilder
	DocURL(mime DocumentType, url string) MsgBuilder
	DocFile(mime DocumentType, fileID string) MsgBuilder

	Thinking(Thinking) MsgBuilder

	// ToolUse is used to add a tool use block to the content. The tool ID
	// should be a unique identifier for the tool being used, and should
	// match the ID used in ToolResult.
	//
	// The input expects anything that can be marshaled to JSON, including
	// RawMessage.
	ToolUse(ToolUse) MsgBuilder

	// ToolResult is used to add the result of a tool use to the content.
	// The tool ID should match the ID used in ToolUse. The content depends
	// on the tool. If IsError is true, the content will be treated as an
	// error interface.
	//
	// For standard tools (those with names starting with "std/"), the Result
	// should be a specific struct defined for that tool. For example, the web
	// search tool expects a WebSearchResult struct.
	//
	// For non-standard tools, the content expects anything that can be marshaled
	// to JSON, including RawMessage.
	ToolResult(ToolResult) MsgBuilder
}

type RawMessage = json.RawMessage

// -----------------------------------------------------------------------------

type Thinking struct {
	Text      string
	Signature string // redacted data is saved here, not in Text

	// true if the thinking is redacted, meaning the Text field is empty and
	// the Signature field contains the redacted data.
	Redacted bool

	Underlying any // for provider-specific extensions
}

type ToolUse struct {
	ID   string // tool ID
	Name string // tool Name

	// Arguments for the tool use
	Input any

	Underlying any // for provider-specific extensions
}

type ToolResult struct {
	ID   string // tool ID
	Name string // tool Name

	// The result of the tool use. The content depends on the tool.
	// If IsError is true, the content should be treated as an error interface.
	//
	// For standard tools (those with names starting with "std/"), the Result
	// should be a specific struct defined for that tool. For example, the web
	// search tool expects a WebSearchResult struct.
	Result  any
	IsError bool

	Underlying any // for provider-specific extensions
}

type Part interface {
	AsThinking() (ret Thinking, ok bool)
	AsToolUse() (ret ToolUse, ok bool)
	AsToolResult() (ret ToolResult, ok bool)
	Text() string
	Underlying() any
}

// -----------------------------------------------------------------------------
