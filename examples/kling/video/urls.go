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

// Run: go run ./examples/kling/video
package main

import (
	"fmt"

	xai "github.com/goplus/xai/spec"

	"github.com/goplus/xai/examples/kling/output"
)

// DemoVideoURLs holds real, publicly accessible image/video URLs for Kling video examples.
// Aligned with qnaigc.com API curl examples.
var DemoVideoURLs = struct {
	FirstFrame   string // for keyframe
	EndFrame     string // for keyframe
	RunningMan   string // for img2video (qnaigc example)
	MultiRef1    string // for multi_ref
	MultiRef2    string // for multi_ref (first_frame type)
	MotionImage  string // character image for motion control (V2.6)
	MotionVideo  string // reference video for motion control (V2.6)
	VideoFeature string // for video2video (refer_type: feature)
}{
	FirstFrame:   "https://picsum.photos/1280/720",
	EndFrame:     "https://picsum.photos/1280/720",
	RunningMan:   "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
	MultiRef1:    "https://picsum.photos/1280/720",
	MultiRef2:    "https://picsum.photos/1280/720",
	MotionImage:  "https://p2-kling.klingai.com/kcdn/cdn-kcdn112452/kling-qa-test/multi-3.ng.png",
	MotionVideo:  "https://p2-kling.klingai.com/kcdn/cdn-kcdn112452/kling-qa-test/dance.mp4",
	VideoFeature: "https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4",
}

// printVideoResults prints and saves video URLs for verification.
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
