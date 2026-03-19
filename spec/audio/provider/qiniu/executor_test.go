package qiniu

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/audio"
)

func TestToAudioMapUsesFallbackFormat(t *testing.T) {
	got := toAudioMap("https://example.com/audio.wav", "wav")
	if got["format"] != "wav" {
		t.Fatalf("expected fallback format wav, got %q", got["format"])
	}
	if got["url"] != "https://example.com/audio.wav" {
		t.Fatalf("unexpected url: %q", got["url"])
	}
}

func TestToAudioMapBuildsDataURIWithFallbackFormat(t *testing.T) {
	got := toAudioMap("ZmFrZS1hdWRpby1ieXRlcw==", "wav")
	if got["format"] != "wav" {
		t.Fatalf("expected fallback format wav, got %q", got["format"])
	}
	if got["url"] != "data:audio/wav;base64,ZmFrZS1hdWRpby1ieXRlcw==" {
		t.Fatalf("unexpected data uri: %q", got["url"])
	}
}

func TestToAudioMapPrefersEmbeddedFormat(t *testing.T) {
	got := toAudioMap(map[string]any{
		"format": "flac",
		"url":    "https://example.com/audio.flac",
	}, "wav")
	if got["format"] != "flac" {
		t.Fatalf("expected embedded format flac, got %q", got["format"])
	}
}

func TestASRExecutorTranscribeMapsRequestAndResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != EndpointASR {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		audioBody := body["audio"].(map[string]any)
		if got := audioBody["format"]; got != "wav" {
			t.Fatalf("unexpected audio format: %#v", got)
		}
		if got := audioBody["url"]; got != "https://example.com/in.wav" {
			t.Fatalf("unexpected audio url: %#v", got)
		}
		_, _ = w.Write([]byte(`{"data":{"result":{"text":"hello world"},"duration":1.25}}`))
	}))
	defer ts.Close()

	client := NewClient("token",
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
		WithLogger(log.New(io.Discard, "", 0)),
		WithDebugLog(false),
	)
	exec := NewASRExecutor(client)
	params := audio.NewParams().
		Set(audio.ParamAudio, "https://example.com/in.wav").
		Set(audio.ParamFormat, "wav")

	resp, err := exec.Transcribe(context.Background(), xai.Model(audio.ModelASR), params)
	if err != nil {
		t.Fatalf("Transcribe failed: %v", err)
	}
	out := resp.Results().At(0).(*xai.OutputText)
	if out.Text != "hello world" {
		t.Fatalf("unexpected transcribed text: %q", out.Text)
	}
	if out.Duration == nil || *out.Duration != 1.25 {
		t.Fatalf("unexpected duration: %#v", out.Duration)
	}
}

func TestTTSExecutorSynthesizeUsesDefaults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != EndpointTTS {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request failed: %v", err)
		}
		audioBody := body["audio"].(map[string]any)
		if got := audioBody["voice_type"]; got != "qiniu_zh_female_wwxkjx" {
			t.Fatalf("unexpected default voice: %#v", got)
		}
		if got := audioBody["encoding"]; got != "mp3" {
			t.Fatalf("unexpected default encoding: %#v", got)
		}
		if got := audioBody["speed_ratio"]; got != 1.0 {
			t.Fatalf("unexpected default speed ratio: %#v", got)
		}
		requestBody := body["request"].(map[string]any)
		if got := requestBody["text"]; got != "hello there" {
			t.Fatalf("unexpected request text: %#v", got)
		}
		_, _ = w.Write([]byte(`{"data":"https://example.com/out.mp3","addition":{"duration":"2.4"}}`))
	}))
	defer ts.Close()

	client := NewClient("token",
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
		WithLogger(log.New(io.Discard, "", 0)),
		WithDebugLog(false),
	)
	exec := NewTTSExecutor(client)
	params := audio.NewParams().Set(audio.ParamInput, "hello there")

	resp, err := exec.Synthesize(context.Background(), xai.Model(audio.ModelTTSV1), params)
	if err != nil {
		t.Fatalf("Synthesize failed: %v", err)
	}
	out := resp.Results().At(0).(*xai.OutputAudio)
	if out.Audio != "https://example.com/out.mp3" || out.Format != "mp3" || out.Duration != "2.4" {
		t.Fatalf("unexpected synthesized output: %#v", out)
	}
}

func TestVoiceListerSupportsWrappedAndFlatResponses(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "wrapped",
			body: `{"data":[{"voice_name":"A","voice_type":"voice-a","url":"https://example.com/a.mp3","category":"demo","updatetime":1}]}`,
		},
		{
			name: "flat",
			body: `[{"voice_name":"B","voice_type":"voice-b","url":"https://example.com/b.mp3","category":"demo","updatetime":2}]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != EndpointVoiceList {
					t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
				}
				_, _ = w.Write([]byte(tt.body))
			}))
			defer ts.Close()

			client := NewClient("token",
				WithBaseURL(ts.URL),
				WithHTTPClient(ts.Client()),
				WithLogger(log.New(io.Discard, "", 0)),
				WithDebugLog(false),
			)
			lister := NewVoiceLister(client)

			voices, err := lister.ListVoices(context.Background())
			if err != nil {
				t.Fatalf("ListVoices failed: %v", err)
			}
			if len(voices) != 1 {
				t.Fatalf("expected 1 voice, got %d", len(voices))
			}
			if voices[0].VoiceType == "" {
				t.Fatalf("expected non-empty voice type: %#v", voices[0])
			}
		})
	}
}

func TestNewServiceSupportsSetApiKeyAndVoiceListing(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != EndpointVoiceList {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer updated-token" {
			t.Fatalf("unexpected auth header after SetApiKey: %q", got)
		}
		_, _ = w.Write([]byte(`[{"voice_name":"A","voice_type":"voice-a","url":"https://example.com/a.mp3","category":"demo","updatetime":1}]`))
	}))
	defer ts.Close()

	svc := NewService("initial-token",
		WithBaseURL(ts.URL),
		WithHTTPClient(ts.Client()),
		WithLogger(log.New(io.Discard, "", 0)),
		WithDebugLog(false),
	)
	svc.SetApiKey("updated-token")

	voices, err := svc.ListVoices(context.Background())
	if err != nil {
		t.Fatalf("ListVoices failed: %v", err)
	}
	if len(voices) != 1 || voices[0].VoiceType != "voice-a" {
		t.Fatalf("unexpected voices: %#v", voices)
	}
}
