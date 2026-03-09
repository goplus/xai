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

package main

import (
	"context"
	"os"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/audio"
	"github.com/goplus/xai/spec/audio/provider/qiniu"
)

// DemoAudioURL is a public sample audio for ASR demos.
const DemoAudioURL = "https://static.qiniu.com/ai-inference/example-resources/example.mp3"

func newService() *audio.Service {
	token := os.Getenv("QINIU_API_KEY")
	if token != "" {
		return qiniu.NewService(token)
	}
	// Mock mode: no API key, returns placeholder results
	asrExec := &mockASRExecutor{}
	ttsExec := &mockTTSExecutor{}
	voiceLister := &mockVoiceLister{}
	return audio.NewService(asrExec, ttsExec, audio.WithVoiceLister(voiceLister))
}

type mockASRExecutor struct{}

func (m *mockASRExecutor) Transcribe(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	_ = model
	_ = params
	duration := 9.336
	return &audio.SyncOperationResponse{
		R: audio.NewOutputText("This is mock ASR result. Set QINIU_API_KEY for real transcription.", &duration),
	}, nil
}

type mockTTSExecutor struct{}

func (m *mockTTSExecutor) Synthesize(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	_ = model
	_ = params
	return &audio.SyncOperationResponse{
		R: audio.NewOutputAudio("https://aitoken-public.qnaigc.com/example/tts-sample.mp3", "mp3", "2.5"),
	}, nil
}

type mockVoiceLister struct{}

func (m *mockVoiceLister) ListVoices(ctx context.Context) ([]audio.VoiceListItem, error) {
	_ = ctx
	return []audio.VoiceListItem{
		{VoiceName: "Calm Executive (mock)", VoiceType: "qiniu_zh_female_wwxkjx", Category: "Traditional"},
		{VoiceName: "Mellow Male (mock)", VoiceType: "qiniu_zh_male_chunhou", Category: "Traditional"},
	}, nil
}
