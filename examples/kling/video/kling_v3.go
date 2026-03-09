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

// Run: go run ./examples/kling/video kling-v3
//
// kling-v3 是可灵 V3 视频生成模型，支持以下功能：
//
// 1. text2video    - 纯文本生成视频
// 2. img2video     - 图片参考生成视频（使用 input_reference）
// 3. sound_video   - 有声视频生成（sound="on"）
// 4. multi_shot    - 多镜头分段视频（multi_shot=true + shot_type=customize + multi_prompt）
//
// 参数说明：
//   - input_reference: 参考图片 URL（用于 img2video）
//   - sound: "on" 启用音频生成，"off" 关闭
//   - mode: "std"（标准 720P）或 "pro"（专家 1080P）
//   - seconds: "5" 或 "10"
//   - size: 视频尺寸（1920x1080, 1080x1920, 1280x720 等）
//   - multi_prompt: 分镜提示词列表（每项包含 index、prompt、duration）
//
// 注意：V3 不支持 image_list/video_list，使用 input_reference 传递单张参考图
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunKlingV3 runs all kling-v3 demos.
func RunKlingV3() {
	svc, _ := shared.NewService()
	ctx := context.Background()

	// =========================================================================
	// 1. text2video - 纯文本生成视频（标准模式）
	// =========================================================================
	fmt.Println("--- text2video (std) ---")
	op1, _ := svc.Operation(xai.Model(kling.ModelKlingV3), xai.GenVideo)
	op1.Params().
		Set(kling.ParamPrompt, "一只可爱的小猫在阳光下玩耍").
		Set(kling.ParamMode, kling.ModeStd). // 标准模式 720P
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1280x720) // 720P 横屏
	results1, _ := xai.Call(ctx, svc, op1, svc.Options(), nil)
	printVideoResults("kling-v3", "text2video-std", results1)

	// =========================================================================
	// 2. text2video - 纯文本生成视频（专家模式）
	// =========================================================================
	fmt.Println("--- text2video (pro) ---")
	op2, _ := svc.Operation(xai.Model(kling.ModelKlingV3), xai.GenVideo)
	op2.Params().
		Set(kling.ParamPrompt, "城市夜景，霓虹灯闪烁，电影质感").
		Set(kling.ParamMode, kling.ModePro). // 专家模式 1080P
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080) // 1080P 横屏
	results2, _ := xai.Call(ctx, svc, op2, svc.Options(), nil)
	printVideoResults("kling-v3", "text2video-pro", results2)

	// =========================================================================
	// 3. text2video - 竖屏视频（适合短视频平台）
	// =========================================================================
	fmt.Println("--- text2video (竖屏) ---")
	op3, _ := svc.Operation(xai.Model(kling.ModelKlingV3), xai.GenVideo)
	op3.Params().
		Set(kling.ParamPrompt, "一个人在跳舞，动感十足").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1080x1920) // 竖屏 9:16
	results3, _ := xai.Call(ctx, svc, op3, svc.Options(), nil)
	printVideoResults("kling-v3", "text2video-vertical", results3)

	// =========================================================================
	// 4. img2video - 图片参考生成视频
	// =========================================================================
	// 使用 input_reference 传递参考图
	fmt.Println("--- img2video ---")
	op4, _ := svc.Operation(xai.Model(kling.ModelKlingV3), xai.GenVideo)
	op4.Params().
		Set(kling.ParamPrompt, "让图片中的角色动起来").
		Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results4, _ := xai.Call(ctx, svc, op4, svc.Options(), nil)
	printVideoResults("kling-v3", "img2video", results4)

	// =========================================================================
	// 5. sound_video - 有声视频生成
	// =========================================================================
	// sound="on" 启用音频生成
	fmt.Println("--- sound_video ---")
	op5, _ := svc.Operation(xai.Model(kling.ModelKlingV3), xai.GenVideo)
	op5.Params().
		Set(kling.ParamPrompt, "一个人在演讲，声音洪亮").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSound, kling.SoundOn). // 启用音频
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results5, _ := xai.Call(ctx, svc, op5, svc.Options(), nil)
	printVideoResults("kling-v3", "sound_video", results5)

	// =========================================================================
	// 6. img2video + sound - 图片参考 + 有声视频
	// =========================================================================
	fmt.Println("--- img2video + sound ---")
	op6, _ := svc.Operation(xai.Model(kling.ModelKlingV3), xai.GenVideo)
	op6.Params().
		Set(kling.ParamPrompt, "人物开始说话，表情生动").
		Set(kling.ParamInputReference, DemoVideoURLs.RunningMan).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSound, kling.SoundOn).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results6, _ := xai.Call(ctx, svc, op6, svc.Options(), nil)
	printVideoResults("kling-v3", "img2video-sound", results6)

	// =========================================================================
	// 7. 长视频 - 10秒视频
	// =========================================================================
	fmt.Println("--- long video (10s) ---")
	op7, _ := svc.Operation(xai.Model(kling.ModelKlingV3), xai.GenVideo)
	op7.Params().
		Set(kling.ParamPrompt, "日出到日落的延时摄影，壮观的云层变化").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds10). // 10秒视频
		Set(kling.ParamSize, kling.Size1920x1080)
	results7, _ := xai.Call(ctx, svc, op7, svc.Options(), nil)
	printVideoResults("kling-v3", "long-video", results7)

	// =========================================================================
	// 8. multi_shot - 多镜头分段视频（自定义分镜）
	// =========================================================================
	// 当 multi_shot=true 且 shot_type=customize 时，multi_prompt 必填。
	// 所有分镜 duration 之和需等于 seconds（此处 3+4+3=10）。
	fmt.Println("--- multi_shot (customize) ---")
	op8, _ := svc.Operation(xai.Model(kling.ModelKlingV3), xai.GenVideo)
	op8.Params().
		Set(kling.ParamPrompt, "城市一天的节奏感短片").
		Set(kling.ParamMultiShot, true).
		Set(kling.ParamShotType, "customize").
		Set(kling.ParamMultiPrompt, []kling.MultiPromptItem{
			{Index: 1, Prompt: "清晨，阳光照进窗台，城市逐渐苏醒", Duration: "3"},
			{Index: 2, Prompt: "午后，街道人流与车流交织，节奏加快", Duration: "4"},
			{Index: 3, Prompt: "夜晚，霓虹点亮天际线，镜头拉远收束", Duration: "3"},
		}).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds10).
		Set(kling.ParamSize, kling.Size1920x1080)
	results8, _ := xai.Call(ctx, svc, op8, svc.Options(), nil)
	printVideoResults("kling-v3", "multi_shot-customize", results8)
}
