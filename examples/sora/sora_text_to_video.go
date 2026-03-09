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

	xai "github.com/goplus/xai/spec"
)

const (
	modelSora2    = "sora-2"
	modelSora2Pro = "sora-2-pro"
)

func runTextToVideo() {
	svc := newService()
	ctx := context.Background()

	op, err := svc.Operation(xai.Model(modelSora2), xai.GenVideo)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().
		Set("Prompt", "A cute orange cat chasing butterflies in a sunny garden, cinematic shot, high quality").
		Set("Seconds", "4").
		Set("Size", "1280x720")

	runOperation(ctx, svc, op, "text-to-video")
}

func runTextToVideoSora2Pro() {
	svc := newService()
	ctx := context.Background()

	op, err := svc.Operation(xai.Model(modelSora2Pro), xai.GenVideo)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().
		Set("Prompt", "A cute orange cat chasing butterflies in a sunny garden, cinematic shot, high quality").
		Set("Seconds", "4").
		Set("Size", "1280x720")

	runOperation(ctx, svc, op, "text-to-video-sora2-pro")
}

func runTextToVideoPortrait() {
	svc := newService()
	ctx := context.Background()

	op, err := svc.Operation(xai.Model(modelSora2), xai.GenVideo)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().
		Set("Prompt", "A person walking through a rainy street, neon lights reflecting on wet pavement").
		Set("Seconds", "4").
		Set("Size", "720x1280")

	runOperation(ctx, svc, op, "text-to-video-portrait")
}

func runTextToVideo8Sec() {
	svc := newService()
	ctx := context.Background()

	op, err := svc.Operation(xai.Model(modelSora2), xai.GenVideo)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().
		Set("Prompt", "Ocean waves crashing on a rocky shore at sunset, golden hour lighting").
		Set("Seconds", "8").
		Set("Size", "1280x720")

	runOperation(ctx, svc, op, "text-to-video-8sec")
}

func runTextToVideo12Sec() {
	svc := newService()
	ctx := context.Background()

	op, err := svc.Operation(xai.Model(modelSora2), xai.GenVideo)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().
		Set("Prompt", "A timelapse of clouds moving across a mountain range at dawn").
		Set("Seconds", "12").
		Set("Size", "1280x720")

	runOperation(ctx, svc, op, "text-to-video-12sec")
}

func runTextToVideoSora2ProPortrait() {
	svc := newService()
	ctx := context.Background()

	op, err := svc.Operation(xai.Model(modelSora2Pro), xai.GenVideo)
	if err != nil {
		fmt.Println("Operation error:", err)
		return
	}
	op.Params().
		Set("Prompt", "A portrait of a woman with flowing hair in a wind tunnel, cinematic lighting").
		Set("Seconds", "4").
		Set("Size", "1024x1792")

	runOperation(ctx, svc, op, "text-to-video-sora2-pro-portrait")
}
