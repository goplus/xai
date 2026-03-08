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

// Run: go test ./examples/kling/images -run Example
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

type mockImageExecutor struct {
	images []string
}

func (m *mockImageExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	return &kling.SyncOperationResponse{R: kling.NewOutputImages(m.images)}, nil
}

func (m *mockImageExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	return &kling.SyncOperationResponse{R: kling.NewOutputImages(m.images)}, nil
}

type mockVideoExecutor struct{}

func (m *mockVideoExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	return &kling.SyncOperationResponse{R: kling.NewOutputVideos([]string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"})}, nil
}

func (m *mockVideoExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	return &kling.SyncOperationResponse{R: kling.NewOutputVideos([]string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"})}, nil
}

// -----------------------------------------------------------------------------
// Example tests by model
// -----------------------------------------------------------------------------

func Example_klingV1() {
	imgExec := &mockImageExecutor{images: []string{"https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"}}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV1), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "一只可爱的橘猫坐在窗台上看着夕阳,照片风格,高清画质")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		fmt.Println(out.Image.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png
}

func Example_klingV15() {
	imgExec := &mockImageExecutor{images: []string{"https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"}}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV15), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "一只可爱的橘猫坐在窗台上看着夕阳,照片风格,高清画质")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		fmt.Println(out.Image.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png
}

func Example_klingV15Image2Image() {
	imgExec := &mockImageExecutor{images: []string{"https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg"}}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV15), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "一个穿着西装的商务人士")
	op.Params().Set(kling.ParamImage, DemoImageURLs.Subject1)
	op.Params().Set(kling.ParamImageReference, kling.ImageRefSubject)
	op.Params().Set(kling.ParamImageFidelity, 0.7)
	op.Params().Set(kling.ParamHumanFidelity, 0.6)
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect1x1)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		fmt.Println(out.Image.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg
}

func Example_klingV2() {
	imgExec := &mockImageExecutor{images: []string{"https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"}}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV2), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "一只可爱的橘猫坐在窗台上看着夕阳,照片风格,高清画质")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		fmt.Println(out.Image.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png
}

func Example_klingV2New() {
	imgExec := &mockImageExecutor{images: []string{"https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg"}}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV2New), xai.GenImage)
	op.Params().Set(kling.ParamImage, DemoImageURLs.RunningMan)
	op.Params().Set(kling.ParamPrompt, "将这张图片转换为赛博朋克风格")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		fmt.Println(out.Image.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-image/image-to-image-with-mask-1.jpg
}

func Example_klingV21() {
	imgExec := &mockImageExecutor{images: []string{"https://huggingface.co/datasets/huggingface/documentation-images/resolve/4a5c8349eb8172fff604d547dc4991fbab6078e3/diffusers/controlnet-img2img.jpg"}}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "a sunset over the ocean, cinematic lighting")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	op.Params().Set(kling.ParamReferenceImages, []string{DemoImageURLs.RefStyle})
	op.Params().Set(kling.ParamNegativePrompt, "blurry, low quality")

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		fmt.Println(out.Image.StgUri())
	}
	// Output: https://huggingface.co/datasets/huggingface/documentation-images/resolve/4a5c8349eb8172fff604d547dc4991fbab6078e3/diffusers/controlnet-img2img.jpg
}

func Example_klingImageO1() {
	imgExec := &mockImageExecutor{images: []string{"https://huggingface.co/datasets/huggingface/documentation-images/resolve/4a5c8349eb8172fff604d547dc4991fbab6078e3/diffusers/landscape.png"}}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingImageO1), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "a serene mountain landscape at sunset, cinematic")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	op.Params().Set(kling.ParamResolution, kling.Resolution2K)

	ctx := context.Background()
	results, _ := xai.Call(ctx, svc, op, svc.Options(), nil)
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		fmt.Println(out.Image.StgUri())
	}
	// Output: https://huggingface.co/datasets/huggingface/documentation-images/resolve/4a5c8349eb8172fff604d547dc4991fbab6078e3/diffusers/landscape.png
}
