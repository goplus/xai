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
		Set("AspectRatio", "16:9")
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
		Set("References", []genai.ReferenceImage{ref.(genai.ReferenceImage)})
	resp, err := op.Call(context.Background(), svc, nil)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if resp.Results().Len() != 1 {
		t.Fatalf("unexpected results len: %d", resp.Results().Len())
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
