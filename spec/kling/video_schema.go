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

// Video schema for Kling models. Param constants are in params.go.

package kling

import (
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling/video"
)

// VideoSchemaFor returns the VideoSchema for the given Kling model.
// Returns nil if the model is not a video model.
func VideoSchemaFor(model string) xai.VideoSchema {
	m := strings.ToLower(strings.TrimSpace(model))
	if !IsVideoModel(m) {
		return nil
	}
	return &klingVideoSchema{model: m}
}

// klingVideoSchema implements xai.VideoSchema for Kling models.
type klingVideoSchema struct {
	model string
}

// SupportedModes returns the video generation modes supported by this model.
func (s *klingVideoSchema) SupportedModes() []xai.VideoGenMode {
	switch s.model {
	case ModelKlingV21Video:
		return []xai.VideoGenMode{
			xai.VideoGenModeImage,
			xai.VideoGenModeStartEnd,
		}
	case ModelKlingV25Turbo:
		return []xai.VideoGenMode{
			xai.VideoGenModeText,
			xai.VideoGenModeImage,
			xai.VideoGenModeStartEnd,
		}
	case ModelKlingV26, ModelKlingV27, ModelKlingV28, ModelKlingV29:
		return []xai.VideoGenMode{
			xai.VideoGenModeText,
			xai.VideoGenModeImage,
			xai.VideoGenModeStartEnd,
		}
	case ModelKlingVideoO1:
		return []xai.VideoGenMode{
			xai.VideoGenModeText,
			xai.VideoGenModeImage,
			xai.VideoGenModeMultiRef,
			xai.VideoGenModeStartEnd,
		}
	case ModelKlingV3:
		return []xai.VideoGenMode{
			xai.VideoGenModeText,
			xai.VideoGenModeImage,
		}
	case ModelKlingV3Omni:
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
func (s *klingVideoSchema) Fields() []xai.Field {
	return video.SchemaForVideo(s.model)
}

// Restrict returns the restriction for a field.
func (s *klingVideoSchema) Restrict(name string) *xai.Restriction {
	return video.Restrict(s.model, name)
}

// FieldModes returns the modes that a field is applicable to.
// Returns nil if the field is applicable to all modes.
func (s *klingVideoSchema) FieldModes(name string) []xai.VideoGenMode {
	switch name {
	case ParamInputReference:
		return []xai.VideoGenMode{xai.VideoGenModeImage}
	case ParamImageTail:
		return []xai.VideoGenMode{xai.VideoGenModeStartEnd}
	case ParamImageList:
		switch s.model {
		case ModelKlingVideoO1:
			return []xai.VideoGenMode{xai.VideoGenModeMultiRef, xai.VideoGenModeStartEnd}
		case ModelKlingV3Omni:
			return []xai.VideoGenMode{xai.VideoGenModeImage, xai.VideoGenModeMultiRef, xai.VideoGenModeStartEnd}
		}
		return nil
	default:
		return nil
	}
}
