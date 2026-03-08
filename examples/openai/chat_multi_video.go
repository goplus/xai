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

// Chat completions: multiple videos with text between (equivalent to curl example 7).
// Text should be inserted between videos to keep multimodal context clear.
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"

	"github.com/goplus/xai/examples/openai/shared"
)

func runChatMultiVideo() {
	svc := shared.NewService("")
	ctx := context.Background()

	msg := svc.UserMsgExt().
		TextExt("What is in this video?").
		VideoFile(DemoURLs.VideoAnimals, "video/mp4").
		TextExt("What is in this video?").
		VideoFile(DemoURLs.VideoAdCopy, "video/mp4")

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
