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

// Run: go run ./examples/kling/video kling-v2-6
//
// kling-v2-6 是可灵 V2.6 视频生成模型，支持以下功能：
//
// 1. text2video      - 纯文本生成视频
// 2. img2video       - 单图参考生成视频
// 3. keyframe        - 首尾帧生成视频
// 4. motion_control  - 动作控制（人物+动作视频）
// 5. sound_video     - 有声视频生成
//
// 动作控制（motion_control）说明：
//   - image_url: 人物图片 URL
//   - video_url: 动作参考视频 URL
//   - character_orientation: 人物朝向 "image"（跟随图片）或 "video"（跟随视频）
//   - keep_original_sound: "yes" 保留原视频声音，"no" 不保留
//
// 参数说明：
//   - input_reference: 参考图片 URL（用于 img2video）
//   - image_tail: 尾帧图片 URL（用于首尾帧模式）
//   - negative_prompt: 负向提示词
//   - sound: "on" 启用音频生成，"off" 关闭
//   - mode: "std"（标准 720P）或 "pro"（专家 1080P）
//   - seconds: "5" 或 "10"
//   - size: 视频尺寸
//
// 注意：V2.6 到 V2.9 使用相同的参数结构，可替换模型名称使用
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunKlingV26 runs all kling-v2-6 demos.
func RunKlingV26() {
	svc, err := shared.NewService()
	if err != nil {
		fmt.Println("Error:", err)
	}
	ctx := context.Background()

	// =========================================================================
	// 1. text2video - 纯文本生成视频
	// =========================================================================
	fmt.Println("--- text2video ---")
	op1, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op1.Params().
		Set(kling.ParamPrompt, "一只可爱的橘猫在阳光下奔跑，慢镜头，电影质感").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results1, err := xai.Call(ctx, svc, op1, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-6", "text2video", results1)

	// =========================================================================
	// 2. text2video - 带负向提示词
	// =========================================================================
	fmt.Println("--- text2video (negative_prompt) ---")
	op2, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op2.Params().
		Set(kling.ParamPrompt, "城市夜景，霓虹灯闪烁，赛博朋克风格").
		Set(kling.ParamNegativePrompt, "blurry, low quality, distorted, ugly").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results2, err := xai.Call(ctx, svc, op2, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-6", "text2video-negative", results2)

	// =========================================================================
	// 3. img2video - 单图参考生成视频

	// =========================================================================
	fmt.Println("--- img2video ---")
	op3, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op3.Params().
		Set(kling.ParamPrompt, "让图片中的角色动起来，镜头缓慢右移").
		Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results3, err := xai.Call(ctx, svc, op3, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-6", "img2video", results3)

	// =========================================================================
	// 4. keyframe - 首尾帧生成视频
	// =========================================================================
	fmt.Println("--- keyframe (首尾帧) ---")
	op4, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op4.Params().
		Set(kling.ParamPrompt, "平滑过渡，从白天到黑夜").
		Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame). // 首帧
		Set(kling.ParamImageTail, DemoVideoURLs.EndFrame).        // 尾帧
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results4, err := xai.Call(ctx, svc, op4, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-6", "keyframe", results4)

	// =========================================================================
	// 5. motion_control - 动作控制（人物朝向跟随图片）
	// =========================================================================
	// 将人物图片与动作视频结合，生成人物执行该动作的视频
	// character_orientation="image" 表示人物朝向跟随原图
	fmt.Println("--- motion_control (image orientation) ---")
	op5, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op5.Params().
		Set(kling.ParamPrompt, "女孩穿着灰色宽松T恤和牛仔短裤").
		Set(kling.ParamImageUrl, DemoVideoURLs.MotionImage). // 人物图片
		Set(kling.ParamVideoUrl, DemoVideoURLs.MotionVideo). // 动作视频
		Set(kling.ParamCharacterOrientation, "image").       // 朝向跟随图片
		Set(kling.ParamKeepOriginalSound, "yes").            // 保留原声
		Set(kling.ParamMode, kling.ModePro)
	results5, err := xai.Call(ctx, svc, op5, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-6", "motion_control-image", results5)

	// =========================================================================
	// 6. motion_control - 动作控制（人物朝向跟随视频）
	// =========================================================================
	// character_orientation="video" 表示人物朝向跟随动作视频
	fmt.Println("--- motion_control (video orientation) ---")
	op6, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op6.Params().
		Set(kling.ParamPrompt, "女孩在跳舞").
		Set(kling.ParamImageUrl, DemoVideoURLs.MotionImage).
		Set(kling.ParamVideoUrl, DemoVideoURLs.MotionVideo).
		Set(kling.ParamCharacterOrientation, "video"). // 朝向跟随视频
		Set(kling.ParamKeepOriginalSound, "no").       // 不保留原声
		Set(kling.ParamMode, kling.ModePro)
	results6, err := xai.Call(ctx, svc, op6, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-6", "motion_control-video", results6)

	// =========================================================================
	// 7. sound_video - 有声视频生成
	// =========================================================================
	// sound="on" 启用音频生成
	fmt.Println("--- sound_video ---")
	op7, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op7.Params().
		Set(kling.ParamPrompt, "一个人在演讲，声音洪亮").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSound, kling.SoundOn). // 启用音频
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results7, err := xai.Call(ctx, svc, op7, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-6", "sound_video", results7)

	// =========================================================================
	// 8. img2video + sound - 图片参考 + 有声视频
	// =========================================================================
	fmt.Println("--- img2video + sound ---")
	op8, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op8.Params().
		Set(kling.ParamPrompt, "人物开始说话，表情生动").
		Set(kling.ParamInputReference, DemoVideoURLs.RunningMan).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSound, kling.SoundOn).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1920x1080)
	results8, err := xai.Call(ctx, svc, op8, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-6", "img2video-sound", results8)

	// =========================================================================
	// 9. 竖屏视频 - 适合短视频平台
	// =========================================================================
	fmt.Println("--- vertical video ---")
	op9, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op9.Params().
		Set(kling.ParamPrompt, "一个人在跳街舞，动感十足").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5).
		Set(kling.ParamSize, kling.Size1080x1920) // 竖屏 9:16
	results9, err := xai.Call(ctx, svc, op9, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-6", "vertical-video", results9)

	// =========================================================================
	// 10. 长视频 - 10秒视频
	// =========================================================================
	fmt.Println("--- long video (10s) ---")
	op10, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op10.Params().
		Set(kling.ParamPrompt, "延时摄影，云层快速流动，日出到日落").
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds10). // 10秒视频
		Set(kling.ParamSize, kling.Size1920x1080)
	results10, err := xai.Call(ctx, svc, op10, svc.Options(), nil)
	if err != nil {
		fmt.Println("Error:", err)
	}
	printVideoResults("kling-v2-6", "long-video", results10)
}
