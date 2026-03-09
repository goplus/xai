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

	xai "github.com/goplus/xai/spec"
)

// NewOutputText creates xai.Results from ASR transcribed text.
func NewOutputText(text string, duration *float64) xai.Results {
	return &outputText{text: text, duration: duration}
}

// NewOutputAudio creates xai.Results from TTS synthesized audio.
func NewOutputAudio(audio, format, duration string) xai.Results {
	return &outputAudio{audio: audio, format: format, duration: duration}
}

// SyncOperationResponse is a synchronous xai.OperationResponse (Done()==true).
type SyncOperationResponse struct {
	R xai.Results
}

func (p *SyncOperationResponse) Done() bool   { return true }
func (p *SyncOperationResponse) Sleep()       {}
func (p *SyncOperationResponse) Retry(ctx context.Context, svc xai.Service) (xai.OperationResponse, error) {
	return p, nil
}
func (p *SyncOperationResponse) Results() xai.Results { return p.R }
func (p *SyncOperationResponse) TaskID() string        { return "" }

// outputText implements xai.Results for ASR (Transcribe).
type outputText struct {
	text     string
	duration *float64
}

func (p *outputText) XGo_Attr(name string) any {
	switch name {
	case "text":
		return p.text
	case "duration":
		return p.duration
	}
	return nil
}

func (p *outputText) Len() int { return 1 }

func (p *outputText) At(i int) xai.Generated {
	if i != 0 {
		return nil
	}
	return &xai.OutputText{Text: p.text, Duration: p.duration}
}

// outputAudio implements xai.Results for TTS (Synthesize).
type outputAudio struct {
	audio    string
	format   string
	duration string
}

func (p *outputAudio) XGo_Attr(name string) any {
	switch name {
	case "audio":
		return p.audio
	case "format":
		return p.format
	case "duration":
		return p.duration
	}
	return nil
}

func (p *outputAudio) Len() int { return 1 }

func (p *outputAudio) At(i int) xai.Generated {
	if i != 0 {
		return nil
	}
	return &xai.OutputAudio{Audio: p.audio, Format: p.format, Duration: p.duration}
}
