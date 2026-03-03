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

// all standard tool names start with "std/". The rest of the name should be unique
// and descriptive of the tool's function. For example, "std/web_search" for a web
// search tool, "std/code_execution" for a code execution tool, etc.
const (
	ToolWebSearch               = "std/web_search"
	ToolWebFetch                = "std/web_fetch"
	ToolCodeExecution           = "std/code_execution"
	ToolBashCodeExecution       = "std/bash_code_execution"
	ToolTextEditorCodeExecution = "std/text_editor_code_execution"
	ToolSearchToolRegex         = "std/tool_search_tool_regex"
	ToolSearchToolBm25          = "std/tool_search_tool_bm25"
)

// -----------------------------------------------------------------------------

type WebSearchResultItem struct {
	Title   string
	URL     string
	PageAge string

	// implementation-specific content that can be used for tool result conversion
	Underlying any
}

type WebSearchResult struct {
	Result []WebSearchResultItem
}

// -----------------------------------------------------------------------------

type WebFetchResult struct {
	Content any // TODO(xsw): define a more specific type for this
	Caller  string
}

// -----------------------------------------------------------------------------

type CodeExecutionResult struct {
	ReturnCode int64
	Stderr     string
	Stdout     string
}

// -----------------------------------------------------------------------------

type BashCodeExecutionResult struct {
}

// -----------------------------------------------------------------------------

type TextEditorCodeExecutionResult struct {
}

// -----------------------------------------------------------------------------

type SearchToolRegexResult struct {
}

// -----------------------------------------------------------------------------

type SearchToolBm25Result struct {
}

// -----------------------------------------------------------------------------

type StdTool interface {
	UnderlyingAssignTo(any) // don't call it directly
}

type WebSearchTool interface {
	StdTool

	MaxUses(int64) WebSearchTool
	AllowedDomains(...string) WebSearchTool
	BlockedDomains(...string) WebSearchTool
}

// -----------------------------------------------------------------------------

type Tool interface {
	Description(string) Tool
}

// -----------------------------------------------------------------------------
