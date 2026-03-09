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

// Run: go run ./examples/veo
package main

// DemoURLs holds publicly accessible image/video URLs for Veo examples.
// All use URLs (no local files). Aligned with kling/video/urls.go style.
var DemoURLs = struct {
	// Image-to-video
	Image string
	// First+last frame
	FirstFrame string
	LastFrame  string
	// Video input (引用视频)
	Video string
	// Reference images (多参考图, veo-2.0-generate-exp / veo-3.1-generate-preview)
	// Up to 3 asset or 1 style; duration must be 8s
	RefAsset1 string
	RefAsset2 string
	RefStyle  string
}{
	Image:      "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
	FirstFrame: "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
	LastFrame:  "https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png",
	Video:      "https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4",
	RefAsset1:  "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
	RefAsset2:  "https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png",
	RefStyle:   "https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png",
}
