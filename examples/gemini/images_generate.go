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

	"github.com/goplus/xai/examples/gemini/shared"
)

func runImageGenerate() {
	// shared.NewService returns xai.Service; examples only rely on interface APIs.
	service := shared.NewService("")
	ctx := context.Background()

	op, err := service.Operation(xai.Model(shared.ModelFlashImage), xai.GenImage)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	op.Params().
		Set("Prompt", "梦幻森林中的精灵小屋，魔法光芒环绕").
		Set("AspectRatio", "16:9").
		Set("ImageSize", "1K")

	resp, err := op.Call(ctx, service, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printImageResults(resp.Results())
}

func runImageGenerateSimple() {
	service := shared.NewService("")
	ctx := context.Background()

	op, err := service.Operation(xai.Model(shared.ModelFlashImage), xai.GenImage)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	op.Params().
		Set("Prompt", "一只可爱的橘猫坐在窗台上看着夕阳，照片风格，高清画质")

	resp, err := op.Call(ctx, service, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printImageResults(resp.Results())
}

func runImageGeneratePortrait() {
	service := shared.NewService("")
	ctx := context.Background()

	op, err := service.Operation(xai.Model(shared.ModelV31ImagePreview), xai.GenImage)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	op.Params().
		Set("Prompt", "梦幻森林中的精灵小屋，魔法光芒环绕").
		Set("AspectRatio", "9:16").
		Set("ImageSize", "1K")

	resp, err := op.Call(ctx, service, nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	printImageResults(resp.Results())
}
