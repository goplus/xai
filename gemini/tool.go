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

package gemini

import (
	"encoding/json"
	"strings"
	"unsafe"

	"github.com/goplus/xai"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

func (p *contentBuilder) ToolUse(toolID, name string, input any) xai.ContentBuilder {
	// TODO(xsw): name as toolUseID
	p.content = append(p.content, genai.NewPartFromFunctionCall(name, input.(map[string]any)))
	return p
}

// -----------------------------------------------------------------------------

var stdToolResultConv = map[string]func(toolID string, result any, isError bool) *genai.Part{
	xai.ToolWebSearch: webSearchResultConv,
}

func webSearchResultConv(toolID string, result any, isError bool) *genai.Part {
	// genai.GoogleSearch
	panic("todo")
}

func (p *contentBuilder) ToolResult(toolID, name string, result any, isError bool) xai.ContentBuilder {
	var (
		content *genai.Part
	)
	if strings.HasPrefix(name, "std/") {
		conv, ok := stdToolResultConv[name]
		if !ok {
			panic("unsupported standard tool: " + name)
		}
		content = conv(toolID, result, isError)
	} else {
		var ret map[string]any
		if v, ok := result.(xai.RawText); ok {
			err := json.Unmarshal(unsafe.Slice(unsafe.StringData(string(v)), len(v)), &ret)
			if err != nil {
				panic("invalid tool result: " + err.Error())
			}
		} else if isError {
			ret = map[string]any{"error": result.(error).Error()}
		} else {
			ret = result.(map[string]any)
		}
		content = genai.NewPartFromFunctionResponse(toolID, ret)
	}
	p.content = append(p.content, content)
	return p
}

// -----------------------------------------------------------------------------
