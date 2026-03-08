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

// Run: go test ./examples/kling/video -run Example
package main

import (
	"context"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"
)

// -----------------------------------------------------------------------------
// Mock executors for example tests
// -----------------------------------------------------------------------------

type mockImageExecutor struct{}

func (m *mockImageExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	return &kling.SyncOperationResponse{R: kling.NewOutputImages([]string{"https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"})}, nil
}

func (m *mockImageExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	return &kling.SyncOperationResponse{R: kling.NewOutputImages([]string{"https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"})}, nil
}

type mockVideoExecutor struct {
	videoURLs []string
}

func (m *mockVideoExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	return &kling.SyncOperationResponse{R: kling.NewOutputVideos(m.videoURLs)}, nil
}

func (m *mockVideoExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	return &kling.SyncOperationResponse{R: kling.NewOutputVideos(m.videoURLs)}, nil
}

// -----------------------------------------------------------------------------
// Example tests by model
// -----------------------------------------------------------------------------

func Example_klingV21() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{videoURLs: []string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
	op.Params().Set(kling.ParamPrompt, "camera pans slowly to the right, cinematic")
	op.Params().Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame)
	op.Params().Set(kling.ParamMode, kling.ModePro)
	op.Params().Set(kling.ParamSeconds, kling.Seconds5)
	op.Params().Set(kling.ParamSize, kling.Size1920x1080)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Println(out.Video.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4
}

func Example_klingV25Turbo() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{videoURLs: []string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV25Turbo), xai.GenVideo)
	op.Params().Set(kling.ParamPrompt, "a hero enters the battlefield, dramatic lighting")
	op.Params().Set(kling.ParamMode, kling.ModePro)
	op.Params().Set(kling.ParamSeconds, kling.Seconds5)
	op.Params().Set(kling.ParamSize, kling.Size1920x1080)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Println(out.Video.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4
}

func Example_klingV26() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{videoURLs: []string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op.Params().Set(kling.ParamPrompt, "一只可爱的橘猫在阳光下奔跑，慢镜头，电影质感")
	op.Params().Set(kling.ParamMode, kling.ModePro)
	op.Params().Set(kling.ParamSeconds, kling.Seconds5)
	op.Params().Set(kling.ParamSize, kling.Size1920x1080)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Println(out.Video.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4
}

func Example_klingV26Img2Video() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{videoURLs: []string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	op.Params().Set(kling.ParamPrompt, "让图片中的角色动起来，镜头缓慢右移")
	op.Params().Set(kling.ParamInputReference, DemoVideoURLs.FirstFrame)
	op.Params().Set(kling.ParamMode, kling.ModePro)
	op.Params().Set(kling.ParamSeconds, kling.Seconds5)
	op.Params().Set(kling.ParamSize, kling.Size1920x1080)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Println(out.Video.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4
}

func Example_klingVideoO1() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{videoURLs: []string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
	op.Params().Set(kling.ParamPrompt, "一只可爱的橘猫在阳光下奔跑")
	op.Params().Set(kling.ParamSize, kling.Size1920x1080)
	op.Params().Set(kling.ParamMode, kling.ModePro)
	op.Params().Set(kling.ParamSeconds, kling.Seconds5)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Println(out.Video.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4
}

func Example_klingV3() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{videoURLs: []string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV3), xai.GenVideo)
	op.Params().Set(kling.ParamPrompt, "一只可爱的小猫在阳光下玩耍")
	op.Params().Set(kling.ParamMode, kling.ModeStd)
	op.Params().Set(kling.ParamSeconds, kling.Seconds5)
	op.Params().Set(kling.ParamSize, kling.Size1280x720)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Println(out.Video.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4
}

func Example_klingV3Omni() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{videoURLs: []string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV3Omni), xai.GenVideo)
	op.Params().Set(kling.ParamPrompt, "这个人在跑马拉松")
	op.Params().Set(kling.ParamImageList, []kling.ImageInput{
		{Image: DemoVideoURLs.FirstFrame},
	})
	op.Params().Set(kling.ParamMode, kling.ModePro)
	op.Params().Set(kling.ParamSeconds, kling.Seconds5)
	op.Params().Set(kling.ParamSize, kling.Size1920x1080)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputVideo)
		fmt.Println(out.Video.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4
}
