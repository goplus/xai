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

// Run: go test ./examples/kling -run Example
package main

import (
	"context"
	"errors"
	"fmt"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling"
	"github.com/goplus/xai/spec/kling/provider/qiniu"
)

// -----------------------------------------------------------------------------
// Mock executors for testing
// -----------------------------------------------------------------------------

type mockImageExecutor struct {
	images    []string
	async     bool
	pollCount int
	fail      bool
}

func (m *mockImageExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	if m.async {
		return kling.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
			return m.GetTaskStatus(ctx, "task-async-1")
		}, "task-async-1"), nil
	}
	return &kling.SyncOperationResponse{R: kling.NewOutputImages(m.images)}, nil
}

func (m *mockImageExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	if m.fail {
		return nil, fmt.Errorf("%w: content policy violation", qiniu.ErrTaskFailed)
	}
	m.pollCount++
	if m.pollCount >= 2 {
		return &kling.SyncOperationResponse{R: kling.NewOutputImages(m.images)}, nil
	}
	return kling.NewAsyncOperationResponse(func(ctx context.Context) (xai.OperationResponse, error) {
		return m.GetTaskStatus(ctx, taskID)
	}, taskID), nil
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
// Example: Basic sync image generation
// -----------------------------------------------------------------------------

func Example_basicImageGen() {
	imgExec := &mockImageExecutor{images: []string{"https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"}}
	vidExec := &mockVideoExecutor{videoURLs: []string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "a sunset over the ocean")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)
	op.Params().Set(kling.ParamN, 1)

	ctx := context.Background()
	results, err := xai.Call(ctx, svc, op, svc.Options(), nil)
	if err != nil {
		return
	}
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		fmt.Println(out.Image.StgUri())
	}
	// Output: https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png
}

// -----------------------------------------------------------------------------
// Example: Basic sync video generation (img2video)
// -----------------------------------------------------------------------------

func Example_basicVideoGen() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{videoURLs: []string{"https://aitoken-public.qnaigc.com/example/generate-video/the-little-dog-is-running-on-the-lawn.mp4"}}
	svc := kling.NewService(imgExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
	op.Params().Set(kling.ParamPrompt, "camera pans slowly to the right")
	op.Params().Set(kling.ParamInputReference, "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg")
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

// -----------------------------------------------------------------------------
// Example: Image params by model
// -----------------------------------------------------------------------------

func Example_imageParamsByModel() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(imgExec, vidExec)

	opV1, _ := svc.Operation(xai.Model(kling.ModelKlingV1), xai.GenImage)
	opV1.Params().Set(kling.ParamPrompt, "a cat")
	opV1.Params().Set(kling.ParamImage, "https://ref.png")
	fmt.Println("kling-v1 schema:", len(opV1.InputSchema().Fields()), "fields")

	opV2, _ := svc.Operation(xai.Model(kling.ModelKlingV2), xai.GenImage)
	opV2.Params().Set(kling.ParamPrompt, "a cat")
	opV2.Params().Set(kling.ParamSubjectImageList, []map[string]string{{kling.ParamSubjectImage: "https://ref1.png"}, {kling.ParamSubjectImage: "https://ref2.png"}})
	fmt.Println("kling-v2 schema:", len(opV2.InputSchema().Fields()), "fields")

	opV21, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
	opV21.Params().Set(kling.ParamPrompt, "a cat")
	opV21.Params().Set(kling.ParamReferenceImages, []string{"https://ref1.png"})
	fmt.Println("kling-v2-1 schema:", len(opV21.InputSchema().Fields()), "fields")

	opO1, _ := svc.Operation(xai.Model(kling.ModelKlingImageO1), xai.GenImage)
	opO1.Params().Set(kling.ParamPrompt, "a cat")
	opO1.Params().Set(kling.ParamResolution, kling.Resolution2K)
	fmt.Println("kling-image-o1 schema:", len(opO1.InputSchema().Fields()), "fields")

	// Output:
	// kling-v1 schema: 7 fields
	// kling-v2 schema: 8 fields
	// kling-v2-1 schema: 9 fields
	// kling-image-o1 schema: 5 fields
}

// -----------------------------------------------------------------------------
// Example: Video params by model
// -----------------------------------------------------------------------------

func Example_videoParamsByModel() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(imgExec, vidExec)

	opV21, _ := svc.Operation(xai.Model(kling.ModelKlingV21Video), xai.GenVideo)
	opV21.Params().Set(kling.ParamPrompt, "a scene")
	opV21.Params().Set(kling.ParamInputReference, "https://first.png")
	opV21.Params().Set(kling.ParamImageTail, "https://last.png")
	fmt.Println("kling-v2-1:", len(opV21.InputSchema().Fields()), "fields")

	opO1, _ := svc.Operation(xai.Model(kling.ModelKlingVideoO1), xai.GenVideo)
	fmt.Println("kling-video-o1:", len(opO1.InputSchema().Fields()), "fields")

	opV26, _ := svc.Operation(xai.Model(kling.ModelKlingV26), xai.GenVideo)
	opV26.Params().Set(kling.ParamPrompt, "a scene")
	opV26.Params().Set(kling.ParamSound, kling.SoundOn)
	fmt.Println("kling-v2-6:", len(opV26.InputSchema().Fields()), "fields")

	// Output:
	// kling-v2-1: 7 fields
	// kling-video-o1: 9 fields
	// kling-v2-6: 12 fields
}

// -----------------------------------------------------------------------------
// Example: Async polling with progress
// -----------------------------------------------------------------------------

func Example_asyncWithProgress() {
	asyncExec := &mockImageExecutor{
		images: []string{"https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png"},
		async:  true,
	}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(asyncExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "async generation")
	op.Params().Set(kling.ParamAspectRatio, kling.Aspect16x9)

	ctx := context.Background()
	resp, _ := op.Call(ctx, svc, svc.Options())
	results, _ := xai.Wait(ctx, svc, resp, func(r xai.OperationResponse) {
		if !r.Done() {
			fmt.Println("polling...")
		}
	})
	fmt.Println("done:", results.Len(), "image(s)")
	for i := 0; i < results.Len(); i++ {
		out := results.At(i).(*xai.OutputImage)
		fmt.Println(out.Image.StgUri())
	}
	// Output:
	// polling...
	// polling...
	// done: 1 image(s)
	// https://aitoken-public.qnaigc.com/example/generate-image/smile-woman.png
}

// -----------------------------------------------------------------------------
// Example: List models and actions
// -----------------------------------------------------------------------------

func Example_listModelsAndActions() {
	imgExec := &mockImageExecutor{}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(imgExec, vidExec)

	fmt.Println("Image models:", kling.ImageModels())
	fmt.Println("Video models:", kling.VideoModels())
	actions := svc.Actions(xai.Model(kling.ModelKlingV21))
	fmt.Println("kling-v2-1 actions:", actions)
	actionsO1 := svc.Actions(xai.Model(kling.ModelKlingVideoO1))
	fmt.Println("kling-video-o1 actions:", actionsO1)

	// Output:
	// Image models: [kling-v1 kling-v1-5 kling-v2 kling-v2-new kling-v2-1 kling-image-o1]
	// Video models: [kling-v2-1 kling-v2-5-turbo kling-video-o1 kling-v2-6 kling-v2-7 kling-v2-8 kling-v2-9 kling-v3 kling-v3-omni]
	// kling-v2-1 actions: [gen_image gen_video]
	// kling-video-o1 actions: [gen_video]
}

// -----------------------------------------------------------------------------
// Example: Task failure with errors.Is
// -----------------------------------------------------------------------------

func Example_taskFailureError() {
	failingExec := &mockImageExecutor{async: true, fail: true}
	vidExec := &mockVideoExecutor{}
	svc := kling.NewService(failingExec, vidExec)

	op, _ := svc.Operation(xai.Model(kling.ModelKlingV21), xai.GenImage)
	op.Params().Set(kling.ParamPrompt, "test")
	ctx := context.Background()

	resp, _ := op.Call(ctx, svc, svc.Options())
	_, err := xai.Wait(ctx, svc, resp, nil)
	if err != nil && errors.Is(err, qiniu.ErrTaskFailed) {
		fmt.Println("task failed (detected via errors.Is)")
	}
	// Output: task failed (detected via errors.Is)
}
