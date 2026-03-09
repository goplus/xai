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

// Veo video generation examples via spec/gemini/provider/qiniu.
// Run: go run ./examples/veo [demo]
package main

import (
	"fmt"
	"os"
)

var demos = map[string]func(){
	"veo-2.0-generate-001":          runVeo20Generate001,
	"veo-2.0-generate-exp":          runVeo20GenerateExp,
	"veo-2.0-generate-preview":      runVeo20GeneratePreview,
	"veo-3.0-generate-preview":      runVeo30GeneratePreview,
	"veo-3.0-fast-generate-preview": runVeo30FastGeneratePreview,
	"veo-3.1-generate-preview":      runVeo31GeneratePreview,
	"veo-3.1-fast-generate-preview": runVeo31FastGeneratePreview,
	"veo-callback":                  runVeoCallback,
	"veo-image-to-video":            runVeoImageToVideo,
	"veo-first-last-frame":          runVeoFirstLastFrame,
	"veo-video-input":               runVeoVideoInput,
	"veo-reference-images":          runVeoReferenceImages,
}

var demoOrder = []string{
	"veo-2.0-generate-001",
	"veo-2.0-generate-exp",
	"veo-2.0-generate-preview",
	"veo-3.0-generate-preview",
	"veo-3.0-fast-generate-preview",
	"veo-3.1-generate-preview",
	"veo-3.1-fast-generate-preview",
	"veo-callback",
	"veo-image-to-video",
	"veo-first-last-frame",
	"veo-video-input",
	"veo-reference-images",
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Veo examples (Qiniu provider):")
		fmt.Println("  Set QINIU_API_KEY for real API calls")
		fmt.Println()
		for _, name := range demoOrder {
			fmt.Printf("  %-30s %s\n", name, demoDesc(name))
		}
		fmt.Printf("  %-30s %s\n", "all", "run all model demos")
		fmt.Println()
		fmt.Println("Usage: go run ./examples/veo [demo|all]")
		return
	}

	for _, arg := range args {
		if arg == "all" {
			for _, name := range demoOrder {
				fmt.Println("---", name, "---")
				demos[name]()
			}
			continue
		}
		if fn, ok := demos[arg]; ok {
			fmt.Println("---", arg, "---")
			fn()
		} else {
			fmt.Printf("Unknown demo: %s\nAvailable: %v + all\n", arg, demoOrder)
		}
	}
}

func demoDesc(name string) string {
	switch name {
	case "veo-2.0-generate-001":
		return "Veo 2.0 baseline text-to-video"
	case "veo-2.0-generate-exp":
		return "Veo 2.0 experimental model"
	case "veo-2.0-generate-preview":
		return "Veo 2.0 preview model"
	case "veo-3.0-generate-preview":
		return "Veo 3.0 preview with full params"
	case "veo-3.0-fast-generate-preview":
		return "Veo 3.0 fast preview"
	case "veo-3.1-generate-preview":
		return "Veo 3.1 preview"
	case "veo-3.1-fast-generate-preview":
		return "Veo 3.1 fast preview"
	case "veo-callback":
		return "Text-to-video with callback_url (PubsubTopic)"
	case "veo-image-to-video":
		return "Image-to-video (veo.md 3.2)"
	case "veo-first-last-frame":
		return "First+last frame (veo.md 3.4)"
	case "veo-video-input":
		return "Video as input (引用视频)"
	case "veo-reference-images":
		return "Multi reference images (多参考图)"
	default:
		return ""
	}
}
