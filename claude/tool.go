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
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/goplus/xai"
)

// -----------------------------------------------------------------------------

func (p *contentBuilder) SearchResult(content xai.TextBuilder, source, title string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaSearchResultBlock(buildTexts(content), source, title))
	return p
}

func (p *contentBuilder) ToolUse(id string, input any, name string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaToolUseBlock(id, input, name))
	return p
}

func (p *contentBuilder) ToolResult(toolUseID string, content any, isError bool) xai.ContentBuilder {
	// TODO(xsw): validate content
	p.content = append(p.content, anthropic.NewBetaToolResultBlock(toolUseID, content.(string), isError))
	return p
}

func (p *contentBuilder) ServerToolUse(id string, input any, name xai.ServerToolName) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaServerToolUseBlock(id, input, anthropic.BetaServerToolUseBlockParamName(name)))
	return p
}

// -----------------------------------------------------------------------------
