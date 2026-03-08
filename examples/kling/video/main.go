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

var modelOrder = []string{"kling-v2-1", "kling-v2-5-turbo", "kling-v2-6", "kling-video-o1", "kling-v3", "kling-v3-omni"}

func main() {
	models := map[string]func(){
		"kling-v2-1":       RunKlingV21,
		"kling-v2-5-turbo": RunKlingV25Turbo,
		"kling-v2-6":       RunKlingV26,
		"kling-video-o1":   RunKlingVideoO1,
		"kling-v3":         RunKlingV3,
		"kling-v3-omni":    RunKlingV3Omni,
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Kling video examples (by model):")
		for _, name := range modelOrder {
			fmt.Println("\n---", name, "---")
			models[name]()
		}
		return
	}

	for _, arg := range args {
		if fn, ok := models[arg]; ok {
			fmt.Println("---", arg, "---")
			fn()
		} else {
			fmt.Printf("Unknown model: %s\nAvailable: %v\n", arg, modelOrder)
		}
	}
}
