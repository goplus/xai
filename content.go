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

// -----------------------------------------------------------------------------

type TextBuilder interface {
	Text(text string) TextBuilder
}

// -----------------------------------------------------------------------------

type ServerToolName string

const (
	ToolWebSearch               ServerToolName = "web_search"
	ToolWebFetch                ServerToolName = "web_fetch"
	ToolCodeExecution           ServerToolName = "code_execution"
	ToolBashCodeExecution       ServerToolName = "bash_code_execution"
	ToolTextEditorCodeExecution ServerToolName = "text_editor_code_execution"
	ToolSearchToolRegex         ServerToolName = "tool_search_tool_regex"
	ToolSearchToolBm25          ServerToolName = "tool_search_tool_bm25"
)

type ContentBuilder interface {
	Text(text string) ContentBuilder

	Image(image ImageData) ContentBuilder
	ImageURL(mime ImageType, url string) ContentBuilder
	ImageFile(mime ImageType, fileID string) ContentBuilder

	Doc(doc DocumentData) ContentBuilder
	DocURL(mime DocumentType, url string) ContentBuilder
	DocFile(mime DocumentType, fileID string) ContentBuilder

	SearchResult(content TextBuilder, source, title string) ContentBuilder
	ToolUse(id string, input any, name string) ContentBuilder
	ToolResult(toolUseID string, content any, isError bool) ContentBuilder
	ServerToolUse(id string, input any, name ServerToolName) ContentBuilder

	Thinking(signature, thinking string) ContentBuilder
	RedactedThinking(data string) ContentBuilder
}

// -----------------------------------------------------------------------------

type MessageBuilder interface {
	User(content ContentBuilder) MessageBuilder
	Assistant(content ContentBuilder) MessageBuilder
}

// -----------------------------------------------------------------------------

type Message interface {
	AsContent() ContentBuilder
}

// -----------------------------------------------------------------------------

type ToolBuilder interface {
}

// -----------------------------------------------------------------------------
