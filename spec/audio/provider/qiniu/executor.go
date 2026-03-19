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
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/audio"
)

const (
	EndpointASR       = "/voice/asr"
	EndpointTTS       = "/voice/tts"
	EndpointVoiceList = "/voice/list"
)

// ASRExecutor implements audio.ASRExecutor for Qiniu API.
type ASRExecutor struct {
	client *Client
}

// NewASRExecutor creates a new ASRExecutor.
func NewASRExecutor(client *Client) *ASRExecutor {
	return &ASRExecutor{client: client}
}

// Transcribe submits an ASR request and returns the result (sync).
func (e *ASRExecutor) Transcribe(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	audioParams, ok := params.(*audio.Params)
	if !ok {
		return nil, fmt.Errorf("qiniu: expected *audio.Params, got %T", params)
	}

	audioVal, _ := audioParams.Get(audio.ParamAudio)
	audioMap := toAudioMap(audioVal, audioParams.GetString(audio.ParamFormat))

	payload := map[string]interface{}{
		"model": "asr",
		"audio": audioMap,
	}

	var res struct {
		Data struct {
			Result struct {
				Text string `json:"text"`
			} `json:"result"`
			Duration  float64     `json:"duration"`
			AudioInfo interface{} `json:"audio_info"`
		} `json:"data"`
	}

	respBody, err := e.client.Post(ctx, EndpointASR, payload)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse ASR response: %w", err)
	}

	var duration *float64
	if res.Data.Duration > 0 {
		duration = &res.Data.Duration
	}

	r := audio.NewOutputText(res.Data.Result.Text, duration)
	return &audio.SyncOperationResponse{R: r}, nil
}

// TTSExecutor implements audio.TTSExecutor for Qiniu API.
type TTSExecutor struct {
	client *Client
}

// NewTTSExecutor creates a new TTSExecutor.
func NewTTSExecutor(client *Client) *TTSExecutor {
	return &TTSExecutor{client: client}
}

// Synthesize submits a TTS request and returns the result (sync).
func (e *TTSExecutor) Synthesize(ctx context.Context, model xai.Model, params xai.Params) (xai.OperationResponse, error) {
	audioParams, ok := params.(*audio.Params)
	if !ok {
		return nil, fmt.Errorf("qiniu: expected *audio.Params, got %T", params)
	}

	input := audioParams.GetString(audio.ParamInput)
	voice := audioParams.GetString(audio.ParamVoice)
	if voice == "" {
		voice = "qiniu_zh_female_wwxkjx"
	}
	format := audioParams.GetString(audio.ParamFormat)
	if format == "" {
		format = "mp3"
	}

	speedRatio := 1.0
	if v, ok := audioParams.Get(audio.ParamSpeed); ok {
		if f, ok := v.(float64); ok {
			speedRatio = f
		}
	}

	payload := map[string]interface{}{
		"audio": map[string]interface{}{
			"voice_type":  voice,
			"encoding":    format,
			"speed_ratio": speedRatio,
		},
		"request": map[string]interface{}{
			"text": input,
		},
	}

	var res struct {
		Data     string `json:"data"`
		Addition struct {
			Duration string `json:"duration"`
		} `json:"addition"`
	}

	respBody, err := e.client.Post(ctx, EndpointTTS, payload)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(respBody, &res); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse TTS response: %w", err)
	}

	r := audio.NewOutputAudio(res.Data, format, res.Addition.Duration)
	return &audio.SyncOperationResponse{R: r}, nil
}

// toAudioMap converts audio param to {format, url} map.
func toAudioMap(audioVal interface{}, fallbackFormat string) map[string]string {
	format := normalizeAudioFormat(fallbackFormat)
	if a, ok := audioVal.(map[string]interface{}); ok {
		if f, ok := a["format"].(string); ok && f != "" {
			format = normalizeAudioFormat(f)
		}
		url := ""
		if u, ok := a["url"].(string); ok {
			url = u
		}
		return map[string]string{"format": format, "url": url}
	}

	if a, ok := audioVal.(map[string]string); ok {
		if f, ok := a["format"]; ok && f != "" {
			format = normalizeAudioFormat(f)
		}
		url := ""
		if u, ok := a["url"]; ok {
			url = u
		}
		return map[string]string{"format": format, "url": url}
	}

	str, ok := audioVal.(string)
	if !ok {
		return map[string]string{"format": format, "url": ""}
	}

	if strings.HasPrefix(str, "http") {
		return map[string]string{"format": format, "url": str}
	}

	if strings.HasPrefix(str, "data:audio/") {
		return map[string]string{"format": format, "url": str}
	}

	return map[string]string{"format": format, "url": "data:audio/" + format + ";base64," + str}
}

func normalizeAudioFormat(format string) string {
	format = strings.TrimSpace(format)
	if format == "" {
		return "mp3"
	}
	return format
}

// VoiceLister implements audio.VoiceLister for Qiniu API.
type VoiceLister struct {
	client *Client
}

// NewVoiceLister creates a new VoiceLister.
func NewVoiceLister(client *Client) *VoiceLister {
	return &VoiceLister{client: client}
}

// ListVoices fetches available TTS voices from GET /voice/list.
func (e *VoiceLister) ListVoices(ctx context.Context) ([]audio.VoiceListItem, error) {
	respBody, err := e.client.Get(ctx, EndpointVoiceList)
	if err != nil {
		return nil, err
	}

	var raw interface{}
	if err := json.Unmarshal(respBody, &raw); err != nil {
		return nil, fmt.Errorf("qiniu: failed to parse voice list response: %w", err)
	}

	// Handle different response structures: {data: [...]} or [...]
	if dataMap, ok := raw.(map[string]interface{}); ok {
		if data, ok := dataMap["data"]; ok {
			raw = data
		}
	}

	var list []audio.VoiceListItem
	if dataList, ok := raw.([]interface{}); ok {
		b, _ := json.Marshal(dataList)
		if err := json.Unmarshal(b, &list); err != nil {
			return nil, fmt.Errorf("qiniu: failed to decode voice list: %w", err)
		}
	}

	return list, nil
}

// NewExecutors creates both ASRExecutor and TTSExecutor.
func NewExecutors(client *Client) (*ASRExecutor, *TTSExecutor) {
	return NewASRExecutor(client), NewTTSExecutor(client)
}

// Service wraps audio.Service with ApiKeySetter for runtime configuration.
type Service struct {
	*audio.Service
	client *Client
}

// SetApiKey updates the API key at runtime. Implements xai.ApiKeySetter.
func (s *Service) SetApiKey(apiKey string) {
	s.client.SetApiKey(apiKey)
}

// AudioService returns the embedded *audio.Service for operation.Call.
func (s *Service) AudioService() *audio.Service { return s.Service }

// NewService creates an audio.Service with Qiniu executors.
// Includes VoiceLister for ListVoices support.
// The returned *Service supports SetApiKey(apiKey) for runtime API key updates.
func NewService(apiKey string, opts ...ClientOption) *Service {
	client := NewClient(apiKey, opts...)
	asrExec, ttsExec := NewExecutors(client)
	voiceLister := NewVoiceLister(client)
	return &Service{
		Service: audio.NewService(asrExec, ttsExec, audio.WithVoiceLister(voiceLister)),
		client:  client,
	}
}

// Register creates an audio.Service with Qiniu executors and registers it with xai.
func Register(apiKey string, opts ...ClientOption) {
	svc := NewService(apiKey, opts...)
	audio.Register(svc)
}
