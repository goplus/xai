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

package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"

	oshared "github.com/goplus/xai/examples/openai/shared"
)

func runChatTool() {
	service := oshared.NewService("")
	ctx := context.Background()

	weatherTool := service.ToolDef("get_weather").Description("Get weather by city")
	firstParams := service.Params().
		Model(xai.Model("gemini-2.5-flash-image")).
		Tools(weatherTool).
		Messages(service.UserMsg().Text("上海今天天气如何？如果能查天气请调用工具"))

	firstResp, err := service.Gen(ctx, firstParams, oshared.DebugOptions(service, nil))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	oshared.PrintResponseBlocksWithTitle("first_response", firstResp)

	if firstResp.Len() == 0 || firstResp.At(0).Parts() == 0 {
		return
	}
	toolUse, ok := firstResp.At(0).Part(0).AsToolUse()
	if !ok {
		fmt.Println("No tool call returned")
		return
	}

	toolResult := map[string]any{
		"city":        "Shanghai",
		"temperature": "26C",
		"condition":   "Sunny",
	}

	finalParams := service.Params().
		Model(xai.Model("gemini-2.5-flash-image")).
		Tools(weatherTool).
		Messages(
			service.UserMsg().Text("上海今天天气如何？如果能查天气请调用工具"),
			firstResp.At(0).ToMsg(),
			service.UserMsg().ToolResult(xai.ToolResult{
				ID:     toolUse.ID,
				Name:   toolUse.Name,
				Result: toolResult,
			}),
		)

	finalResp, err := service.Gen(ctx, finalParams, oshared.DebugOptions(service, nil))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	oshared.PrintResponseBlocksWithTitle("final_response", finalResp)
}
