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

// Run: go run ./examples/kling/images
package main

import (
	"fmt"

	xai "github.com/goplus/xai/spec"

	"github.com/goplus/xai/examples/kling/output"
)

// DemoImageURLs holds real, publicly accessible image URLs for Kling image examples.
// All URLs are from qnaigc.com for consistency.
var DemoImageURLs = struct {
	RefStyle   string // for image2image reference_images
	Subject1   string // for multi_image
	Subject2   string // for multi_image
	RunningMan string // for image2image, single image edit
}{
	RefStyle:   "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
	Subject1:   "https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png",
	Subject2:   "https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg",
	RunningMan: "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg",
}

// printImageResults prints and saves image URLs for verification.
func printImageResults(model, demo string, results xai.Results) {
	if results == nil {
		fmt.Println("  No results")
		return
	}
	urls := make([]string, 0, results.Len())
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		url := out.Image.StgUri()
		urls = append(urls, url)
		fmt.Println("  Image URL:", url)
	}
	output.Append(model, demo, urls, "")
}
