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

// Run: go run ./examples/vidu/video call-sync
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/vidu"
)

// RunCallSyncExample demonstrates CallSync + Wait:
// start operation, get taskID, then resume through xai.GetTask.
func RunCallSyncExample() {
	svc, err := newService()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	ctx := context.Background()
	model := xai.Model(vidu.ModelViduQ2)
	op, err := svc.Operation(model, xai.GenVideo)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	op.Params().
		Set(vidu.ParamPrompt, "Wide cinematic shot: a small boat crossing a glowing lake at dusk.").
		Set(vidu.ParamSeed, 31).
		Set(vidu.ParamDuration, 4).
		Set(vidu.ParamResolution, vidu.Resolution720p).
		Set(vidu.ParamMovementAmplitude, "auto").
		Set(vidu.ParamWatermark, true)

	// CallSync: start operation and get resp immediately.
	resp, err := xai.CallSync(ctx, svc, op, newViduOptions(svc, "demo-user-call-sync"))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	taskID := resp.TaskID()
	var results xai.Results
	if taskID != "" {
		// Persist taskID to DB, then resume by taskID.
		fmt.Println("TaskID saved to DB:", taskID)
		resp2, err := xai.GetTask(ctx, svc, model, xai.GenVideo, taskID)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		results, err = xai.Wait(ctx, svc, resp2, progressPrinter("call-sync"))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	} else {
		fmt.Println("Sync response (no taskID), waiting directly...")
		results, err = xai.Wait(ctx, svc, resp, progressPrinter("call-sync-direct"))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}
	printVideoResults(string(model), "call-sync", results)
}
