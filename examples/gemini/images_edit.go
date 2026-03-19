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
	"google.golang.org/genai"

	"github.com/goplus/xai/examples/gemini/shared"
)

func runImageEdit() {
	// shared.NewService returns xai.Service; examples only rely on interface APIs.
	service := shared.NewService("")
	ctx := context.Background()

	op, err := service.Operation(xai.Model(shared.ModelFlashImage), xai.EditImage)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	img1 := service.ImageFromStgUri(xai.ImageJPEG, DemoURLs.RunningMan)
	img2 := service.ImageFromStgUri(xai.ImageJPEG, DemoURLs.Lawn)
	ref1, _ := service.ReferenceImage(img1, 0, xai.RawReferenceImage)
	ref2, _ := service.ReferenceImage(img2, 1, xai.StyleReferenceImage)

	op.Params().
		Set("Prompt", "结合这两张图片的风格，生成一张新的艺术作品").
		Set("References", []genai.ReferenceImage{ref1.(genai.ReferenceImage), ref2.(genai.ReferenceImage)}).
		Set("AspectRatio", "16:9").
		Set("ImageSize", "1K")

	resp, err := op.Call(ctx, service, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printImageResults(resp.Results())
}

func runImageEditSingle() {
	service := shared.NewService("")
	ctx := context.Background()

	op, err := service.Operation(xai.Model(shared.ModelV31ImagePreview), xai.EditImage)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	img := service.ImageFromStgUri(xai.ImageJPEG, DemoURLs.RunningMan)
	ref, _ := service.ReferenceImage(img, 0, xai.RawReferenceImage)

	op.Params().
		Set("Prompt", "为这个场景添加日落效果，让整体色调更温暖").
		Set("References", []genai.ReferenceImage{ref.(genai.ReferenceImage)}).
		Set("ImageSize", "1K")

	resp, err := op.Call(ctx, service, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printImageResults(resp.Results())
}

func runImageEditMask() {
	service := shared.NewService("")
	ctx := context.Background()

	op, err := service.Operation(xai.Model(shared.ModelV31ImagePreview), xai.EditImage)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	imgBase := service.ImageFromStgUri(xai.ImageJPEG, DemoURLs.MaskBase)
	imgMask := service.ImageFromStgUri(xai.ImagePNG, DemoURLs.MaskMask)
	refBase, _ := service.ReferenceImage(imgBase, 0, xai.RawReferenceImage)
	refMask, _ := service.ReferenceImage(imgMask, 1, xai.MaskReferenceImage)

	op.Params().
		Set("Prompt", "使用第二张图片作为遮罩图，仅在遮罩图中的白色区域允许生成内容。在第一张图片的对应位置添加两个人正在拥抱的场景。遮罩以白色区域为可生成区域，黑色区域保持第一张图片不变，不要修改遮罩外的背景、建筑或已有物体。不要把遮罩的白色保留到第一个图片。").
		Set("References", []genai.ReferenceImage{refBase.(genai.ReferenceImage), refMask.(genai.ReferenceImage)}).
		Set("AspectRatio", "16:9").
		Set("ImageSize", "1K")

	resp, err := op.Call(ctx, service, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printImageResults(resp.Results())
}
