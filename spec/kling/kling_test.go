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

package kling

import (
	"context"
	"errors"
	"testing"

	xai "github.com/goplus/xai/spec"
)

// mockImageExecutor returns sync image results for demo/testing.
type mockImageExecutor struct {
	urls []string
}

func (m *mockImageExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	_ = params.(*Params).Export()
	return &SyncOperationResponse{R: NewOutputImages(m.urls)}, nil
}

func (m *mockImageExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	_ = taskID
	return nil, xai.ErrNotSupported
}

// mockVideoExecutor returns sync video results for demo/testing.
type mockVideoExecutor struct {
	urls []string
}

func (m *mockVideoExecutor) Submit(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	_ = params.(*Params).Export()
	return &SyncOperationResponse{R: NewOutputVideos(m.urls)}, nil
}

func (m *mockVideoExecutor) GetTaskStatus(ctx context.Context, taskID string) (xai.OperationResponse, error) {
	_ = taskID
	return nil, xai.ErrNotSupported
}

func TestKlingService_Actions(t *testing.T) {
	imgExec := &mockImageExecutor{urls: []string{"https://example.com/img.png"}}
	vidExec := &mockVideoExecutor{urls: []string{"https://example.com/vid.mp4"}}
	svc := NewService(imgExec, vidExec)

	actions := svc.Actions(xai.Model("kling-v2-1"))
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions for kling-v2-1, got %d", len(actions))
	}

	actions = svc.Actions(xai.Model("kling-video-o1"))
	if len(actions) != 1 || actions[0] != xai.GenVideo {
		t.Fatalf("expected [GenVideo] for kling-video-o1, got %v", actions)
	}
}

func TestKlingService_Operation_Call(t *testing.T) {
	imgExec := &mockImageExecutor{urls: []string{"https://example.com/out.png"}}
	vidExec := &mockVideoExecutor{urls: []string{"https://example.com/out.mp4"}}
	svc := NewService(imgExec, vidExec)

	op, err := svc.Operation(xai.Model("kling-v2-1"), xai.GenImage)
	if err != nil {
		t.Fatal(err)
	}
	op.Params().Set(ParamPrompt, "a sunset")
	op.Params().Set(ParamAspectRatio, Aspect16x9)

	ctx := context.Background()
	resp, err := op.Call(ctx, svc, &Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !resp.Done() {
		t.Fatal("expected sync response to be done")
	}
	results := resp.Results()
	if results.Len() != 1 {
		t.Fatalf("expected 1 image, got %d", results.Len())
	}
	out := results.At(0).(*xai.OutputImage)
	if out.Image.StgUri() != "https://example.com/out.png" {
		t.Fatalf("unexpected image URL: %s", out.Image.StgUri())
	}
}

func TestKlingService_Operation_Call_Validation(t *testing.T) {
	imgExec := &mockImageExecutor{urls: []string{"https://example.com/out.png"}}
	vidExec := &mockVideoExecutor{urls: []string{"https://example.com/out.mp4"}}
	svc := NewService(imgExec, vidExec)
	ctx := context.Background()

	// GenImage: missing prompt
	opImg, _ := svc.Operation(xai.Model("kling-v2-1"), xai.GenImage)
	_, err := opImg.Call(ctx, svc, &Options{})
	if !errors.Is(err, ErrPromptRequired) {
		t.Fatalf("expected ErrPromptRequired, got %v", err)
	}

	// GenImage: empty prompt
	opImg.Params().Set(ParamPrompt, "   ")
	_, err = opImg.Call(ctx, svc, &Options{})
	if !errors.Is(err, ErrPromptRequired) {
		t.Fatalf("expected ErrPromptRequired for empty prompt, got %v", err)
	}

	// GenVideo kling-v2-1: missing input_reference
	opVid, _ := svc.Operation(xai.Model("kling-v2-1"), xai.GenVideo)
	opVid.Params().Set(ParamPrompt, "a scene")
	_, err = opVid.Call(ctx, svc, &Options{})
	if !errors.Is(err, ErrInputReferenceRequired) {
		t.Fatalf("expected ErrInputReferenceRequired, got %v", err)
	}

	// GenVideo kling-v2-5-turbo: input_reference optional, prompt only is OK
	opVid25, _ := svc.Operation(xai.Model("kling-v2-5-turbo"), xai.GenVideo)
	opVid25.Params().Set(ParamPrompt, "a hero")
	resp, err := opVid25.Call(ctx, svc, &Options{})
	if err != nil {
		t.Fatalf("kling-v2-5-turbo text2video should not require input_reference: %v", err)
	}
	if !resp.Done() {
		t.Fatal("expected sync response")
	}

	// Keyframe: image_tail set but mode != "pro" → ErrKeyframeModeRequired
	opKf, _ := svc.Operation(xai.Model("kling-v2-1"), xai.GenVideo)
	opKf.Params().Set(ParamPrompt, "scene")
	opKf.Params().Set(ParamInputReference, "https://example.com/img.png")
	opKf.Params().Set(ParamImageTail, "https://example.com/tail.png")
	opKf.Params().Set(ParamMode, ModeStd)
	_, err = opKf.Call(ctx, svc, &Options{})
	if !errors.Is(err, ErrKeyframeModeRequired) {
		t.Fatalf("expected ErrKeyframeModeRequired when image_tail set and mode=std, got %v", err)
	}

	// Keyframe: image_tail set but mode empty → ErrKeyframeModeRequired
	opKf2, _ := svc.Operation(xai.Model("kling-v2-1"), xai.GenVideo)
	opKf2.Params().Set(ParamPrompt, "scene")
	opKf2.Params().Set(ParamInputReference, "https://example.com/img.png")
	opKf2.Params().Set(ParamImageTail, "https://example.com/tail.png")
	_, err = opKf2.Call(ctx, svc, &Options{})
	if !errors.Is(err, ErrKeyframeModeRequired) {
		t.Fatalf("expected ErrKeyframeModeRequired when image_tail set and mode empty, got %v", err)
	}

	// Keyframe kling-v2-1: seconds must be "10"
	opKf3, _ := svc.Operation(xai.Model("kling-v2-1"), xai.GenVideo)
	opKf3.Params().Set(ParamPrompt, "scene")
	opKf3.Params().Set(ParamInputReference, "https://example.com/img.png")
	opKf3.Params().Set(ParamImageTail, "https://example.com/tail.png")
	opKf3.Params().Set(ParamMode, ModePro)
	opKf3.Params().Set(ParamSeconds, Seconds5)
	_, err = opKf3.Call(ctx, svc, &Options{})
	if !errors.Is(err, ErrKeyframeSecondsRequired) {
		t.Fatalf("expected ErrKeyframeSecondsRequired when kling-v2-1 keyframe with seconds=5, got %v", err)
	}

	// Keyframe kling-v2-1: seconds missing → ErrKeyframeSecondsRequired
	opKf4, _ := svc.Operation(xai.Model("kling-v2-1"), xai.GenVideo)
	opKf4.Params().Set(ParamPrompt, "scene")
	opKf4.Params().Set(ParamInputReference, "https://example.com/img.png")
	opKf4.Params().Set(ParamImageTail, "https://example.com/tail.png")
	opKf4.Params().Set(ParamMode, ModePro)
	_, err = opKf4.Call(ctx, svc, &Options{})
	if !errors.Is(err, ErrKeyframeSecondsRequired) {
		t.Fatalf("expected ErrKeyframeSecondsRequired when kling-v2-1 keyframe without seconds, got %v", err)
	}

	// Keyframe kling-v2-1: mode=pro, seconds=10 → OK
	opKf5, _ := svc.Operation(xai.Model("kling-v2-1"), xai.GenVideo)
	opKf5.Params().Set(ParamPrompt, "scene")
	opKf5.Params().Set(ParamInputReference, "https://example.com/img.png")
	opKf5.Params().Set(ParamImageTail, "https://example.com/tail.png")
	opKf5.Params().Set(ParamMode, ModePro)
	opKf5.Params().Set(ParamSeconds, Seconds10)
	resp, err = opKf5.Call(ctx, svc, &Options{})
	if err != nil {
		t.Fatalf("keyframe with mode=pro and seconds=10 should succeed: %v", err)
	}
	if !resp.Done() {
		t.Fatal("expected sync response")
	}

	// Restriction: invalid seconds value
	opRestrict, _ := svc.Operation(xai.Model("kling-v2-1"), xai.GenVideo)
	opRestrict.Params().Set(ParamPrompt, "scene")
	opRestrict.Params().Set(ParamInputReference, "https://example.com/img.png")
	opRestrict.Params().Set(ParamSeconds, "15")
	opRestrict.Params().Set(ParamSize, Size1920x1080)
	_, err = opRestrict.Call(ctx, svc, &Options{})
	if err == nil {
		t.Fatal("expected error for invalid seconds=15")
	}
	if !errors.Is(err, xai.ErrValueNotAllowed) {
		t.Fatalf("expected ErrValueNotAllowed, got %v", err)
	}

	// Restriction: invalid size value
	opRestrict2, _ := svc.Operation(xai.Model("kling-v2-1"), xai.GenVideo)
	opRestrict2.Params().Set(ParamPrompt, "scene")
	opRestrict2.Params().Set(ParamInputReference, "https://example.com/img.png")
	opRestrict2.Params().Set(ParamSize, "4K")
	_, err = opRestrict2.Call(ctx, svc, &Options{})
	if err == nil {
		t.Fatal("expected error for invalid size=4K")
	}
	if !errors.Is(err, xai.ErrValueNotAllowed) {
		t.Fatalf("expected ErrValueNotAllowed, got %v", err)
	}

	// Restriction: valid constants pass
	opValid, _ := svc.Operation(xai.Model("kling-v2-1"), xai.GenVideo)
	opValid.Params().Set(ParamPrompt, "scene")
	opValid.Params().Set(ParamInputReference, "https://example.com/img.png")
	opValid.Params().Set(ParamSeconds, Seconds5)
	opValid.Params().Set(ParamSize, Size1920x1080)
	resp, err = opValid.Call(ctx, svc, &Options{})
	if err != nil {
		t.Fatalf("valid params with constants should succeed: %v", err)
	}
	if !resp.Done() {
		t.Fatal("expected sync response")
	}

	// Restriction (image): invalid resolution for kling-image-o1
	opImgRes, _ := svc.Operation(xai.Model("kling-image-o1"), xai.GenImage)
	opImgRes.Params().Set(ParamPrompt, "a cat")
	opImgRes.Params().Set(ParamResolution, "8K")
	_, err = opImgRes.Call(ctx, svc, &Options{})
	if err == nil {
		t.Fatal("expected error for invalid resolution=8K on kling-image-o1")
	}
	if !errors.Is(err, xai.ErrValueNotAllowed) {
		t.Fatalf("expected ErrValueNotAllowed for invalid resolution, got %v", err)
	}

	// Restriction (image): valid resolution passes
	opImgOK, _ := svc.Operation(xai.Model("kling-image-o1"), xai.GenImage)
	opImgOK.Params().Set(ParamPrompt, "a cat")
	opImgOK.Params().Set(ParamResolution, Resolution2K)
	resp, err = opImgOK.Call(ctx, svc, &Options{})
	if err != nil {
		t.Fatalf("valid resolution should succeed: %v", err)
	}
	if !resp.Done() {
		t.Fatal("expected sync response")
	}
}
