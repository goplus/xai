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
 * WITHOUT WARRANTIES OR CONDITIONS OF KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Run: go run ./examples/kling models
package main

import (
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunModels lists ImageModels, VideoModels, and Actions per model. No API call.
func RunModels() {
	svc := shared.NewServiceForModels()

	fmt.Println("Image models:", kling.ImageModels())
	fmt.Println("Video models:", kling.VideoModels())
	fmt.Println()

	actions := svc.Actions(xai.Model(kling.ModelKlingV21))
	fmt.Println("kling-v2-1 actions:", actions)

	actionsO1 := svc.Actions(xai.Model(kling.ModelKlingVideoO1))
	fmt.Println("kling-video-o1 actions:", actionsO1)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
	fields := op.InputSchema().Fields()
	fmt.Printf("kling-v2-1 GenImage schema: %d fields\n", len(fields))
	for _, f := range fields {
		fmt.Printf("  - %s\n", f.Name)
	}
}
