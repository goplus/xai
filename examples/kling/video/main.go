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
 * See the License for the specific language governing permissions and limitations under the License.
 */

// Kling video generation examples by model.
// Run: go run ./examples/kling/video [model]
// With QINIU_API_KEY set, uses real Qnagic API; otherwise mock.
package main

import (
	"fmt"
	"os"
)

var demos = map[string]func(){
	"kling-v2-1":       RunKlingV21,
	"kling-v2-5-turbo": RunKlingV25Turbo,
	"kling-v2-6":       RunKlingV26,
	"kling-video-o1":   RunKlingVideoO1,
	"kling-v3":         RunKlingV3,
	"kling-v3-omni":    RunKlingV3Omni,
}

var demoOrder = []string{
	"kling-v2-1", "kling-v2-5-turbo", "kling-v2-6", "kling-video-o1", "kling-v3", "kling-v3-omni",
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Kling video examples (by model):")
		fmt.Println("  Set QINIU_API_KEY for real API calls")
		fmt.Println()
		for _, name := range demoOrder {
			fmt.Printf("  %-15s %s\n", name, demoDesc(name))
		}
		fmt.Println()
		fmt.Println("Usage: go run ./examples/kling/video [demo]")
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

func demoDesc(name string) string {
	switch name {
	case "kling-v2-1":
		return "Kling v2.1 model"
	case "kling-v2-5-turbo":
		return "Kling v2.5 turbo model"
	case "kling-v2-6":
		return "Kling v2.6 model"
	case "kling-video-o1":
		return "Kling video o1 model"
	case "kling-v3":
		return "Kling v3 model"
	case "kling-v3-omni":
		return "Kling v3 omni model"
	default:
		return ""
	}
}
