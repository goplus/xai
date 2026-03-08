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

// Chat completions: video + text (equivalent to curl example 3).
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"

	"github.com/goplus/xai/examples/openai/shared"
)

func runChatVideo() {
	svc := shared.NewService("")
	ctx := context.Background()

	msg := svc.UserMsgExt().
		TextExt("What is in this video?").
		VideoFile(DemoURLs.VideoMP4, "video/mp4")

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
	shared.PrintResponseBlocksWithTitle("response", resp)
}

func runChatVideoFileID() {
	svc := shared.NewService("")
	ctx := context.Background()

	msg := svc.UserMsgExt().
		TextExt("What is this").
		VideoFile("qfile-xxxx-1770719212268100147-e0011b", "video/mp4")

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
	shared.PrintResponseBlocksWithTitle("response", resp)
}
