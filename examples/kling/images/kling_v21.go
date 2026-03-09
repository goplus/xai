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

// Run: go run ./examples/kling/images kling-v2-1
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunKlingV21 runs all kling-v2-1 demos (text2image, image2image reference_images, multi_image).
func RunKlingV21() {
	svc, err := shared.NewService()
	if err != nil {
		fmt.Println("Error:", err)
	}
	ctx := context.Background()

	// 1. text2image
	fmt.Println("--- text2image ---")
	op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "a sunset over the ocean, cinematic lighting")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	op.Params().Set(kling.ParamN, 1)
	results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printImageResults("kling-v2-1", "text2image", results)

	// 2. image2image (reference_images 风格参考)
	fmt.Println("--- image2image (reference_images) ---")
	op2, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
	op2.Params().Set(kling.ParamPrompt, "same style but with a mountain landscape")
	op2.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	op2.Params().Set(kling.ParamReferenceImages, []string{DemoImageURLs.RefStyle})
	op2.Params().Set(kling.ParamNegativePrompt, "blurry, low quality")
	results2, err := xai.Call(ctx, svc, op2, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printImageResults("kling-v2-1", "image2image_reference_images", results2)

	// 3. multi_image (subject_image_list + scene_image + style_image)
	fmt.Println("--- multi_image ---")
	op3, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
	op3.Params().Set(kling.ParamPrompt, "一个梦幻般的森林场景")
	op3.Params().Set(kling.ParamSubjectImageList, []map[string]string{
		{kling.ParamSubjectImage: DemoImageURLs.Subject1},
		{kling.ParamSubjectImage: DemoImageURLs.Subject2},
	})
	op3.Params().Set(kling.ParamSceneImage, DemoImageURLs.Subject2)
	op3.Params().Set(kling.ParamStyleImage, DemoImageURLs.Subject1)
	op3.Params().Set(kling.ParamAspectRatio, kling.Aspect9x16)
	results3, err := xai.Call(ctx, svc, op3, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printImageResults("kling-v2-1", "multi_image", results3)
}
