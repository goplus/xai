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

// Run: go run ./examples/kling/images call-sync
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunCallSyncExample demonstrates CallSync + Wait: start the operation, get resp and taskID,
// persist taskID to DB, then call Wait. For async tasks, you can later resume via GetTask.
func RunCallSyncExample() {
	svc, err := shared.NewService()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	model := xai.Model(kling.ModelKlingV1)
	op, err := svc.Operation(model, xai.GenImage)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	op.Params().Set(kling.ParamPrompt, "一只可爱的橘猫坐在窗台上看着夕阳,照片风格,高清画质")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)

	ctx := context.Background()

	// CallSync: start operation, get resp (no wait)
	resp, err := xai.CallSync(ctx, svc, op, svc.Options())
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Get taskID for persistence (empty for sync responses; non-empty for async)
	taskID := resp.TaskID()
	var results xai.Results
	if taskID != "" {
		// Save to DB, then resume from taskID (simulate cross-process: load taskID from DB)
		fmt.Println("TaskID saved to DB:", taskID)
		resp2, err := xai.GetTask(ctx, svc, model, xai.GenImage, taskID)
		if err != nil {
			fmt.Println("Error:", err)
		}
		results, err = xai.Wait(ctx, svc, resp2, nil)
	} else {
		// Sync response: Wait directly on resp
		fmt.Println("Sync response (no taskID), waiting...")
		results, err = xai.Wait(ctx, svc, resp, nil)
	}
	if err != nil {
		fmt.Println("Error:", err)
	}
	printImageResults(string(model), "text2image", results)
}
