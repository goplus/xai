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
	"github.com/goplus/xai/spec/gemini"
	"github.com/goplus/xai/spec/gemini/provider/qiniu"
)

// runVeoImageToVideo demonstrates image-to-video generation (veo.md 3.2).
// Uses DemoURLs.Image (public URL, no local files).
func runVeoImageToVideo() {
	service := qiniu.NewService(strings.TrimSpace(os.Getenv("QINIU_API_KEY")))
	ctx := context.Background()

	img := service.ImageFromStgUri(mimeFromURL(DemoURLs.Image), DemoURLs.Image)

	op, err := service.Operation(xai.Model("veo-3.0-generate-preview"), xai.GenVideo)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Image-to-video: Image + Prompt (matching veo.md 3.2 curl)
	op.Params().
		Set("Image", img).
		Set("Prompt", "A gentle breeze blowing through the scene").
		Set("AspectRatio", gemini.Aspect16x9).
		Set("DurationSeconds", gemini.Duration6).
		Set("NumberOfVideos", int32(1)).
		Set("Seed", int32(100)).
		Set("NegativePrompt", "").
		Set("PersonGeneration", gemini.PersonAllowAdult) //PersonDontAllow)

	resp, err := xai.CallSync(ctx, service, op, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if taskID := resp.TaskID(); taskID != "" {
		fmt.Println("task_id:", taskID)
	}

	results, err := xai.Wait(ctx, service, resp, func(resp xai.OperationResponse) {
		if taskID := resp.TaskID(); taskID != "" {
			fmt.Printf("  [veo-image-to-video] polling task: %s\n", taskID)
		}
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("veo-image-to-video: videos=%d\n", results.Len())
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Printf("  video[%d]: %s (%s)\n", i, out.Video.StgUri(), out.Video.Type())
	}
}
