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
	"strings"
	"unsafe"

	"github.com/goplus/xai"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
)

// -----------------------------------------------------------------------------

type tools map[string]*responses.FunctionToolParam

type tool struct {
	tool *responses.FunctionToolParam
}

func (p tool) Description(desc string) xai.Tool {
	p.tool.Description = param.NewOpt(desc)
	return p
}

func (p *Provider) ToolIsDefined(name string) bool {
	_, ok := p.tools[name]
	return ok
}

func (p *Provider) ToolDef(name string) xai.Tool {
	if p.ToolIsDefined(name) {
		panic("tool already defined: " + name)
	}
	ret := &responses.FunctionToolParam{Name: name}
	p.tools[name] = ret
	return tool{ret}
}

func buildTools(tools tools, params []any) []responses.ToolUnionParam {
	ret := make([]responses.ToolUnionParam, len(params))
	for i, v := range params {
		var param responses.ToolUnionParam
		if name, ok := v.(string); ok {
			tool, ok := tools[name]
			if !ok {
				panic("undefined tool: " + name)
			}
			param.OfFunction = tool
		} else {
			v.(xai.StdTool).UnderlyingAssignTo(&param)
		}
		ret[i] = param
	}
	return ret
}

// -----------------------------------------------------------------------------

type webSearchTool struct {
	param *responses.WebSearchToolParam
}

func (p webSearchTool) UnderlyingAssignTo(ret any) {
	ret.(*responses.ToolUnionParam).OfWebSearch = p.param
}

func (p webSearchTool) MaxUses(v int64) xai.WebSearchTool {
	// openai web search tool does not support max uses
	return p
}

func (p webSearchTool) AllowedDomains(v ...string) xai.WebSearchTool {
	p.param.Filters.AllowedDomains = v
	return p
}

func (p webSearchTool) BlockedDomains(v ...string) xai.WebSearchTool {
	// openai web search tool does not support blocked domains
	return p
}

func (p *Provider) WebSearchTool() xai.WebSearchTool {
	return webSearchTool{&responses.WebSearchToolParam{
		Type: "web_search_2025_08_26",
	}}
}

// -----------------------------------------------------------------------------

func (p *contentBuilder) ToolUse(toolID, name string, input any) xai.ContentBuilder {
	var (
		content responses.ResponseInputItemUnionParam
	)
	if strings.HasPrefix(name, "std/") {
		panic("todo")
	} else {
		args := jsonStringify(input, "invalid tool input: ")
		content = responses.ResponseInputItemParamOfFunctionCall(toolID, args, name)
	}
	return p.addNonMsg(content)
}

func jsonStringify(v any, errPrompt string) string {
	var args []byte
	if v, ok := v.(json.RawMessage); ok {
		args = []byte(v)
	} else {
		var err error
		args, err = json.Marshal(v)
		if err != nil {
			panic(errPrompt + err.Error())
		}
	}
	return unsafe.String(unsafe.SliceData(args), len(args))
}

// -----------------------------------------------------------------------------

var stdToolResultConv = map[string]func(toolID string, result any, isError bool) responses.ResponseInputItemUnionParam{
	xai.ToolWebSearch: webSearchResultConv,
}

func webSearchResultConv(toolID string, result any, isError bool) responses.ResponseInputItemUnionParam {
	panic("todo")
}

func (p *contentBuilder) ToolResult(toolID, name string, result any, isError bool) xai.ContentBuilder {
	var (
		content responses.ResponseInputItemUnionParam
	)
	if strings.HasPrefix(name, "std/") {
		conv, ok := stdToolResultConv[name]
		if !ok {
			panic("unsupported standard tool: " + name)
		}
		content = conv(toolID, result, isError)
	} else {
		if isError {
			result = map[string]any{"error": result.(error).Error()}
		}
		ret := jsonStringify(result, "invalid tool result: ")
		content = responses.ResponseInputItemParamOfFunctionCallOutput(toolID, ret)
	}
	p.content = append(p.content, content)
	return p
}

// -----------------------------------------------------------------------------
