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

// Video schema for Vidu models. Param constants are in params.go.

package vidu

import (
	xai "github.com/goplus/xai/spec"
)

// VideoSchemaFor returns the VideoSchema for the given Vidu model.
// Returns nil if the model is not a video model.
func VideoSchemaFor(model string) xai.VideoSchema {
	m := normalizeModel(model)
	if !IsVideoModel(m) {
		return nil
	}
	return &viduVideoSchema{model: m}
}

// viduVideoSchema implements xai.VideoSchema for Vidu models.
type viduVideoSchema struct {
	model string
}

// SupportedModes returns the video generation modes supported by this model.
func (s *viduVideoSchema) SupportedModes() []xai.VideoGenMode {
	switch s.model {
	case ModelViduQ1:
		return []xai.VideoGenMode{
			xai.VideoGenModeText,
			xai.VideoGenModeMultiRef,
		}
	case ModelViduQ2, ModelViduQ2Turbo, ModelViduQ2Pro:
		return []xai.VideoGenMode{
			xai.VideoGenModeText,
			xai.VideoGenModeImage,
			xai.VideoGenModeMultiRef,
			xai.VideoGenModeStartEnd,
		}
	default:
		return []xai.VideoGenMode{xai.VideoGenModeText}
	}
}

// Fields returns all input fields for this model.
func (s *viduVideoSchema) Fields() []xai.Field {
	return SchemaForVideo(s.model)
}

// Restrict returns the restriction for a field.
func (s *viduVideoSchema) Restrict(name string) *xai.Restriction {
	return Restrict(s.model, name)
}

// FieldModes returns the modes that a field is applicable to.
// Returns nil if the field is applicable to all modes.
func (s *viduVideoSchema) FieldModes(name string) []xai.VideoGenMode {
	switch name {
	case ParamImageURL:
		return []xai.VideoGenMode{xai.VideoGenModeImage}
	case ParamReferenceImageURLs, ParamSubjects:
		return []xai.VideoGenMode{xai.VideoGenModeMultiRef}
	case ParamStartImageURL, ParamEndImageURL:
		return []xai.VideoGenMode{xai.VideoGenModeStartEnd}
	default:
		return nil
	}
}
