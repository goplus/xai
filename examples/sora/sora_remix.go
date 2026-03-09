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

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	xai "github.com/goplus/xai/spec"
)

func runRemix() {
	svc := newService()
	ctx := context.Background()

	sourceID := strings.TrimSpace(os.Getenv("SORA_SOURCE_VIDEO_ID"))
	if sourceID == "" {
		fmt.Println("Skip remix: set SORA_SOURCE_VIDEO_ID first")
		return
	}

	op, err := svc.Operation(xai.Model(modelSora2), xai.GenVideo)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().
		Set("Prompt", "Change scene to night, add neon signs, cyberpunk look").
		Set("RemixFromVideoID", sourceID)

	runOperation(ctx, svc, op, "remix")
}

func runRemixSora2Pro() {
	svc := newService()
	ctx := context.Background()

	sourceID := strings.TrimSpace(os.Getenv("SORA_SOURCE_VIDEO_ID"))
	if sourceID == "" {
		fmt.Println("Skip remix-sora2-pro: set SORA_SOURCE_VIDEO_ID first")
		return
	}

	op, err := svc.Operation(xai.Model(modelSora2Pro), xai.GenVideo)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().
		Set("Prompt", "Change scene to night, add neon signs, cyberpunk look").
		Set("RemixFromVideoID", sourceID)

	runOperation(ctx, svc, op, "remix-sora2-pro")
}
