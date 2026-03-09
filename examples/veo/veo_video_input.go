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

// runVeoVideoInput demonstrates video-as-input generation (引用视频).
// Uses DemoURLs.Video (public URL, video ≤10MB, no local files).
func runVeoVideoInput() {
	service := qiniu.NewService(strings.TrimSpace(os.Getenv("QINIU_API_KEY")))
	ctx := context.Background()

	video := service.VideoFromStgUri(videoMimeFromURL(DemoURLs.Video), DemoURLs.Video)

	op, err := service.Operation(xai.Model("veo-3.0-generate-preview"), xai.GenVideo)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Video input: Video + Prompt (引用视频，视频 ≤10MB)
	op.Params().
		Set("Video", video).
		Set("Prompt", "Continue the scene with smooth motion").
		Set("AspectRatio", gemini.Aspect16x9).
		Set("DurationSeconds", gemini.Duration6).
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
			fmt.Printf("  [veo-video-input] polling task: %s\n", taskID)
		}
	})
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("veo-video-input: videos=%d\n", results.Len())
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Printf("  video[%d]: %s (%s)\n", i, out.Video.StgUri(), out.Video.Type())
	}
}
