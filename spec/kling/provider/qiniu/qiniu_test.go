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
	"sync/atomic"
	"testing"
	"time"

	"github.com/goplus/xai/spec/kling"
	"github.com/goplus/xai/spec/kling/image"
	"github.com/goplus/xai/spec/kling/video"
)

func TestNewClient(t *testing.T) {
	apiKey := "test-apikey"
	c := NewClient(apiKey)

	if c.ApiKey() != apiKey {
		t.Errorf("expected apiKey %q, got %q", apiKey, c.ApiKey())
	}
	if c.BaseURL() != DefaultBaseURL {
		t.Errorf("expected base URL %q, got %q", DefaultBaseURL, c.BaseURL())
	}
}

func TestClientSetApiKey(t *testing.T) {
	c := NewClient("initial-apikey")
	if c.ApiKey() != "initial-apikey" {
		t.Errorf("expected initial apiKey, got %q", c.ApiKey())
	}
	c.SetApiKey("new-apikey")
	if c.ApiKey() != "new-apikey" {
		t.Errorf("expected new apiKey after SetApiKey, got %q", c.ApiKey())
	}
}

func TestServiceSetApiKey(t *testing.T) {
	svc := NewService("initial-key")
	if svc.client.ApiKey() != "initial-key" {
		t.Errorf("expected initial key, got %q", svc.client.ApiKey())
	}
	svc.SetApiKey("new-key")
	if svc.client.ApiKey() != "new-key" {
		t.Errorf("expected new key after SetApiKey, got %q", svc.client.ApiKey())
	}
}

func TestNewClientWithOptions(t *testing.T) {
	apiKey := "test-apikey"
	customURL := "https://custom.api.com"

	c := NewClient(apiKey, WithBaseURL(customURL))

	if c.BaseURL() != customURL {
		t.Errorf("expected base URL %q, got %q", customURL, c.BaseURL())
	}
}

func TestNewClientWithRetryAndDebug(t *testing.T) {
	apiKey := "test-apikey"
	c := NewClient(apiKey,
		WithRetry(3, 500*time.Millisecond),
		WithDebugLog(true),
	)

	if c.maxRetries != 3 {
		t.Errorf("expected maxRetries 3, got %d", c.maxRetries)
	}
	if c.baseRetryDelay != 500*time.Millisecond {
		t.Errorf("expected baseRetryDelay 500ms, got %v", c.baseRetryDelay)
	}
	if !c.debugLog {
		t.Error("expected debugLog to be true")
	}
}

func TestBuildV1ImageRequest(t *testing.T) {
	n := 2
	params := &image.V1ImageParams{
		Prompt:         "a cute cat",
		N:              &n,
		AspectRatio:    "16:9",
		NegativePrompt: "blurry",
	}

	req, err := BuildImageRequest(kling.ModelKlingV1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Endpoint != EndpointImageGenerations {
		t.Errorf("expected endpoint %q, got %q", EndpointImageGenerations, req.Endpoint)
	}
	if req.IsO1 {
		t.Error("expected IsO1 to be false")
	}
	if req.Body["model"] != kling.ModelKlingV1 {
		t.Errorf("expected model %q, got %q", kling.ModelKlingV1, req.Body["model"])
	}
	if req.Body["prompt"] != "a cute cat" {
		t.Errorf("expected prompt %q, got %q", "a cute cat", req.Body["prompt"])
	}
	if req.Body["n"] != 2 {
		t.Errorf("expected n %d, got %v", 2, req.Body["n"])
	}
}

func TestBuildV2ImageRequestMultiImage(t *testing.T) {
	params := &image.V2ImageParams{
		ModelName: kling.ModelKlingV2,
		Prompt:    "combine images",
		SubjectImageList: []image.SubjectImageItem{
			{SubjectImage: "https://example.com/img1.jpg"},
			{SubjectImage: "https://example.com/img2.jpg"},
		},
		AspectRatio: "16:9",
	}

	req, err := BuildImageRequest(kling.ModelKlingV2, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Endpoint != EndpointImageEdits {
		t.Errorf("expected endpoint %q, got %q", EndpointImageEdits, req.Endpoint)
	}
	if req.Body["image"] != "" {
		t.Errorf("expected empty image for multi-image mode, got %q", req.Body["image"])
	}
	subjectList, ok := req.Body["subject_image_list"].([]map[string]string)
	if !ok {
		t.Fatal("expected subject_image_list to be []map[string]string")
	}
	if len(subjectList) != 2 {
		t.Errorf("expected 2 subject images, got %d", len(subjectList))
	}
}

func TestBuildV2ImageRequestReferenceImages(t *testing.T) {
	params := &image.V2ImageParams{
		ModelName:       kling.ModelKlingV2,
		Prompt:          "stylize this image",
		Image:           "https://example.com/base.jpg",
		ReferenceImages: []string{"https://example.com/style.jpg"},
		AspectRatio:     "16:9",
	}

	req, err := BuildImageRequest(kling.ModelKlingV2, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	refs, ok := req.Body["reference_images"].([]string)
	if !ok {
		t.Fatalf("expected reference_images to be []string, got %T", req.Body["reference_images"])
	}
	if len(refs) != 1 || refs[0] != "https://example.com/style.jpg" {
		t.Fatalf("unexpected reference_images payload: %#v", refs)
	}
}

func TestBuildV21ImageRequestReferenceImages(t *testing.T) {
	params := &image.V21ImageParams{
		Prompt:          "same style but a mountain landscape",
		Image:           "https://example.com/base.jpg",
		ReferenceImages: []string{"https://example.com/style.jpg"},
		AspectRatio:     "16:9",
	}

	req, err := BuildImageRequest(kling.ModelKlingV21, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	refs, ok := req.Body["reference_images"].([]string)
	if !ok {
		t.Fatalf("expected reference_images to be []string, got %T", req.Body["reference_images"])
	}
	if len(refs) != 1 || refs[0] != "https://example.com/style.jpg" {
		t.Fatalf("unexpected reference_images payload: %#v", refs)
	}
}

func TestBuildO1ImageRequest(t *testing.T) {
	params := &image.O1ImageParams{
		Prompt:          "a beautiful sunset",
		N:               2,
		Resolution:      "2K",
		AspectRatio:     "16:9",
		ReferenceImages: []string{"https://example.com/ref.jpg"},
	}

	req, err := BuildImageRequest(kling.ModelKlingImageO1, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Endpoint != EndpointImageO1 {
		t.Errorf("expected endpoint %q, got %q", EndpointImageO1, req.Endpoint)
	}
	if !req.IsO1 {
		t.Error("expected IsO1 to be true")
	}
	if req.Body["num_images"] != 2 {
		t.Errorf("expected num_images %d, got %v", 2, req.Body["num_images"])
	}
	imageURLs, ok := req.Body["image_urls"].([]string)
	if !ok {
		t.Fatal("expected image_urls to be []string")
	}
	if len(imageURLs) != 1 {
		t.Errorf("expected 1 image URL, got %d", len(imageURLs))
	}
}

func TestBuildV21VideoRequest(t *testing.T) {
	params := &video.V21VideoParams{
		ModelName:      kling.ModelKlingV21Video,
		Prompt:         "a person running",
		InputReference: "https://example.com/img.jpg",
		Mode:           "pro",
		Seconds:        "5",
		Size:           "1920x1080",
	}

	req, err := BuildVideoRequest(kling.ModelKlingV21Video, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Body["model"] != kling.ModelKlingV21Video {
		t.Errorf("expected model %q, got %q", kling.ModelKlingV21Video, req.Body["model"])
	}
	if req.Body["input_reference"] != "https://example.com/img.jpg" {
		t.Errorf("expected input_reference, got %v", req.Body["input_reference"])
	}
}

func TestBuildO1VideoRequestLegacyFrames(t *testing.T) {
	params := kling.NewParams()
	params.
		Set(kling.ParamPrompt, "a runner crossing the finish line").
		Set(kling.ParamInputReference, "https://example.com/first.jpg").
		Set(kling.ParamImageTail, "https://example.com/last.jpg").
		Set(kling.ParamMode, kling.ModePro)

	typedParams, err := kling.BuildVideoParams(kling.ModelKlingVideoO1, params)
	if err != nil {
		t.Fatalf("unexpected build params error: %v", err)
	}

	req, err := BuildVideoRequest(kling.ModelKlingVideoO1, typedParams)
	if err != nil {
		t.Fatalf("unexpected build request error: %v", err)
	}

	imageList, ok := req.Body["image_list"].([]map[string]any)
	if !ok {
		t.Fatalf("expected image_list payload, got %T", req.Body["image_list"])
	}
	if len(imageList) != 2 {
		t.Fatalf("expected 2 legacy-converted image refs, got %d", len(imageList))
	}
	if imageList[0]["image"] != "https://example.com/first.jpg" || imageList[0]["type"].(video.ImageRefType) != video.ImageTypeFirstFrame {
		t.Fatalf("unexpected first frame payload: %#v", imageList[0])
	}
	if imageList[1]["image"] != "https://example.com/last.jpg" || imageList[1]["type"].(video.ImageRefType) != video.ImageTypeEndFrame {
		t.Fatalf("unexpected end frame payload: %#v", imageList[1])
	}
}

func TestBuildV3VideoRequestWithMultiPrompt(t *testing.T) {
	params := &video.V3VideoParams{
		ModelName: kling.ModelKlingV3,
		Prompt:    "a city short film",
		MultiShot: true,
		ShotType:  "customize",
		MultiPrompt: []video.MultiPromptItem{
			{Index: 1, Prompt: "scene 1", Duration: "3"},
			{Index: 2, Prompt: "scene 2", Duration: "2"},
		},
		Mode:    "pro",
		Seconds: "5",
	}

	req, err := BuildVideoRequest(kling.ModelKlingV3, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Body["multi_shot"] != true {
		t.Error("expected multi_shot to be true")
	}
	if req.Body["shot_type"] != "customize" {
		t.Errorf("expected shot_type customize, got %v", req.Body["shot_type"])
	}
	multiPrompt, ok := req.Body["multi_prompt"].([]map[string]any)
	if !ok {
		t.Fatal("expected multi_prompt to be []map[string]any")
	}
	if len(multiPrompt) != 2 {
		t.Errorf("expected 2 multi_prompt items, got %d", len(multiPrompt))
	}
}

func TestBuildV3OmniVideoRequest(t *testing.T) {
	params := &video.V3OmniVideoParams{
		ModelName: kling.ModelKlingV3Omni,
		Prompt:    "a cat running",
		MultiShot: true,
		MultiPrompt: []video.MultiPromptItem{
			{Index: 1, Prompt: "scene 1", Duration: "5"},
			{Index: 2, Prompt: "scene 2", Duration: "5"},
		},
		ImageList: []video.ImageInput{
			{Image: "https://example.com/img.jpg", Type: "first_frame"},
		},
		Sound:   "on",
		Mode:    "pro",
		Seconds: "10",
	}

	req, err := BuildVideoRequest(kling.ModelKlingV3Omni, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if req.Body["multi_shot"] != true {
		t.Error("expected multi_shot to be true")
	}
	multiPrompt, ok := req.Body["multi_prompt"].([]map[string]any)
	if !ok {
		t.Fatal("expected multi_prompt to be []map[string]any")
	}
	if len(multiPrompt) != 2 {
		t.Errorf("expected 2 multi_prompt items, got %d", len(multiPrompt))
	}
}

func TestImageTaskStatusResponse(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		wantDone    bool
		wantFailed  bool
		wantProcess bool
	}{
		{"completed", StatusCompleted, true, false, false},
		{"succeeded", StatusSucceeded, true, false, false},
		{"succeed", StatusSucceed, true, false, false},
		{"failed", StatusFailed, false, true, false},
		{"processing", StatusProcessing, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &ImageTaskStatusResponse{Status: tt.status}
			if resp.IsCompleted() != tt.wantDone {
				t.Errorf("IsCompleted() = %v, want %v", resp.IsCompleted(), tt.wantDone)
			}
			if resp.IsFailed() != tt.wantFailed {
				t.Errorf("IsFailed() = %v, want %v", resp.IsFailed(), tt.wantFailed)
			}
			if resp.IsProcessing() != tt.wantProcess {
				t.Errorf("IsProcessing() = %v, want %v", resp.IsProcessing(), tt.wantProcess)
			}
		})
	}
}

func TestVideoTaskStatusResponse(t *testing.T) {
	tests := []struct {
		name        string
		status      string
		wantDone    bool
		wantFailed  bool
		wantProcess bool
	}{
		{"completed", StatusCompleted, true, false, false},
		{"failed", StatusFailed, false, true, false},
		{"initializing", StatusInitializing, false, false, true},
		{"queued", StatusQueued, false, false, true},
		{"in_progress", StatusInProgress, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &VideoTaskResponse{Status: tt.status}
			if resp.IsCompleted() != tt.wantDone {
				t.Errorf("IsCompleted() = %v, want %v", resp.IsCompleted(), tt.wantDone)
			}
			if resp.IsFailed() != tt.wantFailed {
				t.Errorf("IsFailed() = %v, want %v", resp.IsFailed(), tt.wantFailed)
			}
			if resp.IsProcessing() != tt.wantProcess {
				t.Errorf("IsProcessing() = %v, want %v", resp.IsProcessing(), tt.wantProcess)
			}
		})
	}
}

func TestGetImageURLs(t *testing.T) {
	resp := &ImageTaskStatusResponse{
		Status: StatusCompleted,
		Data: []ImageData{
			{URL: "https://example.com/img1.png"},
			{URL: "https://example.com/img2.png"},
		},
	}

	urls := resp.GetImageURLs()
	if len(urls) != 2 {
		t.Errorf("expected 2 URLs, got %d", len(urls))
	}
	if urls[0] != "https://example.com/img1.png" {
		t.Errorf("unexpected URL: %s", urls[0])
	}
}

func TestGetVideoURLs(t *testing.T) {
	resp := &VideoTaskResponse{
		Status: StatusCompleted,
		TaskResult: &VideoTaskResult{
			Videos: []VideoItem{
				{URL: "https://example.com/video1.mp4"},
				{URL: "https://example.com/video2.mp4"},
			},
		},
	}

	urls := resp.GetVideoURLs()
	if len(urls) != 2 {
		t.Errorf("expected 2 URLs, got %d", len(urls))
	}
	if urls[0] != "https://example.com/video1.mp4" {
		t.Errorf("unexpected URL: %s", urls[0])
	}
}

func TestAPIError(t *testing.T) {
	tests := []struct {
		name     string
		err      APIError
		wantErr  bool
		contains string
	}{
		{
			name:     "with message",
			err:      APIError{Message: "invalid request"},
			wantErr:  true,
			contains: "invalid request",
		},
		{
			name: "with error struct",
			err: APIError{
				Error_: &struct {
					Code    string `json:"code,omitempty"`
					Message string `json:"message,omitempty"`
					Type    string `json:"type,omitempty"`
				}{
					Code:    "invalid_param",
					Message: "prompt is required",
				},
			},
			wantErr:  true,
			contains: "prompt is required",
		},
		{
			name:    "no error",
			err:     APIError{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.HasError() != tt.wantErr {
				t.Errorf("HasError() = %v, want %v", tt.err.HasError(), tt.wantErr)
			}
			if tt.wantErr {
				errStr := tt.err.Error()
				if tt.contains != "" && !contains(errStr, tt.contains) {
					t.Errorf("Error() = %q, want to contain %q", errStr, tt.contains)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && searchSubstring(s, substr)))
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestIsO1TaskID(t *testing.T) {
	tests := []struct {
		taskID string
		want   bool
	}{
		{"qimage-root-1770199726278452760", true},
		{"image-1762159125266058362-1383010xxx", false},
		{"qvideo-user123-1766391125174150336", false},
	}

	for _, tt := range tests {
		t.Run(tt.taskID, func(t *testing.T) {
			if got := isO1TaskID(tt.taskID); got != tt.want {
				t.Errorf("isO1TaskID(%q) = %v, want %v", tt.taskID, got, tt.want)
			}
		})
	}
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer test-apikey" {
			t.Errorf("unexpected Authorization header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected Content-Type: %s", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"task_id": "test-task-123"})
	}))
	defer server.Close()

	client := NewClient("test-apikey", WithBaseURL(server.URL))
	ctx := context.Background()
	resp, err := client.Post(ctx, "/v1/images/generations", map[string]string{"prompt": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["task_id"] != "test-task-123" {
		t.Errorf("unexpected task_id: %s", result["task_id"])
	}
}

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "completed"})
	}))
	defer server.Close()

	client := NewClient("test-apikey", WithBaseURL(server.URL))
	ctx := context.Background()
	resp, err := client.Get(ctx, "/v1/images/tasks/test-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["status"] != "completed" {
		t.Errorf("unexpected status: %s", result["status"])
	}
}

func TestGetImageTaskStatusEndpoint(t *testing.T) {
	tests := []struct {
		taskID   string
		isO1     bool
		expected string
	}{
		{"image-123", false, "/v1/images/tasks/image-123"},
		{"qimage-root-123", true, "/queue/fal-ai/kling-image/requests/qimage-root-123/status"},
	}

	for _, tt := range tests {
		t.Run(tt.taskID, func(t *testing.T) {
			got := GetImageTaskStatusEndpoint(tt.taskID, tt.isO1)
			if got != tt.expected {
				t.Errorf("GetImageTaskStatusEndpoint(%q, %v) = %q, want %q", tt.taskID, tt.isO1, got, tt.expected)
			}
		})
	}
}

func TestGetVideoTaskStatusEndpoint(t *testing.T) {
	taskID := "qvideo-user123-1766391125174150336"
	expected := "/v1/videos/qvideo-user123-1766391125174150336"
	got := GetVideoTaskStatusEndpoint(taskID)
	if got != expected {
		t.Errorf("GetVideoTaskStatusEndpoint(%q) = %q, want %q", taskID, got, expected)
	}
}

func TestClientRetry(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attempts, 1)
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{"error": "service unavailable"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewClient("test-apikey",
		WithBaseURL(server.URL),
		WithRetry(3, 10*time.Millisecond),
	)

	ctx := context.Background()
	resp, err := client.Get(ctx, "/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(resp, &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("unexpected status: %s", result["status"])
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClientRetryExhausted(t *testing.T) {
	var attempts int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Header().Set("X-Request-Id", "req-123")
		json.NewEncoder(w).Encode(map[string]string{"code": "SERVICE_UNAVAILABLE", "message": "service unavailable"})
	}))
	defer server.Close()

	client := NewClient("test-apikey",
		WithBaseURL(server.URL),
		WithRetry(2, 10*time.Millisecond),
	)

	ctx := context.Background()
	_, err := client.Get(ctx, "/test")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusServiceUnavailable {
		t.Errorf("expected status code %d, got %d", http.StatusServiceUnavailable, apiErr.StatusCode)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts (1 initial + 2 retries), got %d", attempts)
	}
}

func TestAPIErrorEnhanced(t *testing.T) {
	tests := []struct {
		name     string
		err      APIError
		contains []string
	}{
		{
			name: "with code and request_id",
			err: APIError{
				Code:       "INVALID_PARAM",
				Message:    "prompt is required",
				RequestID:  "req-abc-123",
				StatusCode: 400,
			},
			contains: []string{"INVALID_PARAM", "prompt is required", "req-abc-123", "400"},
		},
		{
			name: "with nested error",
			err: APIError{
				Error_: &struct {
					Code    string `json:"code,omitempty"`
					Message string `json:"message,omitempty"`
					Type    string `json:"type,omitempty"`
				}{
					Code:    "rate_limit_exceeded",
					Message: "too many requests",
				},
				RequestID:  "req-xyz-789",
				StatusCode: 429,
			},
			contains: []string{"rate_limit_exceeded", "too many requests", "req-xyz-789", "429"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, substr := range tt.contains {
				if !contains(errStr, substr) {
					t.Errorf("Error() = %q, want to contain %q", errStr, substr)
				}
			}
		})
	}
}

func TestAPIErrorGetters(t *testing.T) {
	err := &APIError{
		Code:    "outer_code",
		Message: "outer_message",
		Error_: &struct {
			Code    string `json:"code,omitempty"`
			Message string `json:"message,omitempty"`
			Type    string `json:"type,omitempty"`
		}{
			Code:    "inner_code",
			Message: "inner_message",
		},
	}

	if err.GetCode() != "inner_code" {
		t.Errorf("GetCode() = %q, want %q", err.GetCode(), "inner_code")
	}
	if err.GetMessage() != "inner_message" {
		t.Errorf("GetMessage() = %q, want %q", err.GetMessage(), "inner_message")
	}

	errNoInner := &APIError{
		Code:    "outer_code",
		Message: "outer_message",
	}
	if errNoInner.GetCode() != "outer_code" {
		t.Errorf("GetCode() = %q, want %q", errNoInner.GetCode(), "outer_code")
	}
	if errNoInner.GetMessage() != "outer_message" {
		t.Errorf("GetMessage() = %q, want %q", errNoInner.GetMessage(), "outer_message")
	}
}

func TestIsRetryableStatus(t *testing.T) {
	retryable := []int{429, 500, 502, 503, 504}
	nonRetryable := []int{200, 201, 400, 401, 403, 404, 422}

	for _, code := range retryable {
		if !isRetryableStatus(code) {
			t.Errorf("expected %d to be retryable", code)
		}
	}

	for _, code := range nonRetryable {
		if isRetryableStatus(code) {
			t.Errorf("expected %d to not be retryable", code)
		}
	}
}

func TestBuildCurlCommand(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   string
		method   string
		url      string
		body     []byte
		contains []string
	}{
		{
			name:   "POST with body",
			apiKey: "sk-test-apikey",
			method: "POST",
			url:    "https://api.qnaigc.com/v1/images/generations",
			body:   []byte(`{"model":"kling-v1","prompt":"a cute cat"}`),
			contains: []string{
				"curl -X POST",
				"-H 'Authorization: Bearer sk-test-apikey'",
				"-H 'Content-Type: application/json'",
				`-d '{"model":"kling-v1","prompt":"a cute cat"}'`,
				"'https://api.qnaigc.com/v1/images/generations'",
			},
		},
		{
			name:   "GET without body",
			apiKey: "sk-test-apikey",
			method: "GET",
			url:    "https://api.qnaigc.com/v1/images/tasks/task-123",
			body:   nil,
			contains: []string{
				"curl -X GET",
				"-H 'Authorization: Bearer sk-test-apikey'",
				"-H 'Content-Type: application/json'",
				"'https://api.qnaigc.com/v1/images/tasks/task-123'",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(tt.apiKey)
			curlCmd := client.buildCurlCommand(tt.method, tt.url, tt.body)

			for _, substr := range tt.contains {
				if !contains(curlCmd, substr) {
					t.Errorf("buildCurlCommand() = %q, want to contain %q", curlCmd, substr)
				}
			}
		})
	}
}

func TestBuildCurlCommandNoBody(t *testing.T) {
	client := NewClient("test-apikey")
	curlCmd := client.buildCurlCommand("GET", "https://api.example.com/test", nil)

	if contains(curlCmd, "-d") {
		t.Errorf("buildCurlCommand() should not contain -d for empty body, got: %s", curlCmd)
	}
}
