package qiniu

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	xai "github.com/goplus/xai/spec"
)

func TestNewServiceWithErrorRequiresAPIKey(t *testing.T) {
	t.Setenv("QINIU_API_KEY", "")

	_, err := NewServiceWithError("")
	if err == nil || !strings.Contains(err.Error(), "API key is required") {
		t.Fatalf("expected missing API key error, got %v", err)
	}
}

func TestNewServiceWithErrorUsesEnvAndExtendedParsing(t *testing.T) {
	const apiKey = "env-token"
	t.Setenv("QINIU_API_KEY", apiKey)

	imageData := "data:image/png;base64," + base64.StdEncoding.EncodeToString([]byte("png-bytes"))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+apiKey {
			t.Fatalf("unexpected auth header: %q", got)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		if body["model"] != "gemini-2.5-flash-image" {
			t.Fatalf("unexpected model: %#v", body["model"])
		}

		_, _ = w.Write([]byte(`{
			"id":"chatcmpl-1",
			"choices":[
				{
					"index":0,
					"finish_reason":"stop",
					"message":{
						"role":"assistant",
						"content":"drawn",
						"images":[
							{"type":"image_url","image_url":{"url":"` + imageData + `"}}
						]
					}
				}
			]
		}`))
	}))
	defer ts.Close()

	svc, err := NewServiceWithError("", WithBaseURL(ts.URL+"/v1/"))
	if err != nil {
		t.Fatalf("NewServiceWithError failed: %v", err)
	}

	resp, err := svc.Gen(context.Background(),
		svc.Params().
			Model("gemini-2.5-flash-image").
			Messages(svc.UserMsg().Text("draw a cat")),
		nil,
	)
	if err != nil {
		t.Fatalf("Gen failed: %v", err)
	}

	cand := resp.At(0)
	if got := cand.Part(0).Text(); got != "drawn" {
		t.Fatalf("unexpected text part: %q", got)
	}
	if got := cand.Parts(); got != 2 {
		t.Fatalf("unexpected parts count: %d", got)
	}
	blob, ok := cand.Part(1).AsBlob()
	if !ok {
		t.Fatal("expected image blob part")
	}
	if got := blob.MIME; got != "image/png" {
		t.Fatalf("unexpected blob mime: %q", got)
	}
}

func TestRegisterUsesURIOverrides(t *testing.T) {
	const overrideKey = "override-token"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer "+overrideKey {
			t.Fatalf("unexpected auth header: %q", got)
		}
		_, _ = w.Write([]byte(`{
			"id":"chatcmpl-2",
			"choices":[
				{
					"index":0,
					"finish_reason":"stop",
					"message":{"role":"assistant","content":"ok"}
				}
			]
		}`))
	}))
	defer ts.Close()

	Register("fallback-token")

	svc, err := xai.New(context.Background(), "qiniu:key="+overrideKey+"&base="+url.QueryEscape(ts.URL+"/v1/"))
	if err != nil {
		t.Fatalf("xai.New failed: %v", err)
	}

	resp, err := svc.Gen(context.Background(),
		svc.Params().
			Model("gpt-4o-mini").
			Messages(svc.UserMsg().Text("hello")),
		nil,
	)
	if err != nil {
		t.Fatalf("Gen failed: %v", err)
	}
	if got := resp.At(0).Part(0).Text(); got != "ok" {
		t.Fatalf("unexpected response text: %q", got)
	}
}

func TestNewVideoServiceUsesEnvAndSupportsVideoOperations(t *testing.T) {
	const (
		apiKey = "video-env-token"
		taskID = "qvideo-test-123"
	)
	t.Setenv("QINIU_API_KEY", apiKey)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+apiKey {
			t.Fatalf("unexpected auth header: %q", got)
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/videos":
			_, _ = w.Write([]byte(`{
				"id":"` + taskID + `",
				"object":"video",
				"model":"sora-2",
				"status":"queued",
				"created_at":1766453713,
				"updated_at":1766453713
			}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/videos/"+taskID:
			_, _ = w.Write([]byte(`{
				"id":"` + taskID + `",
				"object":"video",
				"model":"sora-2",
				"status":"completed",
				"created_at":1766453713,
				"updated_at":1766453715,
				"completed_at":1766453715,
				"task_result":{"videos":[{"url":"https://example.com/out.mp4"}]}
			}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	svc := NewVideoService("", WithBaseURL(ts.URL+"/v1/"))

	_, err := svc.Gen(context.Background(),
		svc.Params().
			Model("gpt-4o-mini").
			Messages(svc.UserMsg().Text("hello")),
		nil,
	)
	if !errors.Is(err, xai.ErrNotSupported) {
		t.Fatalf("expected ErrNotSupported for video-only Gen, got %v", err)
	}

	opSvc, ok := any(svc).(interface {
		Operation(model xai.Model, action xai.Action) (xai.Operation, error)
	})
	if !ok {
		t.Fatal("service does not implement operation service")
	}
	op, err := opSvc.Operation("sora-2", xai.GenVideo)
	if err != nil {
		t.Fatalf("Operation failed: %v", err)
	}
	op.Params().
		Set("Prompt", "A cat jumps over a puddle").
		Set("Seconds", "4").
		Set("Size", "1280x720")

	resp, err := op.Call(context.Background(), svc, nil)
	if err != nil {
		t.Fatalf("video Call failed: %v", err)
	}
	if got := resp.TaskID(); got != taskID {
		t.Fatalf("unexpected task id: %q", got)
	}
	resp, err = resp.Retry(context.Background(), svc)
	if err != nil {
		t.Fatalf("Retry failed: %v", err)
	}
	if !resp.Done() {
		t.Fatal("expected completed response after retry")
	}
}

func TestRegisterVideoUsesURIOverrides(t *testing.T) {
	const (
		overrideKey = "video-override-token"
		taskID      = "qvideo-override-1"
	)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+overrideKey {
			t.Fatalf("unexpected auth header: %q", got)
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/videos":
			_, _ = w.Write([]byte(`{
				"id":"` + taskID + `",
				"object":"video",
				"model":"sora-2",
				"status":"queued",
				"created_at":1766453713,
				"updated_at":1766453713
			}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1/videos/"+taskID:
			_, _ = w.Write([]byte(`{
				"id":"` + taskID + `",
				"object":"video",
				"model":"sora-2",
				"status":"completed",
				"created_at":1766453713,
				"updated_at":1766453715,
				"completed_at":1766453715,
				"task_result":{"videos":[{"url":"https://example.com/final.mp4"}]}
			}`))
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer ts.Close()

	RegisterVideo("fallback-video-token")

	svc, err := xai.New(context.Background(), "qiniu-video:key="+overrideKey+"&base="+url.QueryEscape(ts.URL+"/v1/"))
	if err != nil {
		t.Fatalf("xai.New failed: %v", err)
	}

	opSvc, ok := svc.(interface {
		Operation(model xai.Model, action xai.Action) (xai.Operation, error)
	})
	if !ok {
		t.Fatal("registered video service does not implement operation service")
	}
	op, err := opSvc.Operation("sora-2", xai.GenVideo)
	if err != nil {
		t.Fatalf("Operation failed: %v", err)
	}
	op.Params().Set("Prompt", "A fox runs through snow")

	resp, err := op.Call(context.Background(), svc, nil)
	if err != nil {
		t.Fatalf("Call failed: %v", err)
	}
	if got := resp.TaskID(); got != taskID {
		t.Fatalf("unexpected task id: %q", got)
	}
}
