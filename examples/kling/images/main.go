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

var modelOrder = []string{"call-sync", "kling-v1", "kling-v1-5", "kling-v2", "kling-v2-new", "kling-v2-1", "kling-image-o1"}

func main() {
	models := map[string]func(){
		"call-sync":      RunCallSyncExample,
		"kling-v1":       RunKlingV1,
		"kling-v1-5":     RunKlingV15,
		"kling-v2":       RunKlingV2,
		"kling-v2-new":   RunKlingV2New,
		"kling-v2-1":     RunKlingV21,
		"kling-image-o1": RunKlingImageO1,
	}

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Kling image examples (by model):")
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
