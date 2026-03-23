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

type tools map[string]tool

type tool struct {
	tool *anthropic.BetaToolParam
}

func (p tool) UnderlyingAssignTo(ret any) {
	ret.(*anthropic.BetaToolUnionParam).OfTool = p.tool
}

func (p tool) Description(desc string) xai.Tool {
	p.tool.Description = param.NewOpt(desc)
	return p
}

func (p *Service) Tool(name string) xai.Tool {
	return p.tools[name]
}

func (p *Service) ToolDef(name string) xai.Tool {
	if _, ok := p.tools[name]; ok {
		panic("tool already defined: " + name)
	}
	ret := tool{&anthropic.BetaToolParam{Name: name}}
	p.tools[name] = ret
	return ret
}

func buildTools(tools []xai.ToolBase) []anthropic.BetaToolUnionParam {
	ret := make([]anthropic.BetaToolUnionParam, len(tools))
	for i, v := range tools {
		v.UnderlyingAssignTo(&ret[i])
	}
	return ret
}

// -----------------------------------------------------------------------------

type webSearchTool struct {
	param *anthropic.BetaWebSearchTool20260209Param
}

func (p webSearchTool) UnderlyingAssignTo(ret any) {
	ret.(*anthropic.BetaToolUnionParam).OfWebSearchTool20260209 = p.param
}

func (p webSearchTool) MaxUses(v int64) xai.WebSearchTool {
	p.param.MaxUses = param.NewOpt(v)
	return p
}

func (p webSearchTool) AllowedDomains(v ...string) xai.WebSearchTool {
	p.param.AllowedDomains = v
	return p
}

func (p webSearchTool) BlockedDomains(v ...string) xai.WebSearchTool {
	p.param.BlockedDomains = v
	return p
}

func (p *Service) WebSearchTool() xai.WebSearchTool {
	return webSearchTool{&anthropic.BetaWebSearchTool20260209Param{}}
}

// -----------------------------------------------------------------------------

func (p *msgBuilder) ToolUse(v xai.ToolUse) xai.MsgBuilder {
	var (
		content anthropic.BetaContentBlockParamUnion
	)
	if strings.HasPrefix(v.Name, "std/") {
		stdToolName := anthropic.BetaServerToolUseBlockParamName(v.Name[4:])
		content = anthropic.NewBetaServerToolUseBlock(v.ID, v.Input, stdToolName)
	} else {
		content = anthropic.NewBetaToolUseBlock(v.ID, v.Input, v.Name)
	}
	p.content = append(p.content, content)
	return p
}

// -----------------------------------------------------------------------------

func (p *msgBuilder) ToolResult(v xai.ToolResult) xai.MsgBuilder {
	var (
		content anthropic.BetaContentBlockParamUnion
	)
	if strings.HasPrefix(v.Name, "std/") {
		switch v.Name {
		case xai.ToolWebSearch:
			in := v.Underlying.(*anthropic.BetaWebSearchToolResultBlock).ToParam()
			content = anthropic.BetaContentBlockParamUnion{OfWebSearchToolResult: &in}
		default:
			panic("unsupported standard tool: " + v.Name)
		}
	} else {
		var ret string
		if v.IsError {
			ret = v.Result.(error).Error()
		} else if msg, ok := v.Result.(xai.RawMessage); ok {
			ret = unsafe.String(unsafe.SliceData(msg), len(msg))
		} else {
			b, err := json.Marshal(v.Result)
			if err != nil {
				panic("failed to marshal tool result: " + err.Error())
			}
			ret = unsafe.String(unsafe.SliceData(b), len(b))
		}
		content = anthropic.NewBetaToolResultBlock(v.ID, ret, v.IsError)
	}
	p.content = append(p.content, content)
	return p
}

// -----------------------------------------------------------------------------
