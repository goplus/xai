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

// Run: go run ./examples/vidu/video
package main

import (
	"fmt"

	xai "github.com/goplus/xai/spec"

	"github.com/goplus/xai/examples/vidu/output"
)

// DemoVideoURLs are public URLs aligned with Vidu route examples.
var DemoVideoURLs = struct {
	Reference1 string
	Reference2 string
	Reference3 string
	ImageInput string
	StartFrame string
	EndFrame   string
}{
	Reference1: "https://storage.googleapis.com/falserverless/web-examples/vidu/new-examples/reference1.png",
	Reference2: "https://storage.googleapis.com/falserverless/web-examples/vidu/new-examples/reference2.png",
	Reference3: "https://storage.googleapis.com/falserverless/web-examples/vidu/new-examples/reference3.png",
	ImageInput: "https://picsum.photos/1024/1024",
	StartFrame: "https://v3.fal.media/files/zebra/sgsdKvPigPhJ1S7Hl5bWc_first_frame_q1.png",
	EndFrame:   "https://v3.fal.media/files/kangaroo/CASBu_OmOnZ8IafirarFL_last_frame_q1.png",
}

func printVideoResults(model, demo string, results xai.Results) {
	if results == nil {
		fmt.Println("  No results")
		return
	}
	urls := make([]string, 0, results.Len())
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		url := out.Video.StgUri()
		urls = append(urls, url)
		fmt.Println("  Video URL:", url)
	}
	output.Append(model, demo, urls, "")
}
