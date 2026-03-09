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

// Run: go run ./examples/kling/video kling-v2-1
//
// kling-v2-1 是可灵 V2.1 视频生成模型，仅支持图生视频：
//
// 1. img2video     - 单图参考生成视频（必须提供 input_reference）
// 2. keyframe      - 首尾帧生成视频（input_reference + image_tail）
//
// 重要限制：
//   - V2.1 不支持纯文本生成视频（text2video）
//   - input_reference 是必填参数
//   - 首尾帧模式要求：mode="pro" 且 seconds="10"
//
// 参数说明：
//   - input_reference: 参考图片 URL（必填，作为首帧）
//   - image_tail: 尾帧图片 URL（可选，用于首尾帧模式）
//   - negative_prompt: 负向提示词（描述不希望出现的内容）
//   - mode: "std"（标准 720P）或 "pro"（专家 1080P）
//   - seconds: "5" 或 "10"（首尾帧模式必须为 "10"）
//   - size: 视频尺寸
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunKlingV21 runs all kling-v2-1 demos.
func RunKlingV21() {
	svc, err := shared.NewService()
	if err != nil {
		fmt.Println("Error:", err)
	}
	ctx := context.Background()

	// =========================================================================
	// 1. img2video - 单图参考生成视频（标准模式）
	// =========================================================================
	fmt.Println("--- img2video (std) ---")
	op1, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
	op1.Params().
		Set(kling.ParamPrompt, "镜头缓慢右移，人物开始奔跑").
		Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame). // 必填
		Set(kling.ParamMode, kling.ModeStd).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1280x720)
	results1, err := xai.Call(ctx, svc, op1, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-1", "img2video-std", results1)

	// =========================================================================
	// 2. img2video - 单图参考生成视频（专家模式）
	// =========================================================================
	fmt.Println("--- img2video (pro) ---")
	op2, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
	op2.Params().
		Set(kling.ParamPrompt, "camera pans slowly to the right, cinematic").
		Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame).
		Set(kling.ParamMode, kling.ModePro).          // 专家模式 1080P
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results2, err := xai.Call(ctx, svc, op2, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-1", "img2video-pro", results2)

	// =========================================================================
	// 3. img2video - 带负向提示词
	// =========================================================================
	// negative_prompt 描述不希望出现的内容
	fmt.Println("--- img2video (negative_prompt) ---")
	op3, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
	op3.Params().
		Set(kling.ParamPrompt, "人物优雅地行走，电影质感").
		Set(kling.ParamInputReference, DemoVideoURLs.RunningMan).
		Set(kling.ParamNegativePrompt, "blurry, low quality, jittery, unstable"). // 负向提示词
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results3, err := xai.Call(ctx, svc, op3, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-1", "img2video-negative", results3)

	// =========================================================================
	// 4. keyframe - 首尾帧生成视频
	// =========================================================================
	// 重要：首尾帧模式要求 mode="pro" 且 seconds="10"
	fmt.Println("--- keyframe (首尾帧) ---")
	op4, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
	op4.Params().
		Set(kling.ParamPrompt, "smooth transition from day to night").
		Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame). // 首帧
		Set(kling.ParamImageTail, DemoVideoURLs.EndFrame).        // 尾帧
		Set(kling.ParamMode, kling.ModePro).                      // 首尾帧必须 pro
		Set(kling.ParamSeconds, kling.Seconds10).                 // 首尾帧必须 10秒
		Set(kling.ParamNegativePrompt, "jittery, unstable").
		Set(kling.ParamSize, kling.Size1920x1080)
	results4, err := xai.Call(ctx, svc, op4, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-1", "keyframe", results4)

	// =========================================================================
	// 5. img2video - 竖屏视频
	// =========================================================================
	fmt.Println("--- img2video (竖屏) ---")
	op5, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
	op5.Params().
		Set(kling.ParamPrompt, "人物面对镜头，微笑").
		Set(kling.ParamInputReference, DemoVideoURLs.RunningMan).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1080x1920)     // 竖屏 9:16
	results5, err := xai.Call(ctx, svc, op5, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-1", "img2video-vertical", results5)

	// =========================================================================
	// 6. img2video - 方形视频
	// =========================================================================
	fmt.Println("--- img2video (方形) ---")
	op6, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
	op6.Params().
		Set(kling.ParamPrompt, "产品展示，360度旋转").
		Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1080x1080)     // 方形 1:1
	results6, _ := xai.Call(ctx, svc, op6, svc.Options(), nil)
	printVideoResults("kling-v2-1", "img2video-square", results6)
}
