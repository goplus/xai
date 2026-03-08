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

// DemoURLs holds publicly accessible URLs for OpenAI chat examples.
var DemoURLs = struct {
	RunningManImage string
	VideoMP4        string
	VideoAnimals    string
	VideoAdCopy     string
}{
	RunningManImage: "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
	VideoMP4:        "https://aitoken-public.qnaigc.com/example/generate-video/kling-video-o1-first-end-frame.mp4",
	// gs://cloud-samples-data/... equivalents
	VideoAnimals: "https://storage.googleapis.com/cloud-samples-data/video/animals.mp4",
	VideoAdCopy:  "https://storage.googleapis.com/cloud-samples-data/generative-ai/video/ad_copy_from_video.mp4",
}
