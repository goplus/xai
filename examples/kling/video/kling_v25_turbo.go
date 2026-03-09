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

// Run: go run ./examples/kling/video kling-v2-5-turbo
//
// kling-v2-5-turbo 是可灵 V2.5 Turbo 视频生成模型，支持以下功能：
//
// 1. text2video    - 纯文本生成视频
// 2. img2video     - 单图参考生成视频
// 3. keyframe      - 首尾帧生成视频
//
// 特点：
//   - 相比 V2.1，支持纯文本生成视频
//   - 生成速度更快（Turbo）
//   - 支持负向提示词
//
// 参数说明：
//   - input_reference: 参考图片 URL（可选，用于 img2video）
//   - image_tail: 尾帧图片 URL（可选，用于首尾帧模式）
//   - negative_prompt: 负向提示词（描述不希望出现的内容）
//   - mode: "std"（标准 720P）或 "pro"（专家 1080P）
//   - seconds: "5" 或 "10"
//   - size: 视频尺寸
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunKlingV25Turbo runs all kling-v2-5-turbo demos.
func RunKlingV25Turbo() {
	svc, _ := shared.NewService()
	ctx := context.Background()

	// =========================================================================
	// 1. text2video - 纯文本生成视频（标准模式）
	// =========================================================================
	fmt.Println("--- text2video (std) ---")
	op1, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
	op1.Params().
		Set(kling.ParamPrompt, "一只可爱的小猫在草地上玩耍").
		Set(kling.ParamMode, kling.ModeStd).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1280x720)
	results1, _ := xai.Call(ctx, svc, op1, svc.Options(), nil)
	printVideoResults("kling-v2-5-turbo", "text2video-std", results1)

	// =========================================================================
	// 2. text2video - 纯文本生成视频（专家模式 + 负向提示词）
	// =========================================================================
	fmt.Println("--- text2video (pro + negative) ---")
	op2, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
	op2.Params().
		Set(kling.ParamPrompt, "a hero enters the battlefield, dramatic lighting, slow motion").
		Set(kling.ParamNegativePrompt, "blurry, low quality, distorted").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results2, _ := xai.Call(ctx, svc, op2, svc.Options(), nil)
	printVideoResults("kling-v2-5-turbo", "text2video-pro", results2)

	// =========================================================================
	// 3. text2video - 竖屏视频
	// =========================================================================
	fmt.Println("--- text2video (竖屏) ---")
	op3, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
	op3.Params().
		Set(kling.ParamPrompt, "一个人在跳街舞，动感十足").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1080x1920)     // 竖屏 9:16
	results3, _ := xai.Call(ctx, svc, op3, svc.Options(), nil)
	printVideoResults("kling-v2-5-turbo", "text2video-vertical", results3)

	// =========================================================================
	// 4. img2video - 单图参考生成视频
	// =========================================================================
	fmt.Println("--- img2video ---")
	op4, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
	op4.Params().
		Set(kling.ParamPrompt, "人在奔跑，镜头跟随").
		Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results4, _ := xai.Call(ctx, svc, op4, svc.Options(), nil)
	printVideoResults("kling-v2-5-turbo", "img2video", results4)

	// =========================================================================
	// 5. img2video - 带负向提示词
	// =========================================================================
	fmt.Println("--- img2video (negative_prompt) ---")
	op5, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
	op5.Params().
		Set(kling.ParamPrompt, "人物优雅地行走，电影质感").
		Set(kling.ParamInputReference, DemoVideoURLs.RunningMan).
		Set(kling.ParamNegativePrompt, "blurry, jittery, unstable, low quality").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results5, _ := xai.Call(ctx, svc, op5, svc.Options(), nil)
	printVideoResults("kling-v2-5-turbo", "img2video-negative", results5)

	// =========================================================================
	// 6. keyframe - 首尾帧生成视频
	// =========================================================================
	fmt.Println("--- keyframe (首尾帧) ---")
	op6, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
	op6.Params().
		Set(kling.ParamPrompt, "人在跑到了天涯海角").
		Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame). // 首帧
		Set(kling.ParamImageTail, DemoVideoURLs.EndFrame).        // 尾帧
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results6, _ := xai.Call(ctx, svc, op6, svc.Options(), nil)
	printVideoResults("kling-v2-5-turbo", "keyframe", results6)

	// =========================================================================
	// 7. 长视频 - 10秒视频
	// =========================================================================
	fmt.Println("--- long video (10s) ---")
	op7, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
	op7.Params().
		Set(kling.ParamPrompt, "延时摄影，城市从白天到黑夜的变化").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds10).     // 10秒视频
		Set(kling.ParamSize, kling.Size1920x1080)
	results7, _ := xai.Call(ctx, svc, op7, svc.Options(), nil)
	printVideoResults("kling-v2-5-turbo", "long-video", results7)

	// =========================================================================
	// 8. 方形视频 - 适合社交媒体
	// =========================================================================
	fmt.Println("--- square video ---")
	op8, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
	op8.Params().
		Set(kling.ParamPrompt, "产品展示，旋转展示各个角度").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1080x1080)     // 方形 1:1
	results8, _ := xai.Call(ctx, svc, op8, svc.Options(), nil)
	printVideoResults("kling-v2-5-turbo", "square-video", results8)
}
