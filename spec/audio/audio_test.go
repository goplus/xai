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

package audio

import (
	"context"
	"errors"
	"testing"

	xai "github.com/goplus/xai/spec"
)

type mockASRExecutor struct {
	transcribe func(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error)
}

func (m *mockASRExecutor) Transcribe(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	if m.transcribe != nil {
		return m.transcribe(ctx, model, params)
	}
	return &SyncOperationResponse{R: NewOutputText("hello", nil)}, nil
}

type mockTTSExecutor struct {
	synthesize func(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error)
}

func (m *mockTTSExecutor) Synthesize(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	if m.synthesize != nil {
		return m.synthesize(ctx, model, params)
	}
	return &SyncOperationResponse{R: NewOutputAudio("https://example.com/audio.mp3", "mp3", "2.5")}, nil
}

func TestAudioService_Actions(t *testing.T) {
	svc := NewService(&mockASRExecutor{}, &mockTTSExecutor{})

	actions := svc.Actions(xai.Model("asr"))
	if len(actions) != 1 || actions[0] != xai.Transcribe {
		t.Fatalf("expected [Transcribe] for asr, got %v", actions)
	}

	actions = svc.Actions(xai.Model("tts-v1"))
	if len(actions) != 1 || actions[0] != xai.Synthesize {
		t.Fatalf("expected [Synthesize] for tts-v1, got %v", actions)
	}

	actions = svc.Actions(xai.Model("unknown"))
	if len(actions) != 0 {
		t.Fatalf("expected [] for unknown model, got %v", actions)
	}
}

func TestAudioService_Operation(t *testing.T) {
	svc := NewService(&mockASRExecutor{}, &mockTTSExecutor{})

	op, err := svc.Operation(xai.Model("asr"), xai.Transcribe)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if op == nil {
		t.Fatal("expected non-nil operation")
	}

	op, err = svc.Operation(xai.Model("tts-v1"), xai.Synthesize)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if op == nil {
		t.Fatal("expected non-nil operation")
	}

	_, err = svc.Operation(xai.Model("asr"), xai.Synthesize)
	if !errors.Is(err, xai.ErrNotFound) {
		t.Fatalf("expected ErrNotFound for asr+Synthesize, got %v", err)
	}
}

func TestTranscribe_Call(t *testing.T) {
	svc := NewService(&mockASRExecutor{}, &mockTTSExecutor{})
	ctx := context.Background()

	op, _ := svc.Operation(xai.Model("asr"), xai.Transcribe)
	op.Params().Set(ParamAudio, "https://example.com/audio.mp3")

	resp, err := op.Call(ctx, svc, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.Done() {
		t.Fatal("expected sync response to be done")
	}
	text := resp.Results().At(0).(*xai.OutputText)
	if text.Text != "hello" {
		t.Fatalf("expected text 'hello', got %q", text.Text)
	}
}

func TestTranscribe_MissingAudio(t *testing.T) {
	svc := NewService(&mockASRExecutor{}, &mockTTSExecutor{})
	ctx := context.Background()

	op, _ := svc.Operation(xai.Model("asr"), xai.Transcribe)
	// no ParamAudio set

	_, err := op.Call(ctx, svc, nil)
	if !errors.Is(err, ErrAudioRequired) {
		t.Fatalf("expected ErrAudioRequired, got %v", err)
	}
}

func TestSynthesize_Call(t *testing.T) {
	svc := NewService(&mockASRExecutor{}, &mockTTSExecutor{})
	ctx := context.Background()

	op, _ := svc.Operation(xai.Model("tts-v1"), xai.Synthesize)
	op.Params().Set(ParamInput, "你好")

	resp, err := op.Call(ctx, svc, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.Done() {
		t.Fatal("expected sync response to be done")
	}
	out := resp.Results().At(0).(*xai.OutputAudio)
	if out.Audio != "https://example.com/audio.mp3" || out.Format != "mp3" {
		t.Fatalf("unexpected output: audio=%q format=%q", out.Audio, out.Format)
	}
}

func TestSynthesize_MissingInput(t *testing.T) {
	svc := NewService(&mockASRExecutor{}, &mockTTSExecutor{})
	ctx := context.Background()

	op, _ := svc.Operation(xai.Model("tts-v1"), xai.Synthesize)
	// no ParamInput set

	_, err := op.Call(ctx, svc, nil)
	if !errors.Is(err, ErrInputRequired) {
		t.Fatalf("expected ErrInputRequired, got %v", err)
	}
}

func TestListVoices_NoVoiceLister(t *testing.T) {
	svc := NewService(&mockASRExecutor{}, &mockTTSExecutor{})
	ctx := context.Background()

	_, err := svc.ListVoices(ctx)
	if !errors.Is(err, xai.ErrNotSupported) {
		t.Fatalf("expected ErrNotSupported when no VoiceLister, got %v", err)
	}
}

func TestListVoices_WithVoiceLister(t *testing.T) {
	mockLister := &mockVoiceLister{
		voices: []VoiceListItem{
			{VoiceName: "沉稳高管", VoiceType: "qiniu_zh_female_wwxkjx"},
			{VoiceName: "醇厚男声", VoiceType: "qiniu_zh_male_chunhou"},
		},
	}
	svc := NewService(&mockASRExecutor{}, &mockTTSExecutor{}, WithVoiceLister(mockLister))
	ctx := context.Background()

	voices, err := svc.ListVoices(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(voices) != 2 {
		t.Fatalf("expected 2 voices, got %d", len(voices))
	}
	if voices[0].VoiceType != "qiniu_zh_female_wwxkjx" {
		t.Fatalf("expected first voice type qiniu_zh_female_wwxkjx, got %q", voices[0].VoiceType)
	}
}

type mockVoiceLister struct {
	voices []VoiceListItem
}

func (m *mockVoiceLister) ListVoices(ctx context.Context) ([]VoiceListItem, error) {
	return m.voices, nil
}
