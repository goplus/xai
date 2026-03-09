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

// Chat completions: function calling (tool use + tool result round-trip).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	xai "github.com/goplus/xai/spec"

	"github.com/goplus/xai/examples/openai/shared"
)

const weatherToolName = "get_weather"

type weatherToolArgs struct {
	City string `json:"city"`
	Unit string `json:"unit"`
}

type weatherToolResult struct {
	City        string  `json:"city"`
	Unit        string  `json:"unit"`
	Temperature float64 `json:"temperature"`
	Condition   string  `json:"condition"`
	Suggestion  string  `json:"suggestion"`
	Source      string  `json:"source"`
}

func runChatFunctionCall() {
	svc := shared.NewService("")
	ctx := context.Background()

	if shared.StreamMode() {
		fmt.Println("function-call demo uses non-stream requests to complete the tool loop.")
	}

	svc.ToolDef(weatherToolName).Description(
		`Get weather for a city. Input JSON: {"city":"city name","unit":"celsius|fahrenheit"}.`,
	)

	userPrompt := "Please call get_weather for Shanghai, then give one short outfit suggestion."
	firstResp, err := svc.Gen(ctx, svc.Params().
		Model(xai.Model(shared.ModelGeminiPro)).
		Messages(svc.UserMsg().Text(userPrompt)).
		Tools(svc.Tool(weatherToolName)), nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if firstResp == nil || firstResp.Len() == 0 {
		fmt.Println("Error: empty response")
		return
	}
	shared.PrintResponseBlocksWithTitle("first_response", firstResp)

	toolCalls := normalizeToolCalls(collectToolCalls(firstResp.At(0)))
	if len(toolCalls) == 0 {
		fmt.Println("No tool call returned in first_response.")
		return
	}

	followup := []xai.MsgBuilder{svc.UserMsg().Text(userPrompt)}
	assistantCallMsg := svc.AssistantMsg()
	for _, call := range toolCalls {
		assistantCallMsg = assistantCallMsg.ToolUse(call)
	}
	followup = append(followup, assistantCallMsg)
	for _, call := range toolCalls {
		followup = append(followup, buildWeatherToolResultMsg(svc, call))
	}

	finalResp, err := svc.Gen(ctx, svc.Params().
		Model(xai.Model(shared.ModelGeminiPro)).
		Messages(followup...).
		Tools(svc.Tool(weatherToolName)), nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if finalResp == nil || finalResp.Len() == 0 {
		fmt.Println("Error: empty final response")
		return
	}
	shared.PrintResponseBlocksWithTitle("final_response", finalResp)
}

func collectToolCalls(cand xai.Candidate) []xai.ToolUse {
	var calls []xai.ToolUse
	for i := 0; i < cand.Parts(); i++ {
		if call, ok := cand.Part(i).AsToolUse(); ok {
			calls = append(calls, call)
		}
	}
	return calls
}

func normalizeToolCalls(calls []xai.ToolUse) []xai.ToolUse {
	for i := range calls {
		if strings.TrimSpace(calls[i].ID) == "" {
			calls[i].ID = fmt.Sprintf("toolcall-%d", i+1)
		}
	}
	return calls
}

func buildWeatherToolResultMsg(svc xai.Service, call xai.ToolUse) xai.MsgBuilder {
	if call.Name != weatherToolName {
		return svc.AssistantMsg().ToolResult(xai.ToolResult{
			ID:      call.ID,
			Name:    call.Name,
			Result:  fmt.Errorf("unsupported tool: %s", call.Name),
			IsError: true,
		})
	}

	var args weatherToolArgs
	if err := decodeToolInput(call.Input, &args); err != nil {
		return svc.AssistantMsg().ToolResult(xai.ToolResult{
			ID:      call.ID,
			Name:    call.Name,
			Result:  fmt.Errorf("invalid tool args: %w", err),
			IsError: true,
		})
	}

	ret, err := mockWeatherTool(args)
	if err != nil {
		return svc.AssistantMsg().ToolResult(xai.ToolResult{
			ID:      call.ID,
			Name:    call.Name,
			Result:  err,
			IsError: true,
		})
	}
	return svc.AssistantMsg().ToolResult(xai.ToolResult{
		ID:     call.ID,
		Name:   call.Name,
		Result: ret,
	})
}

func decodeToolInput(input any, out any) error {
	switch v := input.(type) {
	case json.RawMessage:
		if len(v) == 0 {
			return fmt.Errorf("empty tool input")
		}
		return json.Unmarshal(v, out)
	case []byte:
		if len(v) == 0 {
			return fmt.Errorf("empty tool input")
		}
		return json.Unmarshal(v, out)
	case string:
		if strings.TrimSpace(v) == "" {
			return fmt.Errorf("empty tool input")
		}
		return json.Unmarshal([]byte(v), out)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if len(b) == 0 {
			return fmt.Errorf("empty tool input")
		}
		return json.Unmarshal(b, out)
	}
}

func mockWeatherTool(args weatherToolArgs) (weatherToolResult, error) {
	city := strings.TrimSpace(args.City)
	if city == "" {
		return weatherToolResult{}, fmt.Errorf("city is required")
	}

	unit := strings.ToLower(strings.TrimSpace(args.Unit))
	if unit == "" {
		unit = "celsius"
	}
	if unit != "celsius" && unit != "fahrenheit" {
		return weatherToolResult{}, fmt.Errorf("unsupported unit: %s", args.Unit)
	}

	var (
		tempC      = 22.0
		condition  = "clear"
		suggestion = "A T-shirt with a light jacket should work well."
	)

	switch strings.ToLower(city) {
	case "shanghai":
		tempC = 18
		condition = "cloudy"
		suggestion = "Wear a light jacket, especially in the morning and evening."
	case "beijing":
		tempC = 11
		condition = "partly cloudy"
		suggestion = "Wear a warmer outer layer since it may feel cool."
	case "shenzhen":
		tempC = 25
		condition = "light rain"
		suggestion = "Wear a T-shirt with a light waterproof jacket and carry an umbrella."
	}

	temperature := tempC
	if unit == "fahrenheit" {
		temperature = tempC*9/5 + 32
	}

	return weatherToolResult{
		City:        city,
		Unit:        unit,
		Temperature: temperature,
		Condition:   condition,
		Suggestion:  suggestion,
		Source:      "mock-local-tool",
	}, nil
}
