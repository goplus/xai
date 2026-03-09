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

// OpenAI-compatible chat completions examples via Qiniu API (api.qnaigc.com), using provider_v1.
// Run: go run ./examples/openai [demo]
// Set QINIU_API_KEY for real API calls.
package main

import (
	"fmt"
	"os"

	"github.com/goplus/xai/examples/openai/shared"
)

var demos = map[string]func(){
	"text":          runChatText,
	"image":         runChatImage,
	"image-detail":  runChatImageDetailLow,
	"image-ultra":   runChatImageDetailUltraHigh,
	"video":         runChatVideo,
	"video-fileid":  runChatVideoFileID,
	"multi-video":   runChatMultiVideo,
	"function-call": runChatFunctionCall,
	"thinking":      runChatThinking,
}

var demoOrder = []string{
	"text", "image", "image-detail", "image-ultra", "video", "video-fileid", "multi-video", "function-call", "thinking",
}

func main() {
	args := parseDemoArgs(os.Args[1:])
	if len(args) == 0 {
		fmt.Println("OpenAI (Qiniu, provider_v1) chat completions examples:")
		fmt.Println("  Set QINIU_API_KEY for real API calls")
		fmt.Println()
		for _, name := range demoOrder {
			fmt.Printf("  %-16s %s\n", name, demoDesc(name))
		}
		fmt.Println()
		fmt.Println("Usage: go run ./examples/openai [--stream|--no-stream] [demo]")
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
	stream := shared.StreamMode()
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
	shared.SetStreamMode(stream)
	return demos
}

func demoDesc(name string) string {
	switch name {
	case "text":
		return "Text-only: What is the Sun?"
	case "image":
		return "Image + text: What is in this image?"
	case "image-detail":
		return "Image with detail=low"
	case "image-ultra":
		return "Image with detail=ultra_high"
	case "video":
		return "Video URL + text: What is in this video?"
	case "video-fileid":
		return "Video with qfile-xxx id"
	case "multi-video":
		return "Multiple videos with text between"
	case "function-call":
		return "Function calling: tool call + tool result"
	case "thinking":
		return "DeepSeek response style: reasoning vs final-only"
	default:
		return ""
	}
}
