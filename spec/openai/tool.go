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
	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

type tools map[string]tool

type tool struct {
	def *toolDef
}

func (p tool) UnderlyingAssignTo(ret any) {
	td := ret.(*toolDef)
	*td = *p.def
}

func (p tool) Description(desc string) xai.Tool {
	p.def.Description = desc
	return p
}

func (p *Service) Tool(name string) xai.Tool {
	return p.tools[name]
}

func (p *Service) ToolDef(name string) xai.Tool {
	if _, ok := p.tools[name]; ok {
		panic("tool already defined: " + name)
	}
	ret := tool{&toolDef{Name: name}}
	p.tools[name] = ret
	return ret
}

// -----------------------------------------------------------------------------

type webSearchTool struct {
	def *toolDef
}

func (p webSearchTool) UnderlyingAssignTo(ret any) {
	td := ret.(*toolDef)
	*td = *p.def
}

func (p webSearchTool) MaxUses(v int64) xai.WebSearchTool {
	return p
}

func (p webSearchTool) AllowedDomains(v ...string) xai.WebSearchTool {
	return p
}

func (p webSearchTool) BlockedDomains(v ...string) xai.WebSearchTool {
	return p
}

func (p *Service) WebSearchTool() xai.WebSearchTool {
	return webSearchTool{&toolDef{IsWebSearch: true}}
}

// -----------------------------------------------------------------------------
