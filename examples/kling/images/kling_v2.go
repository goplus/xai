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

// Run: go run ./examples/kling/images kling-v2
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunKlingV2 runs all kling-v2 demos (text2image, image2image, multi_image, negative_prompt).
func RunKlingV2() {
	svc, err := shared.NewService()
	if err != nil {
		fmt.Println("Error:", err)
	}
	ctx := context.Background()

	// 1. text2image
	fmt.Println("--- text2image ---")
	op, _ := svc.Operation(xai.Model(kling.ModelKlingV2), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "一只可爱的橘猫坐在窗台上看着夕阳,照片风格,高清画质")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printImageResults("kling-v2", "text2image", results)

	// 2. image2image
	fmt.Println("--- image2image ---")
	op2, _ := svc.Operation(xai.Model(kling.ModelKlingV2), xai.GenImage)
	op2.Params().Set(kling.ParamPrompt, "将这张图片转换为水彩画风格")
	op2.Params().Set(kling.ParamImage, DemoImageURLs.RunningMan)
	op2.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	results2, err := xai.Call(ctx, svc, op2, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printImageResults("kling-v2", "image2image", results2)

	// 3. multi_image
	fmt.Println("--- multi_image ---")
	op3, _ := svc.Operation(xai.Model(kling.ModelKlingV2), xai.GenImage)
	op3.Params().Set(kling.ParamPrompt, "综合两个图像画一个漫画图")
	op3.Params().Set(kling.ParamSubjectImageList, []map[string]string{
		{kling.ParamSubjectImage: DemoImageURLs.Subject1},
		{kling.ParamSubjectImage: DemoImageURLs.Subject2},
	})
	op3.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	results3, err := xai.Call(ctx, svc, op3, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printImageResults("kling-v2", "multi_image", results3)

	// 4. text2image + negative_prompt
	fmt.Println("--- text2image + negative_prompt ---")
	op4, _ := svc.Operation(xai.Model(kling.ModelKlingV2), xai.GenImage)
	op4.Params().Set(kling.ParamPrompt, "梦幻森林，萤火虫飞舞")
	op4.Params().Set(kling.ParamNegativePrompt, "模糊,低质量")
	op4.Params().Set(kling.ParamAspectRatio, kling.Aspect9x16)
	results4, _ := xai.Call(ctx, svc, op4, svc.Options(), nil)
	printImageResults("kling-v2", "text2image+negative_prompt", results4)
}
