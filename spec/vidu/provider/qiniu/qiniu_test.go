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

package qiniu

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goplus/xai/spec/vidu"
)

func TestBuildVideoRequestRouting(t *testing.T) {
	tests := []struct {
		name         string
		params       *vidu.VideoParams
		wantEndpoint string
	}{
		{
			name: "q1 text-to-video",
			params: &vidu.VideoParams{
				ModelName: vidu.ModelViduQ1,
				Prompt:    "a cat running",
			},
			wantEndpoint: EndpointQ1TextToVideo,
		},
		{
			name: "q1 reference-to-video",
			params: &vidu.VideoParams{
				ModelName:          vidu.ModelViduQ1,
				Prompt:             "a cat running",
				ReferenceImageURLs: []string{"https://example.com/1.png"},
			},
			wantEndpoint: EndpointQ1ReferenceToVideo,
		},
		{
			name: "q2 image-to-video",
			params: &vidu.VideoParams{
				ModelName: vidu.ModelViduQ2,
				Prompt:    "a woman walking",
				ImageURL:  "https://example.com/1.png",
			},
			wantEndpoint: EndpointQ2ImageToVideoPro,
		},
		{
			name: "q2 start-end-to-video",
			params: &vidu.VideoParams{
				ModelName:     vidu.ModelViduQ2,
				Prompt:        "dragon lands",
				StartImageURL: "https://example.com/start.png",
				EndImageURL:   "https://example.com/end.png",
			},
			wantEndpoint: EndpointQ2StartEndToVideoPro,
		},
		{
			name: "viduq2-turbo image-to-video",
			params: &vidu.VideoParams{
				ModelName: vidu.ModelViduQ2Turbo,
				Prompt:    "a woman walking",
				ImageURL:  "https://example.com/1.png",
			},
			wantEndpoint: EndpointQ2ImageToVideoTurbo,
		},
		{
			name: "viduq2-turbo start-end-to-video",
			params: &vidu.VideoParams{
				ModelName:     vidu.ModelViduQ2Turbo,
				Prompt:        "dragon lands",
				StartImageURL: "https://example.com/start.png",
				EndImageURL:   "https://example.com/end.png",
			},
			wantEndpoint: EndpointQ2StartEndToVideoTurbo,
		},
		{
			name: "viduq2-pro image-to-video",
			params: &vidu.VideoParams{
				ModelName: vidu.ModelViduQ2Pro,
				Prompt:    "a woman walking",
				ImageURL:  "https://example.com/1.png",
			},
			wantEndpoint: EndpointQ2ImageToVideoPro,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := BuildVideoRequest(tt.params.ModelName, tt.params)
			if err != nil {
				t.Fatalf("BuildVideoRequest error: %v", err)
			}
			if req.Endpoint != tt.wantEndpoint {
				t.Fatalf("endpoint = %s, want %s", req.Endpoint, tt.wantEndpoint)
			}
			if req.Body["prompt"] == "" {
				t.Fatal("expected prompt in request body")
			}
		})
	}
}

func TestVideoStatusResponse(t *testing.T) {
	resp := &VideoStatusResponse{
		Status: StatusCompleted,
		Result: &VideoResult{
			Video: &VideoItem{URL: "https://example.com/v1.mp4"},
		},
	}

	if !resp.IsCompleted() {
		t.Fatal("expected completed status")
	}
	if resp.IsFailed() {
		t.Fatal("did not expect failed status")
	}
	urls := resp.GetVideoURLs()
	if len(urls) != 1 || urls[0] != "https://example.com/v1.mp4" {
		t.Fatalf("unexpected urls: %+v", urls)
	}
}

func TestBuildVideoRequestReferenceSubjectsBody(t *testing.T) {
	params := &vidu.VideoParams{
		ModelName: vidu.ModelViduQ2,
		Prompt:    "a cat and dog",
		Subjects: []vidu.Subject{
			{ID: "cat", Images: []string{"https://example.com/cat.png"}},
			{ID: "dog", Images: []string{"https://example.com/dog.png"}, VoiceID: "voice-dog"},
		},
	}

	req, err := BuildVideoRequest("", params)
	if err != nil {
		t.Fatalf("BuildVideoRequest error: %v", err)
	}
	if req.Endpoint != EndpointQ2ReferenceToVideo {
		t.Fatalf("endpoint = %s, want %s", req.Endpoint, EndpointQ2ReferenceToVideo)
	}

	subjects, ok := req.Body["subjects"].([]map[string]any)
	if !ok {
		t.Fatalf("subjects type = %T, want []map[string]any", req.Body["subjects"])
	}
	if len(subjects) != 2 {
		t.Fatalf("subjects length = %d, want 2", len(subjects))
	}
	if _, exists := subjects[0]["voice_id"]; exists {
		t.Fatal("did not expect empty voice_id to be encoded")
	}
	if subjects[1]["voice_id"] != "voice-dog" {
		t.Fatalf("voice_id = %v, want voice-dog", subjects[1]["voice_id"])
	}
}

func TestBuildVideoRequestIncludesCommonOptionalFields(t *testing.T) {
	params := &vidu.VideoParams{
		ModelName:   vidu.ModelViduQ1,
		Prompt:      "anime cat under lanterns",
		AspectRatio: vidu.AspectRatio9_16,
		Style:       vidu.StyleAnime,
		BGM:         boolPtr(true),
		Watermark:   boolPtr(false),
	}

	req, err := BuildVideoRequest("", params)
	if err != nil {
		t.Fatalf("BuildVideoRequest error: %v", err)
	}

	if req.Body["aspect_ratio"] != vidu.AspectRatio9_16 {
		t.Fatalf("aspect_ratio = %v, want %s", req.Body["aspect_ratio"], vidu.AspectRatio9_16)
	}
	if req.Body["style"] != vidu.StyleAnime {
		t.Fatalf("style = %v, want %s", req.Body["style"], vidu.StyleAnime)
	}
	if req.Body["bgm"] != true {
		t.Fatalf("bgm = %v, want true", req.Body["bgm"])
	}
	if req.Body["watermark"] != false {
		t.Fatalf("watermark = %v, want false", req.Body["watermark"])
	}
}

func TestBuildVideoRequestReferenceSubjectsAudioBody(t *testing.T) {
	params := &vidu.VideoParams{
		ModelName: vidu.ModelViduQ1,
		Prompt:    "the @narrator walks by the @lantern",
		Audio:     boolPtr(true),
		Subjects: []vidu.Subject{
			{ID: "narrator", Images: []string{"https://example.com/narrator.png"}, VoiceID: "voice-narrator"},
			{ID: "lantern", Images: []string{"https://example.com/lantern.png"}},
		},
	}

	req, err := BuildVideoRequest("", params)
	if err != nil {
		t.Fatalf("BuildVideoRequest error: %v", err)
	}

	if req.Body["audio"] != true {
		t.Fatalf("audio = %v, want true", req.Body["audio"])
	}
}

func TestBuildVideoRequestImageAudioBody(t *testing.T) {
	params := &vidu.VideoParams{
		ModelName: vidu.ModelViduQ2Pro,
		Prompt:    "woman walking through a neon alley",
		ImageURL:  "https://example.com/ref.png",
		Audio:     boolPtr(true),
		VoiceID:   "voice-city",
		IsRec:     boolPtr(true),
	}

	req, err := BuildVideoRequest("", params)
	if err != nil {
		t.Fatalf("BuildVideoRequest error: %v", err)
	}

	if req.Endpoint != EndpointQ2ImageToVideoPro {
		t.Fatalf("endpoint = %s, want %s", req.Endpoint, EndpointQ2ImageToVideoPro)
	}
	if req.Body["audio"] != true {
		t.Fatalf("audio = %v, want true", req.Body["audio"])
	}
	if req.Body["voice_id"] != "voice-city" {
		t.Fatalf("voice_id = %v, want voice-city", req.Body["voice_id"])
	}
	if req.Body["is_rec"] != true {
		t.Fatalf("is_rec = %v, want true", req.Body["is_rec"])
	}
}

func TestSelectEndpointNormalizeModel(t *testing.T) {
	endpoint, err := SelectEndpoint("  VIDU-Q2  ", vidu.RouteImageToVideo)
	if err != nil {
		t.Fatalf("selectEndpoint error: %v", err)
	}
	if endpoint != EndpointQ2ImageToVideoPro {
		t.Fatalf("endpoint = %s, want %s", endpoint, EndpointQ2ImageToVideoPro)
	}
}

func TestVideoStatusResponseNormalizedStateAndMessage(t *testing.T) {
	resp := &VideoStatusResponse{
		Status: "  completed ",
		Result: &VideoResult{
			Video:  &VideoItem{URL: " https://example.com/v1.mp4 "},
			Videos: []VideoItem{{URL: " "}, {URL: "https://example.com/v2.mp4"}},
		},
		Error:   &VideoError{Message: "  failed in worker  "},
		Message: "  fallback message  ",
	}

	if !resp.IsCompleted() {
		t.Fatal("expected completed status after normalization")
	}

	resp.Status = "  failed "
	if !resp.IsFailed() {
		t.Fatal("expected failed status after normalization")
	}

	resp.Status = " running "
	if !resp.IsProcessing() {
		t.Fatal("expected processing status after normalization")
	}

	urls := resp.GetVideoURLs()
	if len(urls) != 2 {
		t.Fatalf("urls length = %d, want 2", len(urls))
	}
	if urls[0] != "https://example.com/v1.mp4" || urls[1] != "https://example.com/v2.mp4" {
		t.Fatalf("unexpected urls: %+v", urls)
	}

	if got := resp.GetErrorMessage(); got != "failed in worker" {
		t.Fatalf("GetErrorMessage() = %q, want %q", got, "failed in worker")
	}
	resp.Error = nil
	if got := resp.GetErrorMessage(); got != "fallback message" {
		t.Fatalf("GetErrorMessage() = %q, want %q", got, "fallback message")
	}
}

func TestBuildVideoRequestUnsupportedModelAndRoute(t *testing.T) {
	_, err := BuildVideoRequest("unknown-model", &vidu.VideoParams{
		ModelName: "unknown-model",
		Prompt:    "test",
	})
	if err == nil || !strings.Contains(err.Error(), `unsupported model "unknown-model"`) {
		t.Fatalf("expected unsupported model error, got %v", err)
	}

	_, err = BuildVideoRequest(vidu.ModelViduQ1, &vidu.VideoParams{
		ModelName: vidu.ModelViduQ1,
		Prompt:    "test",
		ImageURL:  "https://example.com/1.png",
	})
	if err == nil || !strings.Contains(err.Error(), "does not support route") {
		t.Fatalf("expected unsupported route error, got %v", err)
	}
}

func TestVideoExecutorSubmitAndPoll(t *testing.T) {
	var pollCount int

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodPost && r.URL.Path == EndpointQ2TextToVideo:
			_ = json.NewEncoder(w).Encode(VideoCreateResponse{
				Status:    StatusInQueue,
				RequestID: "req-123",
			})
			return
		case r.Method == http.MethodGet && r.URL.Path == GetVideoTaskStatusEndpoint("req-123"):
			pollCount++
			if pollCount == 1 {
				_ = json.NewEncoder(w).Encode(VideoStatusResponse{Status: StatusInQueue, RequestID: "req-123"})
				return
			}
			_ = json.NewEncoder(w).Encode(VideoStatusResponse{
				Status:    StatusCompleted,
				RequestID: "req-123",
				Result:    &VideoResult{Video: &VideoItem{URL: "https://example.com/final.mp4"}},
			})
			return
		default:
			http.NotFound(w, r)
			return
		}
	}))
	defer ts.Close()

	client := NewClient("test-apikey", WithBaseURL(ts.URL), WithDebugLog(false))
	backend := NewBackend(client)

	params := vidu.NewParams().
		Set(vidu.ParamPrompt, "a cat running")

	resp, err := backend.Submit(context.Background(), vidu.ModelViduQ2, params)
	if err != nil {
		t.Fatalf("Submit error: %v", err)
	}
	if resp.Done() {
		t.Fatal("expected async response from submit")
	}

	resp, err = resp.Retry(context.Background(), nil)
	if err != nil {
		t.Fatalf("Retry #1 error: %v", err)
	}
	if resp.Done() {
		t.Fatal("expected still processing after first poll")
	}

	resp, err = resp.Retry(context.Background(), nil)
	if err != nil {
		t.Fatalf("Retry #2 error: %v", err)
	}
	if !resp.Done() {
		t.Fatal("expected completed response after second poll")
	}
	if resp.Results() == nil || resp.Results().Len() != 1 {
		t.Fatalf("expected one output video, got %#v", resp.Results())
	}
}

func boolPtr(v bool) *bool { return &v }
