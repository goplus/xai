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

	xai "github.com/goplus/xai/spec"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/responses"
)

// -----------------------------------------------------------------------------

type tools map[string]tool

type tool struct {
	tool *responses.FunctionToolParam
}

func (p tool) UnderlyingAssignTo(ret any) {
	ret.(*responses.ToolUnionParam).OfFunction = p.tool
}

func (p tool) Description(desc string) xai.Tool {
	p.tool.Description = param.NewOpt(desc)
	return p
}

func (p *Provider) Tool(name string) xai.Tool {
	return p.tools[name]
}

func (p *Provider) ToolDef(name string) xai.Tool {
	if _, ok := p.tools[name]; ok {
		panic("tool already defined: " + name)
	}
	ret := tool{&responses.FunctionToolParam{Name: name}}
	p.tools[name] = ret
	return ret
}

func buildTools(tools []xai.ToolBase) []responses.ToolUnionParam {
	ret := make([]responses.ToolUnionParam, len(tools))
	for i, v := range tools {
		v.UnderlyingAssignTo(&ret[i])
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

func (p *msgBuilder) ToolUse(v xai.ToolUse) xai.MsgBuilder {
	var (
		content responses.ResponseInputItemUnionParam
	)
	if strings.HasPrefix(v.Name, "std/") {
		panic("todo")
	} else {
		args := jsonStringify(v.Input, "invalid tool input: ")
		content = responses.ResponseInputItemParamOfFunctionCall(v.ID, args, v.Name)
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

func (p *msgBuilder) ToolResult(v xai.ToolResult) xai.MsgBuilder {
	var (
		content responses.ResponseInputItemUnionParam
	)
	if strings.HasPrefix(v.Name, "std/") {
		panic("todo")
	} else {
		if v.IsError {
			v.Result = map[string]any{"error": v.Result.(error).Error()}
		}
		ret := jsonStringify(v.Result, "invalid tool result: ")
		content = responses.ResponseInputItemParamOfFunctionCallOutput(v.ID, ret)
	}
	p.content = append(p.content, content)
	return p
}

// -----------------------------------------------------------------------------
