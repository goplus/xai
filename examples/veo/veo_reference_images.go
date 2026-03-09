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

// runVeoReferenceImages demonstrates multi-reference image video generation (多参考图).
// Only veo-2.0-generate-exp and veo-3.1-generate-preview support this.
// Up to 3 asset images OR 1 style image; duration must be 8s; cannot use with Image/Video/LastFrame.
func runVeoReferenceImages() {
	service := qiniu.NewService(strings.TrimSpace(os.Getenv("QINIU_API_KEY")))
	ctx := context.Background()

	// Asset refs: up to 3 images of subject/character/product
	img1 := service.ImageFromStgUri(mimeFromURL(DemoURLs.RefAsset1), DemoURLs.RefAsset1)
	img2 := service.ImageFromStgUri(mimeFromURL(DemoURLs.RefAsset2), DemoURLs.RefAsset2)
	refs := service.GenVideoReferenceImages(
		xai.GenVideoReferenceImage{Image: img1, ReferenceType: "asset"},
		xai.GenVideoReferenceImage{Image: img2, ReferenceType: "asset"},
	)

	op, err := service.Operation(xai.Model("veo-3.1-generate-preview"), xai.GenVideo)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// ReferenceImages: Prompt + refs; duration must be 8
	op.Params().
		Set("Prompt", "A cinematic scene with the characters walking through a garden").
		Set("ReferenceImages", refs).
		Set("AspectRatio", gemini.Aspect16x9).
		Set("DurationSeconds", gemini.Duration8).
		Set("NumberOfVideos", int32(1)).
		Set("PersonGeneration", gemini.PersonDontAllow)

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
			fmt.Printf("  [veo-reference-images] polling task: %s\n", taskID)
		}
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("veo-reference-images: videos=%d\n", results.Len())
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Printf("  video[%d]: %s (%s)\n", i, out.Video.StgUri(), out.Video.Type())
	}
}
