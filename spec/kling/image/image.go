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

package image

import (
	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling/internal"
	"github.com/goplus/xai/types"
)

var (
	limitAspectRatio   = &xai.StringEnum{Values: []string{internal.Aspect1x1, internal.Aspect16x9, internal.Aspect9x16, internal.Aspect4x3, internal.Aspect3x4, internal.Aspect3x2, internal.Aspect2x3, internal.Aspect21x9}}
	limitAspectRatioO1 = &xai.StringEnum{Values: []string{internal.AspectAuto, internal.Aspect1x1, internal.Aspect16x9, internal.Aspect9x16, internal.Aspect4x3, internal.Aspect3x4, internal.Aspect3x2, internal.Aspect2x3, internal.Aspect21x9}}
	limitImageRef      = &xai.StringEnum{Values: []string{internal.ImageRefSubject, internal.ImageRefFace}}
	limitResolution    = &xai.StringEnum{Values: []string{internal.Resolution1K, internal.Resolution2K, internal.Resolution4K}}
)

// SchemaForImage returns the InputSchema fields for the given image model.
// Returns defaultImageSchema if the model is unknown.
func SchemaForImage(model string) []xai.Field {
	switch model {
	case internal.ModelKlingV1:
		return SchemaV1Image()
	case internal.ModelKlingV15:
		return SchemaV15Image()
	case internal.ModelKlingV2:
		return SchemaV2Image()
	case internal.ModelKlingV2New:
		return SchemaV2NewImage()
	case internal.ModelKlingV21:
		return SchemaV21Image()
	case internal.ModelKlingImageO1:
		return SchemaImageO1()
	default:
		return defaultImageSchema()
	}
}

// Restrict returns the Restriction for the given param name on image models.
// Returns nil if the param has no restriction.
func Restrict(model, name string) *xai.Restriction {
	switch name {
	case internal.ParamAspectRatio:
		if model == internal.ModelKlingImageO1 {
			return &xai.Restriction{Limit: limitAspectRatioO1}
		}
		return &xai.Restriction{Limit: limitAspectRatio}
	case internal.ParamImageReference:
		if model == internal.ModelKlingV15 {
			return &xai.Restriction{Limit: limitImageRef}
		}
	case internal.ParamResolution:
		if model == internal.ModelKlingImageO1 {
			return &xai.Restriction{Limit: limitResolution}
		}
	}
	return nil
}

// defaultImageSchema returns the union of all image fields for fallback.
func defaultImageSchema() []xai.Field {
	return []xai.Field{
		{Name: internal.ParamPrompt, Kind: types.String},
		{Name: internal.ParamAspectRatio, Kind: types.String},
		{Name: internal.ParamReferenceImages, Kind: types.String | types.List},
		{Name: internal.ParamImage, Kind: types.String},
		{Name: internal.ParamImageReference, Kind: types.String},
		{Name: internal.ParamSubjectImageList, Kind: types.List},
		{Name: internal.ParamSceneImage, Kind: types.String},
		{Name: internal.ParamStyleImage, Kind: types.String},
		{Name: internal.ParamNegativePrompt, Kind: types.String},
		{Name: internal.ParamImageFidelity, Kind: types.Float},
		{Name: internal.ParamHumanFidelity, Kind: types.Float},
		{Name: internal.ParamN, Kind: types.Int},
		{Name: internal.ParamResolution, Kind: types.String},
	}
}
