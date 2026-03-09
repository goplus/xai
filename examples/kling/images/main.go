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

// Kling image generation examples by model (see spec/kling/kling_image.md).
// Run: go run ./examples/kling/images [model]
// With QINIU_API_KEY set, uses real Qnagic API; otherwise mock.
package main

import (
	"fmt"
	"os"
)

var demos = map[string]func(){
	"call-sync":      RunCallSyncExample,
	"kling-v1":       RunKlingV1,
	"kling-v1-5":     RunKlingV15,
	"kling-v2":       RunKlingV2,
	"kling-v2-new":   RunKlingV2New,
	"kling-v2-1":     RunKlingV21,
	"kling-image-o1": RunKlingImageO1,
}

var demoOrder = []string{
	"call-sync", "kling-v1", "kling-v1-5", "kling-v2", "kling-v2-new", "kling-v2-1", "kling-image-o1",
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Kling image examples (by model):")
		fmt.Println("  Set QINIU_API_KEY for real API calls")
		fmt.Println()
		for _, name := range demoOrder {
			fmt.Printf("  %-15s %s\n", name, demoDesc(name))
		}
		fmt.Println()
		fmt.Println("Usage: go run ./examples/kling/images [demo]")
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
	case "call-sync":
		return "Call sync API"
	case "kling-v1":
		return "Kling v1 model"
	case "kling-v1-5":
		return "Kling v1.5 model"
	case "kling-v2":
		return "Kling v2 model"
	case "kling-v2-new":
		return "Kling v2 new model"
	case "kling-v2-1":
		return "Kling v2.1 model"
	case "kling-image-o1":
		return "Kling image o1 model"
	default:
		return ""
	}
}
