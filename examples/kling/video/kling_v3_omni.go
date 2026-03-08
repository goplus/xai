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

// Run: go run ./examples/kling/video kling-v3-omni
//
// kling-v3-omni 是可灵全能视频生成模型，支持以下功能：
//
// 1. text2video    - 纯文本生成视频
// 2. img2video     - 图片参考生成视频（普通参考图/首帧/首尾帧）
// 3. video2video   - 视频参考生成视频（feature/base 模式）
// 4. sound_video   - 有声视频生成（sound="on"）
// 5. multi_shot    - 多镜头分段视频（multi_shot=true + multi_prompt）
//
// image_list 参数说明：
//   - 每个 ImageInput 包含 Image（URL/Base64）和可选的 Type 字段
//   - Type 可选值：
//   - "" (ImageTypeRef): 普通参考图（主体、场景、风格等）
//   - "first_frame" (ImageTypeFirstFrame): 首帧
//   - "end_frame" (ImageTypeEndFrame): 尾帧
//
// multi_shot 参数说明：
//   - multi_shot: true 启用多镜头模式
//   - shot_type: "auto"（自动）或 "customize"（自定义）
//   - multi_prompt: 分段提示词列表，每段包含 index、prompt、duration
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunKlingV3Omni runs all kling-v3-omni demos.
func RunKlingV3Omni() {
	svc, _ := shared.NewService()
	ctx := context.Background()

	// =========================================================================
	// 8. multi_shot - 多镜头分段视频（自定义分镜）
	// =========================================================================
	// shot_type="customize" 时必须提供 multi_prompt，
	// 且各段 duration 之和应等于 seconds（此处 4+3+3=10）。
	fmt.Println("--- multi_shot (customize) ---")
	op8, _ := svc.Operation(xai.Model(kling.ModelKlingV3Omni), xai.GenVideo)
	op8.Params().
		Set(kling.ParamPrompt, "产品展示视频").
		Set(kling.ParamMultiShot, true).
		Set(kling.ParamShotType, "customize"). // 自定义分镜
		Set(kling.ParamMultiPrompt, []kling.MultiPromptItem{
			{Index: 1, Prompt: "产品全景展示，旋转", Duration: "4"},
			{Index: 2, Prompt: "产品细节特写", Duration: "3"},
			{Index: 3, Prompt: "产品使用场景", Duration: "3"},
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds10)
	results8, _ := xai.Call(ctx, svc, op8, svc.Options(), nil)
	printVideoResults("kling-v3-omni", "multi_shot-customize", results8)

	// =========================================================================
	// 1. text2video - 纯文本生成视频
	// =========================================================================
	fmt.Println("--- text2video ---")
	op1, _ := svc.Operation(xai.Model(kling.ModelKlingV3Omni), xai.GenVideo)
	op1.Params().
		Set(kling.ParamPrompt, "一只可爱的橘猫在阳光下奔跑，慢镜头，电影质感").
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5)
	results1, _ := xai.Call(ctx, svc, op1, svc.Options(), nil)
	printVideoResults("kling-v3-omni", "text2video", results1)

	// =========================================================================
	// 2. img2video - 普通参考图生成视频
	// =========================================================================
	// Type 为空 = 普通参考图（主体/场景/风格）
	fmt.Println("--- img2video (普通参考图) ---")
	op2, _ := svc.Operation(xai.Model(kling.ModelKlingV3Omni), xai.GenVideo)
	op2.Params().
		Set(kling.ParamPrompt, "这个人在跑马拉松").
		Set(kling.ParamImageList, []kling.ImageInput{
			{Image: DemoVideoURLs.RunningMan}, // 普通参考图
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro)
	results2, _ := xai.Call(ctx, svc, op2, svc.Options(), nil)
	printVideoResults("kling-v3-omni", "img2video-ref", results2)

	// =========================================================================
	// 3. img2video - 首帧生成视频
	// =========================================================================
	fmt.Println("--- img2video (首帧) ---")
	op3, _ := svc.Operation(xai.Model(kling.ModelKlingV3Omni), xai.GenVideo)
	op3.Params().
		Set(kling.ParamPrompt, "镜头缓慢推进，人物开始动作").
		Set(kling.ParamImageList, []kling.ImageInput{
			{Image: DemoVideoURLs.FirstFrame, Type: kling.ImageTypeFirstFrame},
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro)
	results3, _ := xai.Call(ctx, svc, op3, svc.Options(), nil)
	printVideoResults("kling-v3-omni", "img2video-first", results3)

	// =========================================================================
	// 4. keyframe - 首尾帧生成视频
	// =========================================================================
	fmt.Println("--- keyframe (首尾帧) ---")
	op4, _ := svc.Operation(xai.Model(kling.ModelKlingV3Omni), xai.GenVideo)
	op4.Params().
		Set(kling.ParamPrompt, "视频连贯过渡").
		Set(kling.ParamImageList, []kling.ImageInput{
			{Image: DemoVideoURLs.FirstFrame, Type: kling.ImageTypeFirstFrame},
			{Image: DemoVideoURLs.EndFrame, Type: kling.ImageTypeEndFrame},
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro)
	results4, _ := xai.Call(ctx, svc, op4, svc.Options(), nil)
	printVideoResults("kling-v3-omni", "keyframe", results4)

	// =========================================================================
	// 5. video2video - 视频参考生成视频
	// =========================================================================
	fmt.Println("--- video2video ---")
	op5, _ := svc.Operation(xai.Model(kling.ModelKlingV3Omni), xai.GenVideo)
	op5.Params().
		Set(kling.ParamPrompt, "画面中增加一个小狗").
		Set(kling.ParamVideoList, []kling.VideoRef{
			{
				VideoURL:          DemoVideoURLs.VideoFeature,
				ReferType:         "feature",
				KeepOriginalSound: "yes",
			},
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro)
	results5, _ := xai.Call(ctx, svc, op5, svc.Options(), nil)
	printVideoResults("kling-v3-omni", "video2video", results5)

	// =========================================================================
	// 6. sound_video - 有声视频生成
	// =========================================================================
	// sound="on" 启用音频生成
	fmt.Println("--- sound_video ---")
	op6, _ := svc.Operation(xai.Model(kling.ModelKlingV3Omni), xai.GenVideo)
	op6.Params().
		Set(kling.ParamPrompt, "一个人在演讲，声音洪亮").
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSound, kling.SoundOn) // 启用音频
	results6, _ := xai.Call(ctx, svc, op6, svc.Options(), nil)
	printVideoResults("kling-v3-omni", "sound_video", results6)

	// =========================================================================
	// 7. multi_shot - 多镜头分段视频（自动分镜）
	// =========================================================================
	// shot_type="auto" 让模型自动分配各段时长
	fmt.Println("--- multi_shot (auto) ---")
	op7, _ := svc.Operation(xai.Model(kling.ModelKlingV3Omni), xai.GenVideo)
	op7.Params().
		Set(kling.ParamPrompt, "一部微电影").
		Set(kling.ParamMultiShot, true).  // 启用多镜头
		Set(kling.ParamShotType, "auto"). // 自动分镜
		Set(kling.ParamMultiPrompt, []kling.MultiPromptItem{
			{Index: 1, Prompt: "清晨，阳光洒进房间", Duration: "3"},
			{Index: 2, Prompt: "主角起床，伸懒腰", Duration: "2"},
			{Index: 3, Prompt: "主角走出门，迎接新的一天", Duration: "3"},
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro)
	results7, _ := xai.Call(ctx, svc, op7, svc.Options(), nil)
	printVideoResults("kling-v3-omni", "multi_shot-auto", results7)

}
