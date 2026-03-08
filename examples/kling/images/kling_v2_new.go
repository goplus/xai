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

// Run: go run ./examples/kling/images kling-v2-new
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunKlingV2New runs all kling-v2-new demos (image2image styles).
func RunKlingV2New() {
	svc, err := shared.NewService()
	if err != nil {
		fmt.Println("Error:", err)
	}
	ctx := context.Background()

	// 1. 赛博朋克风格
	fmt.Println("--- image2image: 赛博朋克 ---")
	op, _ := svc.Operation(xai.Model(kling.ModelKlingV2New), xai.GenImage)
	op.Params().Set(kling.ParamImage, DemoImageURLs.RunningMan)
	op.Params().Set(kling.ParamPrompt, "将这张图片转换为赛博朋克风格")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	printImageResults("kling-v2-new", "image2image_赛博朋克", results)

	// 2. 水墨画风格
	fmt.Println("--- image2image: 水墨画 ---")
	op2, _ := svc.Operation(xai.Model(kling.ModelKlingV2New), xai.GenImage)
	op2.Params().Set(kling.ParamImage, DemoImageURLs.RunningMan)
	op2.Params().Set(kling.ParamPrompt, "将这张图片转换为中国水墨画风格")
	op2.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	results2, _ := xai.Call(ctx, svc, op2, svc.Options(), nil)
	printImageResults("kling-v2-new", "image2image_水墨画", results2)

	// 3. 油画风格
	fmt.Println("--- image2image: 油画 ---")
	op3, _ := svc.Operation(xai.Model(kling.ModelKlingV2New), xai.GenImage)
	op3.Params().Set(kling.ParamImage, DemoImageURLs.RunningMan)
	op3.Params().Set(kling.ParamPrompt, "将这张图片转换为梵高星空油画风格")
	op3.Params().Set(kling.ParamAspectRatio, kling.Aspect1x1)
	results3, _ := xai.Call(ctx, svc, op3, svc.Options(), nil)
	printImageResults("kling-v2-new", "image2image_油画", results3)
}
