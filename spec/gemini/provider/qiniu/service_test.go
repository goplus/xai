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
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	xai "github.com/goplus/xai/spec"
	"google.golang.org/genai"
)

func TestParseURIQuery(t *testing.T) {
	values, err := parseURIQuery(Scheme + ":base=https://api.qnaigc.com/v1/&key=token-1")
	if err != nil {
		t.Fatalf("parseURIQuery failed: %v", err)
	}
	if got := values.Get("base"); got != "https://api.qnaigc.com/v1/" {
		t.Fatalf("unexpected base: %q", got)
	}
	if got := values.Get("key"); got != "token-1" {
		t.Fatalf("unexpected key: %q", got)
	}
}

func TestNewService(t *testing.T) {
	svc := NewService("token-1", WithBaseURL("https://openai.sufy.com/v1"))
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if got := svc.Features(); got&(xai.FeatureGen|xai.FeatureGenStream|xai.FeatureOperation) == 0 {
		t.Fatalf("unexpected features: %v", got)
	}
}

func TestRegister(t *testing.T) {
	Register("token-1")
	svc, err := xai.New(context.Background(), Scheme+":key=token-2")
	if err != nil {
		t.Fatalf("xai.New failed: %v", err)
	}
	if svc == nil {
		t.Fatal("xai.New returned nil service")
	}
}

func TestOperationGenImage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/images/generations") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-1" {
			t.Fatalf("unexpected auth header: %s", got)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["model"] != "gemini-2.5-flash-image" {
			t.Fatalf("unexpected model: %#v", body["model"])
		}
		if body["prompt"] != "draw a cat" {
			t.Fatalf("unexpected prompt: %#v", body["prompt"])
		}
		imageConfig, _ := body["image_config"].(map[string]any)
		if imageConfig["aspect_ratio"] != "16:9" {
			t.Fatalf("unexpected image_config.aspect_ratio: %#v", imageConfig["aspect_ratio"])
		}
		if imageConfig["image_size"] != "1K" {
			t.Fatalf("unexpected image_config.image_size: %#v", imageConfig["image_size"])
		}
		_, _ = w.Write([]byte(`{
			"created": 1,
			"output_format": "png",
			"data": [{"b64_json":"aGVsbG8="}],
			"usage": {"total_tokens": 42}
		}`))
	}))
	defer ts.Close()

	svc := NewService("token-1", WithBaseURL(ts.URL+"/v1/"), WithHTTPClient(ts.Client()))
	op, err := svc.Operation("gemini-2.5-flash-image", xai.GenImage)
	if err != nil {
		t.Fatalf("Operation failed: %v", err)
	}
	op.Params().
		Set("Prompt", "draw a cat").
		Set("AspectRatio", "16:9").
		Set("ImageSize", "1K")
	resp, err := op.Call(context.Background(), svc, nil)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if !resp.Done() {
		t.Fatal("expected sync response")
	}
	ret := resp.Results()
	if ret.Len() != 1 {
		t.Fatalf("unexpected results len: %d", ret.Len())
	}
	imgOut := ret.At(0).(*xai.OutputImage)
	if imgOut.Image == nil {
		t.Fatal("expected non-nil image")
	}
	img := imgOut.Image.StgUri()
	if !strings.HasPrefix(img, "data:image/png;base64,") {
		t.Fatalf("unexpected image uri: %s", img)
	}
}

func TestOperationEditImage(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/images/edits") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["prompt"] != "watercolor style" {
			t.Fatalf("unexpected prompt: %#v", body["prompt"])
		}
		imageConfig, _ := body["image_config"].(map[string]any)
		if imageConfig["aspect_ratio"] != "16:9" {
			t.Fatalf("unexpected image_config.aspect_ratio: %#v", imageConfig["aspect_ratio"])
		}
		if imageConfig["image_size"] != "1K" {
			t.Fatalf("unexpected image_config.image_size: %#v", imageConfig["image_size"])
		}
		_, _ = w.Write([]byte(`{"created": 2, "data":[{"url":"https://example.com/edited.png"}]}`))
	}))
	defer ts.Close()

	svc := NewService("token-1", WithBaseURL(ts.URL+"/v1/"), WithHTTPClient(ts.Client()))
	op, err := svc.Operation("gemini-3.0-pro-image-preview", xai.EditImage)
	if err != nil {
		t.Fatalf("Operation failed: %v", err)
	}
	ref, _ := svc.ReferenceImage(svc.ImageFromStgUri(xai.ImageJPEG, "https://example.com/src.png"), 0, xai.RawReferenceImage)
	op.Params().
		Set("Prompt", "watercolor style").
		Set("References", []genai.ReferenceImage{ref.(genai.ReferenceImage)}).
		Set("AspectRatio", "16:9").
		Set("ImageSize", "1K")
	resp, err := op.Call(context.Background(), svc, nil)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if resp.Results().Len() != 1 {
		t.Fatalf("unexpected results len: %d", resp.Results().Len())
	}
}

func TestActions(t *testing.T) {
	svc := NewService("token-1")

	videoActions := svc.Actions("veo-3.0-generate-preview")
	if len(videoActions) != 1 || videoActions[0] != xai.GenVideo {
		t.Fatalf("unexpected veo actions: %v", videoActions)
	}

	imageActions := svc.Actions("gemini-2.5-flash-image")
	if len(imageActions) != 2 || imageActions[0] != xai.GenImage || imageActions[1] != xai.EditImage {
		t.Fatalf("unexpected gemini actions: %v", imageActions)
	}
}

func TestOperationGenVideo(t *testing.T) {
	const (
		taskID      = "chatvideo-1709712000000000000-uid123"
		videoURL    = "https://cdn.example.com/videos/sample_0.mp4?token=xxx&e=1711904000"
		callbackURL = "https://your-server.com/api/veo-callback"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/videos/generations"):
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)

			if body["model"] != "veo-3.0-generate-preview" {
				t.Fatalf("unexpected model: %#v", body["model"])
			}
			if body["callback_url"] != callbackURL {
				t.Fatalf("unexpected callback_url: %#v", body["callback_url"])
			}

			instances, ok := body["instances"].([]any)
			if !ok || len(instances) != 1 {
				t.Fatalf("unexpected instances: %#v", body["instances"])
			}
			inst0, ok := instances[0].(map[string]any)
			if !ok || inst0["prompt"] != "A cat playing with a ball of yarn" {
				t.Fatalf("unexpected prompt instance: %#v", instances[0])
			}

			params, ok := body["parameters"].(map[string]any)
			if !ok {
				t.Fatalf("missing parameters: %#v", body["parameters"])
			}
			if got := params["aspectRatio"]; got != "16:9" {
				t.Fatalf("unexpected aspectRatio: %#v", got)
			}
			if got := params["durationSeconds"]; got != float64(8) {
				t.Fatalf("unexpected durationSeconds: %#v", got)
			}
			if got := params["sampleCount"]; got != float64(1) {
				t.Fatalf("unexpected sampleCount: %#v", got)
			}
			if got := params["seed"]; got != float64(0) {
				t.Fatalf("unexpected seed: %#v", got)
			}
			if got := params["negativePrompt"]; got != "blurry, low quality" {
				t.Fatalf("unexpected negativePrompt: %#v", got)
			}
			if got := params["personGeneration"]; got != "dont_allow" {
				t.Fatalf("unexpected personGeneration: %#v", got)
			}
			if got := params["generateAudio"]; got != true {
				t.Fatalf("unexpected generateAudio: %#v", got)
			}

			_, _ = w.Write([]byte(`{"id":"` + taskID + `"}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/videos/generations/"+taskID):
			_, _ = w.Write([]byte(`{
				"id":"` + taskID + `",
				"model":"veo-3.0-generate-preview",
				"status":"Completed",
				"message":"完成",
				"data":{
					"videos":[{"url":"` + videoURL + `","mimeType":"video/mp4"}],
					"raiMediaFilteredCount":0
				}
			}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	svc := NewService("token-1", WithBaseURL(ts.URL+"/v1/"), WithHTTPClient(ts.Client()))
	op, err := svc.Operation("veo-3.0-generate-preview", xai.GenVideo)
	if err != nil {
		t.Fatalf("Operation failed: %v", err)
	}
	op.Params().
		Set("Prompt", "A cat playing with a ball of yarn").
		Set("AspectRatio", "16:9 (landscape)").
		Set("DurationSeconds", int32(8)).
		Set("NumberOfVideos", int32(1)).
		Set("Seed", int32(0)).
		Set("NegativePrompt", "blurry, low quality").
		Set("PersonGeneration", "dont_allow").
		Set("GenerateAudio", true).
		Set("PubsubTopic", callbackURL)

	resp, err := op.Call(context.Background(), svc, nil)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if resp.Done() {
		t.Fatal("expected async response")
	}
	if got := resp.TaskID(); got != taskID {
		t.Fatalf("unexpected task id: %q", got)
	}

	resp, err = resp.Retry(context.Background(), svc)
	if err != nil {
		t.Fatalf("Retry failed: %v", err)
	}
	if !resp.Done() {
		t.Fatal("expected completed response")
	}
	if got := resp.TaskID(); got != taskID {
		t.Fatalf("unexpected task id after retry: %q", got)
	}

	ret := resp.Results()
	if ret.Len() != 1 {
		t.Fatalf("unexpected results len: %d", ret.Len())
	}
	out := ret.At(0).(*xai.OutputVideo)
	if out.Video == nil {
		t.Fatal("expected non-nil output video")
	}
	if got := out.Video.StgUri(); got != videoURL {
		t.Fatalf("unexpected video url: %s", got)
	}
	if got := out.Video.Type(); got != xai.VideoMP4 {
		t.Fatalf("unexpected video mime: %s", got)
	}
}

func TestOperationGenVideoImageToVideo(t *testing.T) {
	const (
		taskID   = "chatvideo-img2vid-123"
		videoURL = "https://cdn.example.com/videos/img2vid_0.mp4"
	)
	imgBytes := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a} // minimal PNG header

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/videos/generations") {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)

		instances, ok := body["instances"].([]any)
		if !ok || len(instances) != 1 {
			t.Fatalf("unexpected instances: %#v", body["instances"])
		}
		inst0, ok := instances[0].(map[string]any)
		if !ok {
			t.Fatalf("unexpected instance type: %T", instances[0])
		}
		if inst0["prompt"] != "A gentle breeze blowing through the scene" {
			t.Fatalf("unexpected prompt: %#v", inst0["prompt"])
		}
		imgObj, ok := inst0["image"].(map[string]any)
		if !ok {
			t.Fatalf("missing image in instance: %#v", inst0)
		}
		if imgObj["bytesBase64Encoded"] == nil && imgObj["uri"] == nil {
			t.Fatalf("image must have bytesBase64Encoded or uri: %#v", imgObj)
		}
		if imgObj["mimeType"] != "image/png" {
			t.Fatalf("unexpected mimeType: %#v", imgObj["mimeType"])
		}

		_, _ = w.Write([]byte(`{"id":"` + taskID + `"}`))
	}))
	defer ts.Close()

	svc := NewService("token-1", WithBaseURL(ts.URL+"/v1/"), WithHTTPClient(ts.Client()))
	img := svc.ImageFromBytes(xai.ImagePNG, imgBytes)

	op, err := svc.Operation("veo-3.0-generate-preview", xai.GenVideo)
	if err != nil {
		t.Fatalf("Operation failed: %v", err)
	}
	op.Params().
		Set("Image", img).
		Set("Prompt", "A gentle breeze blowing through the scene").
		Set("AspectRatio", "16:9 (landscape)").
		Set("DurationSeconds", int32(6)).
		Set("NumberOfVideos", int32(1)).
		Set("Seed", int32(100)).
		Set("PersonGeneration", "dont_allow")

	resp, err := op.Call(context.Background(), svc, nil)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if resp.TaskID() != taskID {
		t.Fatalf("unexpected task id: %q", resp.TaskID())
	}
	_ = videoURL // used in response mock if we add Retry
}

func TestOperationGenVideoFirstLastFrame(t *testing.T) {
	const taskID = "chatvideo-fl-123"
	firstBytes := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	lastBytes := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/videos/generations") {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)

		instances, ok := body["instances"].([]any)
		if !ok || len(instances) != 1 {
			t.Fatalf("unexpected instances: %#v", body["instances"])
		}
		inst0, ok := instances[0].(map[string]any)
		if !ok {
			t.Fatalf("unexpected instance type: %T", instances[0])
		}
		if inst0["prompt"] != "Smooth transition" {
			t.Fatalf("unexpected prompt: %#v", inst0["prompt"])
		}
		if _, ok := inst0["image"].(map[string]any); !ok {
			t.Fatalf("missing image (first frame): %#v", inst0)
		}
		if _, ok := inst0["lastFrame"].(map[string]any); !ok {
			t.Fatalf("missing lastFrame: %#v", inst0)
		}

		_, _ = w.Write([]byte(`{"id":"` + taskID + `"}`))
	}))
	defer ts.Close()

	svc := NewService("token-1", WithBaseURL(ts.URL+"/v1/"), WithHTTPClient(ts.Client()))
	firstImg := svc.ImageFromBytes(xai.ImagePNG, firstBytes)
	lastImg := svc.ImageFromBytes(xai.ImagePNG, lastBytes)

	op, err := svc.Operation("veo-3.0-generate-preview", xai.GenVideo)
	if err != nil {
		t.Fatalf("Operation failed: %v", err)
	}
	op.Params().
		Set("Image", firstImg).
		Set("LastFrame", lastImg).
		Set("Prompt", "Smooth transition").
		Set("AspectRatio", "16:9 (landscape)").
		Set("DurationSeconds", int32(6)).
		Set("PersonGeneration", "dont_allow")

	resp, err := op.Call(context.Background(), svc, nil)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if resp.TaskID() != taskID {
		t.Fatalf("unexpected task id: %q", resp.TaskID())
	}
}

func TestOperationGenVideoVideoInput(t *testing.T) {
	const taskID = "chatvideo-vid-123"
	videoBytes := []byte{0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70} // minimal MP4 header

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/videos/generations") {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)

		instances, ok := body["instances"].([]any)
		if !ok || len(instances) != 1 {
			t.Fatalf("unexpected instances: %#v", body["instances"])
		}
		inst0, ok := instances[0].(map[string]any)
		if !ok {
			t.Fatalf("unexpected instance type: %T", instances[0])
		}
		if inst0["prompt"] != "Continue the scene" {
			t.Fatalf("unexpected prompt: %#v", inst0["prompt"])
		}
		videoObj, ok := inst0["video"].(map[string]any)
		if !ok {
			t.Fatalf("missing video in instance: %#v", inst0)
		}
		if videoObj["bytesBase64Encoded"] == nil && videoObj["uri"] == nil {
			t.Fatalf("video must have bytesBase64Encoded or uri: %#v", videoObj)
		}
		if videoObj["mimeType"] != "video/mp4" {
			t.Fatalf("unexpected mimeType: %#v", videoObj["mimeType"])
		}

		_, _ = w.Write([]byte(`{"id":"` + taskID + `"}`))
	}))
	defer ts.Close()

	svc := NewService("token-1", WithBaseURL(ts.URL+"/v1/"), WithHTTPClient(ts.Client()))
	vid := svc.VideoFromBytes(xai.VideoMP4, videoBytes)

	op, err := svc.Operation("veo-3.0-generate-preview", xai.GenVideo)
	if err != nil {
		t.Fatalf("Operation failed: %v", err)
	}
	op.Params().
		Set("Video", vid).
		Set("Prompt", "Continue the scene").
		Set("AspectRatio", "16:9 (landscape)").
		Set("DurationSeconds", int32(6)).
		Set("PersonGeneration", "dont_allow")

	resp, err := op.Call(context.Background(), svc, nil)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if resp.TaskID() != taskID {
		t.Fatalf("unexpected task id: %q", resp.TaskID())
	}
}

func TestOperationGenVideoReferenceImages(t *testing.T) {
	const taskID = "chatvideo-ref-123"
	imgURL := "https://aitoken-public.qnaigc.com/example/generate-video/running-man.jpg"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/videos/generations") {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body["model"] != "veo-3.1-generate-preview" {
			t.Fatalf("unexpected model: %#v", body["model"])
		}
		instances, ok := body["instances"].([]any)
		if !ok || len(instances) != 1 {
			t.Fatalf("unexpected instances: %#v", body["instances"])
		}
		inst0, ok := instances[0].(map[string]any)
		if !ok {
			t.Fatalf("unexpected instance type: %T", instances[0])
		}
		if inst0["prompt"] != "A cinematic scene" {
			t.Fatalf("unexpected prompt: %#v", inst0["prompt"])
		}
		refImgs, ok := inst0["referenceImages"].([]any)
		if !ok || len(refImgs) < 1 {
			t.Fatalf("missing referenceImages: %#v", inst0["referenceImages"])
		}
		ref0, ok := refImgs[0].(map[string]any)
		if !ok {
			t.Fatalf("unexpected ref type: %T", refImgs[0])
		}
		if ref0["referenceType"] != "asset" {
			t.Fatalf("unexpected referenceType: %#v", ref0["referenceType"])
		}
		params, _ := body["parameters"].(map[string]any)
		if params["durationSeconds"] != float64(8) {
			t.Fatalf("durationSeconds must be 8 for referenceImages: %#v", params["durationSeconds"])
		}

		_, _ = w.Write([]byte(`{"id":"` + taskID + `"}`))
	}))
	defer ts.Close()

	svc := NewService("token-1", WithBaseURL(ts.URL+"/v1/"), WithHTTPClient(ts.Client()))
	img := svc.ImageFromStgUri(xai.ImageJPEG, imgURL)
	refs := svc.GenVideoReferenceImages(
		xai.GenVideoReferenceImage{Image: img, ReferenceType: "asset"},
	)

	op, err := svc.Operation("veo-3.1-generate-preview", xai.GenVideo)
	if err != nil {
		t.Fatalf("Operation failed: %v", err)
	}
	op.Params().
		Set("Prompt", "A cinematic scene").
		Set("ReferenceImages", refs).
		Set("AspectRatio", "16:9 (landscape)").
		Set("DurationSeconds", int32(8)).
		Set("PersonGeneration", "dont_allow")

	resp, err := op.Call(context.Background(), svc, nil)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if resp.TaskID() != taskID {
		t.Fatalf("unexpected task id: %q", resp.TaskID())
	}
}

func TestOperationGenVideo_Validation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach server when validation fails")
	}))
	defer ts.Close()

	svc := NewService("token-1", WithBaseURL(ts.URL+"/v1/"), WithHTTPClient(ts.Client()))

	tests := []struct {
		name string
		set  func(xai.Params)
		want string
	}{
		{
			name: "invalid DurationSeconds",
			set: func(p xai.Params) {
				p.Set("Prompt", "a cat").Set("DurationSeconds", int32(10))
			},
			want: "DurationSeconds",
		},
		{
			name: "invalid NumberOfVideos",
			set: func(p xai.Params) {
				p.Set("Prompt", "a cat").Set("NumberOfVideos", int32(5))
			},
			want: "NumberOfVideos",
		},
		{
			name: "invalid Seed",
			set: func(p xai.Params) {
				p.Set("Prompt", "a cat").Set("NumberOfVideos", int32(1)).Set("Seed", int32(-1))
			},
			want: "Seed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op, _ := svc.Operation("veo-3.0-generate-preview", xai.GenVideo)
			tt.set(op.Params())
			_, err := op.Call(context.Background(), svc, nil)
			if err == nil {
				t.Fatalf("expected validation error containing %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.want)
			}
		})
	}
}

func TestGenChat(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["model"] != "gemini-2.5-flash-image" {
			t.Fatalf("unexpected model: %#v", body["model"])
		}
		if body["stream"] != false {
			t.Fatalf("unexpected stream: %#v", body["stream"])
		}
		_, _ = w.Write([]byte(`{
			"choices": [{
				"index": 0,
				"finish_reason": "stop",
				"message": {
					"role": "assistant",
					"content": "ok",
					"images": [{
						"type": "image_url",
						"image_url": {"url":"data:image/png;base64,aGVsbG8="}
					}]
				}
			}],
			"usage": {"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30}
		}`))
	}))
	defer ts.Close()

	svc := NewService("token-1", WithBaseURL(ts.URL+"/v1/"), WithHTTPClient(ts.Client()))
	params := svc.Params().
		Model("gemini-2.5-flash-image").
		Messages(svc.UserMsg().Text("draw a cat"))
	resp, err := svc.Gen(context.Background(), params, nil)
	if err != nil {
		t.Fatalf("Gen failed: %v", err)
	}
	if resp.Len() != 1 {
		t.Fatalf("unexpected candidates len: %d", resp.Len())
	}
	cand := resp.At(0)
	if cand.Parts() != 2 {
		t.Fatalf("unexpected parts len: %d", cand.Parts())
	}
	if got := cand.Part(0).Text(); got != "ok" {
		t.Fatalf("unexpected text: %q", got)
	}
	blob, ok := cand.Part(1).AsBlob()
	if !ok || blob.MIME != "image/png" {
		t.Fatalf("unexpected blob: ok=%v mime=%q", ok, blob.MIME)
	}
}

func TestGenStream(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"index":0,"delta":{"role":"assistant","content":"hello "},"finish_reason":""}]}`)
		_, _ = fmt.Fprintf(w, "data: %s\n\n", `{"choices":[{"index":0,"delta":{"content":"world"},"finish_reason":"stop"}]}`)
		_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	}))
	defer ts.Close()

	svc := NewService("token-1", WithBaseURL(ts.URL+"/v1/"), WithHTTPClient(ts.Client()))
	params := svc.Params().
		Model("gemini-3.0-pro-image-preview").
		Messages(svc.UserMsg().Text("say hello"))

	var got strings.Builder
	for chunk, err := range svc.GenStream(context.Background(), params, nil) {
		if err != nil {
			t.Fatalf("GenStream failed: %v", err)
		}
		if chunk == nil || chunk.Len() == 0 {
			continue
		}
		cand := chunk.At(0)
		if cand.Parts() == 0 {
			continue
		}
		got.WriteString(cand.Part(0).Text())
	}
	if got.String() != "hello world" {
		t.Fatalf("unexpected stream text: %q", got.String())
	}
}
