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

// Run: go run ./examples/kling/images kling-image-o1
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunKlingImageO1 runs all kling-image-o1 demos (text2image, aspect ratios, resolution).
func RunKlingImageO1() {
	svc, err := shared.NewService()
	if err != nil {
		fmt.Println("Error:", err)
	}
	ctx := context.Background()

	// 1. text2image 16:9 + 2K
	fmt.Println("--- text2image 16:9 2K ---")
	op, _ := svc.Operation(xai.Model(kling.ModelKlingImageO1), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "a serene mountain landscape at sunset, cinematic")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	op.Params().Set(kling.ParamResolution, kling.Resolution2K)
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	printImageResults("kling-image-o1", "text2image_16:9_2K", results)

	// 2. 带参考图生成（prompt 中用 <<<image_1>>> 引用）
	fmt.Println("--- image2image with reference ---")
	op4, _ := svc.Operation(xai.Model(kling.ModelKlingImageO1), xai.GenImage)
	op4.Params().Set(kling.ParamPrompt, "参考 <<<image_1>>> 的风格，增加一群人，保持背景不变")
	op4.Params().Set(kling.ParamReferenceImages, []string{"https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg"})
	op4.Params().Set(kling.ParamN, 2)
	op4.Params().Set(kling.ParamResolution, kling.Resolution2K)
	op4.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	results4, _ := xai.Call(ctx, svc, op4, svc.Options(), nil)
	printImageResults("kling-image-o1", "image2image_reference", results4)
}
