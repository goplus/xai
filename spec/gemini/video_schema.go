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

package gemini

import (
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/types"
)

// Video param constants are in params.go.

// VideoSchemaFor returns the VideoSchema for the given Gemini/Veo model.
// Returns nil if the model is not a video model.
func VideoSchemaFor(model string) xai.VideoSchema {
	m := strings.ToLower(strings.TrimSpace(model))
	if !strings.HasPrefix(m, "veo-") {
		return nil
	}
	return &geminiVideoSchema{model: m}
}

// geminiVideoSchema implements xai.VideoSchema for Gemini/Veo models.
type geminiVideoSchema struct {
	model string
}

// SupportedModes returns the video generation modes supported by this model.
func (s *geminiVideoSchema) SupportedModes() []xai.VideoGenMode {
	modes := []xai.VideoGenMode{
		xai.VideoGenModeText,
		xai.VideoGenModeImage,
		xai.VideoGenModeStartEnd,
	}
	if s.supportsReferenceImages() {
		modes = append(modes, xai.VideoGenModeMultiRef)
	}
	return modes
}

// supportsReferenceImages returns true if the model supports reference images.
func (s *geminiVideoSchema) supportsReferenceImages() bool {
	return strings.Contains(s.model, "veo-2.0-generate-exp") ||
		strings.Contains(s.model, "veo-3.1-generate-preview")
}

// Fields returns all input fields for this model.
func (s *geminiVideoSchema) Fields() []xai.Field {
	fields := []xai.Field{
		{Name: ParamPrompt, Kind: types.String},
		{Name: ParamImage, Kind: types.Image},
		{Name: ParamVideo, Kind: types.Video},
		{Name: ParamLastFrame, Kind: types.Image},
		{Name: ParamAspectRatio, Kind: types.String},
		{Name: ParamResolution, Kind: types.String},
		{Name: ParamNegativePrompt, Kind: types.String},
		{Name: ParamNumberOfVideos, Kind: types.Int},
		{Name: ParamPersonGeneration, Kind: types.String},
		{Name: ParamDurationSeconds, Kind: types.Int},
		{Name: ParamSeed, Kind: types.Int},
		{Name: ParamGenerateAudio, Kind: types.Bool},
		{Name: ParamEnhancePrompt, Kind: types.Bool},
		{Name: ParamCompressionQuality, Kind: types.String},
	}
	if s.supportsReferenceImages() {
		fields = append(fields, xai.Field{Name: ParamReferenceImages, Kind: types.GenVideoReferenceImage | types.List})
	}
	return fields
}

// Restrict returns the restriction for a field.
func (s *geminiVideoSchema) Restrict(name string) *xai.Restriction {
	return restriction_genVideo[name]
}

// FieldModes returns the modes that a field is applicable to.
// Returns nil if the field is applicable to all modes.
func (s *geminiVideoSchema) FieldModes(name string) []xai.VideoGenMode {
	switch name {
	case ParamImage:
		return []xai.VideoGenMode{xai.VideoGenModeImage, xai.VideoGenModeStartEnd}
	case ParamLastFrame:
		return []xai.VideoGenMode{xai.VideoGenModeStartEnd}
	case ParamReferenceImages:
		return []xai.VideoGenMode{xai.VideoGenModeMultiRef}
	default:
		return nil
	}
}
