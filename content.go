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

// -----------------------------------------------------------------------------

type ImageType string

const (
	ImageJPEG ImageType = "image/jpeg"
	ImagePNG  ImageType = "image/png"
	ImageGIF  ImageType = "image/gif"
	ImageWebP ImageType = "image/webp"
)

// -----------------------------------------------------------------------------

type TextBuilder interface {
	Text(text string) TextBuilder
}

// -----------------------------------------------------------------------------

type MultipartBuilder interface {
	Text(text string) MultipartBuilder
	ImageURL(url string) MultipartBuilder
	ImageBase64(mime ImageType, base64 string) MultipartBuilder
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
	ImageURL(url string) ContentBuilder
	ImageBase64(mime ImageType, base64 string) ContentBuilder

	DocText(text string) ContentBuilder
	DocPDFURL(url string) ContentBuilder
	DocPDFBase64(base64 string) ContentBuilder
	DocMultipart(multi MultipartBuilder) ContentBuilder

	SearchResult(content TextBuilder, source, title string) ContentBuilder
	ToolUse(id string, input any, name string) ContentBuilder
	ToolResult(toolUseID string, content string, isError bool) ContentBuilder
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
}

// -----------------------------------------------------------------------------

type StreamMessage interface {
}

// -----------------------------------------------------------------------------

type ToolBuilder interface {
}

// -----------------------------------------------------------------------------
