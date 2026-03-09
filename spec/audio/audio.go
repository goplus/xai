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

	"github.com/goplus/xai/types"
	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/audio/internal"
)

// Scheme is the URI scheme for Audio: "audio".
const Scheme = "audio"

// Model constants (re-exported from internal).
const (
	ModelASR    = internal.ModelASR
	ModelTTSV1  = internal.ModelTTSV1
)

// IsASRModel returns true if the model supports ASR (Transcribe).
func IsASRModel(m string) bool { return m == internal.ModelASR }

// IsTTSModel returns true if the model supports TTS (Synthesize).
func IsTTSModel(m string) bool { return m == internal.ModelTTSV1 }

// ASRModels returns all supported ASR model IDs.
func ASRModels() []string { return []string{internal.ModelASR} }

// TTSModels returns all supported TTS model IDs.
func TTSModels() []string { return []string{internal.ModelTTSV1} }

// SchemaForTranscribe returns the InputSchema fields for Transcribe.
func SchemaForTranscribe(model string) []xai.Field {
	return []xai.Field{
		{Name: internal.ParamAudio, Kind: types.String | types.List},
		{Name: internal.ParamModel, Kind: types.String},
		{Name: internal.ParamFormat, Kind: types.String},
	}
}

// SchemaForSynthesize returns the InputSchema fields for Synthesize.
func SchemaForSynthesize(model string) []xai.Field {
	return []xai.Field{
		{Name: internal.ParamInput, Kind: types.String},
		{Name: internal.ParamVoice, Kind: types.String},
		{Name: internal.ParamFormat, Kind: types.String},
		{Name: internal.ParamSpeed, Kind: types.Float},
	}
}

// Register registers the Audio service with xai.
//
// Example:
//
//	svc := audio.NewService(asrExec, ttsExec)
//	audio.Register(svc)
//	// Then xai.New(ctx, "audio://") returns svc
func Register(svc *Service) {
	xai.Register(Scheme, func(ctx context.Context, uri string) (xai.Service, error) {
		return svc, nil
	})
}
