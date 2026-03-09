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

// Run: go run ./examples/kling/video kling-video-o1
//
// kling-video-o1 是可灵多参考视频生成模型，支持以下功能：
//
// 1. text2video    - 纯文本生成视频
// 2. img2video     - 单图参考生成视频（普通参考图）
// 3. keyframe      - 首尾帧生成视频（首帧 + 尾帧）
// 4. multi_ref     - 多图参考生成视频（混合普通参考图 + 首帧）
// 5. video2video   - 视频参考生成视频（feature/base 模式）
//
// image_list 参数说明：
//   - 每个 ImageInput 包含 Image（URL/Base64）和可选的 Type 字段
//   - Type 可选值：
//   - "" (ImageTypeRef): 普通参考图（主体、场景、风格等）
//   - "first_frame" (ImageTypeFirstFrame): 首帧
//   - "end_frame" (ImageTypeEndFrame): 尾帧
//   - 首尾帧规则：有尾帧时必须有首帧，暂不支持仅尾帧
//   - 数量限制：有参考视频时最多 4 张，无参考视频时最多 7 张
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"

	"github.com/goplus/xai/examples/shared"
)

// RunKlingVideoO1 runs all kling-video-o1 demos.
func RunKlingVideoO1() {
	svc, _ := shared.NewService()
	ctx := context.Background()

	// =========================================================================
	// 1. text2video - 纯文本生成视频
	// =========================================================================
	// 仅使用 prompt 描述，无需任何参考图/视频
	fmt.Println("--- text2video ---")
	op1, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
	op1.Params().
		Set(kling.ParamPrompt, "一只可爱的橘猫在阳光下奔跑，慢镜头，电影质感").
		Set(kling.ParamSize, kling.Size1920x1080). // 横屏 16:9
		Set(kling.ParamMode, kling.ModePro).       // 专家模式 1080P
		Set(kling.ParamSeconds, kling.Seconds5)    // 5秒视频
	results1, _ := xai.Call(ctx, svc, op1, svc.Options(), nil)
	printVideoResults("kling-video-o1", "text2video", results1)

	// =========================================================================
	// 2. img2video - 单图参考生成视频（普通参考图）
	// =========================================================================
	// 使用普通参考图（主体/场景/风格），Type 为空或 ImageTypeRef
	fmt.Println("--- img2video (普通参考图) ---")
	op2, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
	op2.Params().
		Set(kling.ParamPrompt, "这个人在跑马拉松").
		Set(kling.ParamImageList, []kling.ImageInput{
			{Image: DemoVideoURLs.RunningMan}, // Type 为空 = 普通参考图
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro)
	results2, _ := xai.Call(ctx, svc, op2, svc.Options(), nil)
	printVideoResults("kling-video-o1", "img2video-ref", results2)

	// =========================================================================
	// 3. img2video - 单图首帧生成视频
	// =========================================================================
	// 使用首帧图，视频从该图开始
	fmt.Println("--- img2video (首帧) ---")
	op3, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
	op3.Params().
		Set(kling.ParamPrompt, "镜头缓慢右移，人物开始奔跑").
		Set(kling.ParamImageList, []kling.ImageInput{
			{Image: DemoVideoURLs.FirstFrame, Type: kling.ImageTypeFirstFrame},
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5)
	results3, _ := xai.Call(ctx, svc, op3, svc.Options(), nil)
	printVideoResults("kling-video-o1", "img2video-first", results3)

	// =========================================================================
	// 4. keyframe - 首尾帧生成视频
	// =========================================================================
	// 同时指定首帧和尾帧，生成平滑过渡视频
	// 注意：有尾帧时必须有首帧
	fmt.Println("--- keyframe (首尾帧) ---")
	op4, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
	op4.Params().
		Set(kling.ParamPrompt, "视频连贯过渡").
		Set(kling.ParamImageList, []kling.ImageInput{
			{Image: DemoVideoURLs.FirstFrame, Type: kling.ImageTypeFirstFrame},
			{Image: DemoVideoURLs.EndFrame, Type: kling.ImageTypeEndFrame},
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro)
	results4, _ := xai.Call(ctx, svc, op4, svc.Options(), nil)
	printVideoResults("kling-video-o1", "keyframe", results4)

	// =========================================================================
	// 5. multi_ref - 多图参考生成视频
	// =========================================================================
	// 混合使用普通参考图 + 首帧，融合多种风格/主体
	// video_mode 设为 "multi_ref" 启用多参考模式
	fmt.Println("--- multi_ref (多图参考) ---")
	op5, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
	op5.Params().
		Set(kling.ParamPrompt, "融合两张图的风格，创建连贯场景").
		Set(kling.ParamImageList, []kling.ImageInput{
			{Image: DemoVideoURLs.MultiRef1},                                  // 普通参考图（风格/场景）
			{Image: DemoVideoURLs.MultiRef2, Type: kling.ImageTypeFirstFrame}, // 首帧
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro).
		Set(kling.ParamSeconds, kling.Seconds5)
	results5, _ := xai.Call(ctx, svc, op5, svc.Options(), nil)
	printVideoResults("kling-video-o1", "multi_ref", results5)

	// =========================================================================
	// 6. video2video - 视频参考生成视频 (feature 模式)
	// =========================================================================
	// refer_type: "feature" - 提取视频特征，生成风格相似的新视频
	// keep_original_sound: "yes" - 保留原视频声音
	fmt.Println("--- video2video (feature) ---")
	op6, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
	op6.Params().
		Set(kling.ParamPrompt, "画面中增加一个小狗").
		Set(kling.ParamVideoList, []kling.VideoRef{
			{
				VideoURL:          DemoVideoURLs.VideoFeature,
				ReferType:         kling.VideoReferTypeFeature, // 特征参考
				KeepOriginalSound: kling.KeepOriginalSoundYes,  // 保留原声
			},
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro)
	results6, _ := xai.Call(ctx, svc, op6, svc.Options(), nil)
	printVideoResults("kling-video-o1", "video2video-feature", results6)

	// =========================================================================
	// 7. video2video - 视频参考生成视频 (base 模式)
	// =========================================================================
	// refer_type: "base" - 以视频为基础，进行风格转换或内容编辑
	fmt.Println("--- video2video (base) ---")
	op7, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
	op7.Params().
		Set(kling.ParamPrompt, "将视频转换为动漫风格").
		Set(kling.ParamVideoList, []kling.VideoRef{
			{
				VideoURL:  DemoVideoURLs.VideoFeature,
				ReferType: kling.VideoReferTypeBase, // 基础参考
			},
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro)
	results7, _ := xai.Call(ctx, svc, op7, svc.Options(), nil)
	printVideoResults("kling-video-o1", "video2video-base", results7)

	// =========================================================================
	// 8. 混合模式 - 图片 + 视频参考
	// =========================================================================
	// 同时使用 image_list 和 video_list
	fmt.Println("--- mixed (图片+视频参考) ---")
	op8, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
	op8.Params().
		Set(kling.ParamPrompt, "将人物融入视频场景").
		Set(kling.ParamImageList, []kling.ImageInput{
			{Image: DemoVideoURLs.RunningMan}, // 人物参考图
		}).
		Set(kling.ParamVideoList, []kling.VideoRef{
			{VideoURL: DemoVideoURLs.VideoFeature, ReferType: "feature"},
		}).
		Set(kling.ParamSize, kling.Size1920x1080).
		Set(kling.ParamMode, kling.ModePro)
	results8, _ := xai.Call(ctx, svc, op8, svc.Options(), nil)
	printVideoResults("kling-video-o1", "mixed", results8)
}
