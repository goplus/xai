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

// Gemini examples via Qiniu provider.
// Chat mode uses OpenAI-compatible /v1/chat/completions.
// Image operations use /v1/images/generations and /v1/images/edits.
package main

import (
	"fmt"
	"os"

	gshared "github.com/goplus/xai/examples/gemini/shared"
	oshared "github.com/goplus/xai/examples/openai/shared"
)

var demos = map[string]func(){
	"chat-text":      runChatText,
	"chat-image":     runChatImage,
	"chat-tool":      runChatTool,
	"image-generate": runImageGenerate,
	"image-edit":     runImageEdit,
}

var demoOrder = []string{
	"chat-text",
	"chat-image",
	"chat-tool",
	"image-generate",
	"image-edit",
}

func main() {
	args := parseDemoArgs(os.Args[1:])
	if len(args) == 0 {
		fmt.Println("Gemini (Qiniu provider) examples:")
		fmt.Println("  Set QINIU_API_KEY for real API calls")
		fmt.Println()
		for _, name := range demoOrder {
			fmt.Printf("  %-15s %s\n", name, demoDesc(name))
		}
		fmt.Println()
		fmt.Println("Usage: go run ./examples/gemini [--stream|--no-stream] [demo]")
		return
	}

	for _, arg := range args {
		if fn, ok := demos[arg]; ok {
			fmt.Println("---", arg, "---")
			fn()
		} else {
			fmt.Printf("Unknown demo: %s\nAvailable: %v\n", arg, demoOrder)
		}
	}
}

func parseDemoArgs(args []string) []string {
	stream := gshared.StreamMode()
	demos := make([]string, 0, len(args))
	for _, arg := range args {
		switch arg {
		case "--stream", "-s":
			stream = true
		case "--no-stream":
			stream = false
		default:
			demos = append(demos, arg)
		}
	}
	gshared.SetStreamMode(stream)
	oshared.SetStreamMode(stream)
	return demos
}

func demoDesc(name string) string {
	switch name {
	case "chat-text":
		return "Text-only chat"
	case "chat-image":
		return "Text + image_url chat"
	case "chat-tool":
		return "Tool call round-trip"
	case "image-generate":
		return "Operation GenImage"
	case "image-edit":
		return "Operation EditImage"
	default:
		return ""
	}
}
