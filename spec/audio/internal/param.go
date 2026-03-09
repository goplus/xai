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

package internal

import "errors"

var (
	ErrAudioRequired = errors.New("audio: audio (URL or {format, url}) is required for Transcribe")
	ErrInputRequired = errors.New("audio: input (text) is required for Synthesize")
)

// Param name constants for ASR and TTS.
const (
	// ASR (Transcribe)
	ParamAudio  = "audio"  // string (URL) or map {format, url}
	ParamModel  = "model"  // optional model override
	ParamFormat = "format" // audio format, e.g. mp3

	// TTS (Synthesize)
	ParamInput  = "input"  // text to synthesize
	ParamVoice  = "voice"  // voice type, e.g. qiniu_zh_female_wwxkjx
	ParamSpeed  = "speed"  // speed ratio, float
)
