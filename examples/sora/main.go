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

// Sora video operation examples via spec/openai + Qiniu endpoint.
// Run: go run ./examples/sora [demo]
package main

import (
	"fmt"
	"os"
)

var demos = map[string]func(){
	"text-to-video":                runTextToVideo,
	"text-to-video-sora2-pro":       runTextToVideoSora2Pro,
	"text-to-video-portrait":        runTextToVideoPortrait,
	"text-to-video-8sec":            runTextToVideo8Sec,
	"text-to-video-12sec":           runTextToVideo12Sec,
	"text-to-video-sora2-pro-portrait": runTextToVideoSora2ProPortrait,
	"image-to-video":               runImageToVideo,
	"image-to-video-sora2-pro":      runImageToVideoSora2Pro,
	"remix":                        runRemix,
	"remix-sora2-pro":              runRemixSora2Pro,
}

var demoOrder = []string{
	"text-to-video",
	"text-to-video-sora2-pro",
	"text-to-video-portrait",
	"text-to-video-8sec",
	"text-to-video-12sec",
	"text-to-video-sora2-pro-portrait",
	"image-to-video",
	"image-to-video-sora2-pro",
	"remix",
	"remix-sora2-pro",
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Sora examples (Qiniu backend):")
		fmt.Println("  Set QINIU_API_KEY for real API calls")
		fmt.Println()
		for _, name := range demoOrder {
			fmt.Printf("  %-30s %s\n", name, demoDesc(name))
		}
		fmt.Printf("  %-30s %s\n", "all", "run all demos")
		fmt.Println()
		fmt.Println("Usage: go run ./examples/sora [demo|all]")
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
	case "text-to-video":
		return "Sora-2 text-to-video (4s, 1280x720)"
	case "text-to-video-sora2-pro":
		return "Sora-2-pro text-to-video (4s, 1280x720)"
	case "text-to-video-portrait":
		return "Sora-2 portrait (4s, 720x1280)"
	case "text-to-video-8sec":
		return "Sora-2 text-to-video (8s)"
	case "text-to-video-12sec":
		return "Sora-2 text-to-video (12s)"
	case "text-to-video-sora2-pro-portrait":
		return "Sora-2-pro portrait (4s, 1024x1792)"
	case "image-to-video":
		return "Sora-2 image-to-video with input_reference"
	case "image-to-video-sora2-pro":
		return "Sora-2-pro image-to-video"
	case "remix":
		return "Sora-2 remix from SORA_SOURCE_VIDEO_ID"
	case "remix-sora2-pro":
		return "Sora-2-pro remix from SORA_SOURCE_VIDEO_ID"
	default:
		return ""
	}
}
