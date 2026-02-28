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
	"encoding/json"
	"strings"
	"unsafe"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
	"github.com/goplus/xai"
)

// -----------------------------------------------------------------------------

func (p *contentBuilder) ToolUse(toolID, name string, input any) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaToolUseBlock(toolID, input, name))
	return p
}

/* TODO(xsw):
func (p *contentBuilder) serverToolUse(toolID string, input any, name anthropic.BetaServerToolUseBlockParamName) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaServerToolUseBlock(toolID, input, name))
	return p
}
*/

// -----------------------------------------------------------------------------

var stdToolResultConv = map[string]func(toolID string, result any, isError bool) anthropic.BetaContentBlockParamUnion{
	xai.ToolWebSearch: webSearchResultConv,
}

func webSearchResultConv(toolID string, result any, isError bool) anthropic.BetaContentBlockParamUnion {
	if isError {
		v := result.(error)
		return anthropic.NewBetaWebSearchToolResultBlock(anthropic.BetaWebSearchToolRequestErrorParam{
			ErrorCode: anthropic.BetaWebSearchToolResultErrorCode(v.Error()),
		}, toolID)
	}
	v := result.(*xai.WebSearchResult)
	ret := make([]anthropic.BetaWebSearchResultBlockParam, len(v.Result))
	for i, item := range v.Result {
		ret[i] = anthropic.BetaWebSearchResultBlockParam{
			EncryptedContent: item.Underlying.(string),
			Title:            item.Title,
			URL:              item.URL,
		}
		if item.PageAge != "" {
			ret[i].PageAge = param.NewOpt(item.PageAge)
		}
	}
	return anthropic.NewBetaWebSearchToolResultBlock(ret, toolID)
}

/* TODO(xsw): SearchResult vs. WebSearchResult
func (p *contentBuilder) searchResult(content xai.TextBuilder, source, title string) xai.ContentBuilder {
	p.content = append(p.content, anthropic.NewBetaSearchResultBlock(buildTexts(content), source, title))
	return p
}
*/

func (p *contentBuilder) ToolResult(toolID, name string, result any, isError bool) xai.ContentBuilder {
	var (
		content anthropic.BetaContentBlockParamUnion
	)
	if strings.HasPrefix(name, "std/") {
		conv, ok := stdToolResultConv[name]
		if !ok {
			panic("unsupported standard tool: " + name)
		}
		content = conv(toolID, result, isError)
	} else {
		var ret string
		if v, ok := result.(xai.RawText); ok {
			ret = string(v)
		} else if isError {
			ret = result.(error).Error()
		} else {
			b, err := json.Marshal(result)
			if err != nil {
				panic("failed to marshal tool result: " + err.Error())
			}
			ret = unsafe.String(unsafe.SliceData(b), len(b))
		}
		content = anthropic.NewBetaToolResultBlock(toolID, ret, isError)
	}
	p.content = append(p.content, content)
	return p
}

// -----------------------------------------------------------------------------
