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

import "io"

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

type ContentBuilder interface {
	Text(text string) ContentBuilder

	Image(image ImageData) ContentBuilder
	ImageURL(mime ImageType, url string) ContentBuilder
	ImageFile(mime ImageType, fileID string) ContentBuilder

	Doc(doc DocumentData) ContentBuilder
	DocURL(mime DocumentType, url string) ContentBuilder
	DocFile(mime DocumentType, fileID string) ContentBuilder

	Thinking(signature, thinking string) ContentBuilder
	RedactedThinking(data string) ContentBuilder

	// ToolUse is used to add a tool use block to the content. The toolID
	// should be a unique identifier for the tool being used, and should
	// match the ID used in ToolResult. The input can be any data that the
	// tool requires to perform its function, and the name is a human-readable
	// name for the tool that can be displayed in the UI.
	//
	// For standard tools (those with names starting with "std/"), the input
	// should be a specific struct defined for that tool. For example, the web
	// search tool expects a WebSearchInput struct.
	//
	// For non-standard tools, the input expects a map[string]any or RawText,
	// When the input is a map[string]any, it will be marshaled to JSON and
	// included in the message. When the input is a RawText, it will be treated
	// as a raw string instead of being marshaled to JSON.
	ToolUse(toolID, name string, input any) ContentBuilder

	// ToolResult is used to add the result of a tool use to the content.
	// The toolID should match the ID used in ToolUse. The content depends
	// on the tool. If isError is true, the content will be treated as an
	// error interface.
	//
	// For standard tools (those with names starting with "std/"), the content
	// should be a specific struct defined for that tool. For example, the web
	// search tool expects a WebSearchResult struct.
	//
	// For non-standard tools, the content expects a map[string]any or RawText,
	// When the content is a map[string]any, it will be marshaled to JSON and
	// included in the message. When the content is a RawText, it will be treated
	// as a raw string instead of being marshaled to JSON.
	ToolResult(toolID, name string, result any, isError bool) ContentBuilder
}

// RawText is a wrapper type that indicates the content should be treated as a raw
// string instead of being marshaled to JSON. This is useful for non-standard tool
// inputs or results where the content is already marshaled and does not need to be
// marshaled again.
type RawText string

// -----------------------------------------------------------------------------

type MessageBuilder interface {
	User(content ContentBuilder) MessageBuilder
	Assistant(content ContentBuilder) MessageBuilder
}

type Message interface {
	AsContent() ContentBuilder
}

// -----------------------------------------------------------------------------
