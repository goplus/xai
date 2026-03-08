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

// Chat completions: image with detail (low/medium/high/ultra_high).
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"

	"github.com/goplus/xai/examples/openai/shared"
)

func runChatImageDetailLow() {
	svc := shared.NewService("")
	ctx := context.Background()

	msg := svc.UserMsgExt().
		TextExt("What is shown in this image?").
		ImageURLWithDetail(DemoURLs.RunningManImage, "low")

	params := svc.Params().
		Model(xai.Model(shared.ModelGeminiPro)).
		Messages(msg)

	resp, err := shared.GenOrStream(ctx, svc, params, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if resp == nil {
		return
	}
	shared.PrintResponseBlocksWithTitle("response(detail=low)", resp)
}

func runChatImageDetailUltraHigh() {
	svc := shared.NewService("")
	ctx := context.Background()

	msg := svc.UserMsgExt().
		TextExt("What is this").
		ImageURLWithDetail(DemoURLs.RunningManImage, "ultra_high")

	params := svc.Params().
		Model(xai.Model(shared.ModelGeminiPro)).
		Messages(msg)

	resp, err := shared.GenOrStream(ctx, svc, params, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if resp == nil {
		return
	}
	shared.PrintResponseBlocksWithTitle("response(detail=ultra_high)", resp)
}
