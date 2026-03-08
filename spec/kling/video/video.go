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

package video

import (
	"errors"
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling/internal"
	"github.com/goplus/xai/types"
)

var (
	limitMode              = &xai.StringEnum{Values: []string{internal.ModeStd, internal.ModePro}}
	limitSeconds           = &xai.StringEnum{Values: []string{internal.Seconds5, internal.Seconds10}}
	limitSecondsV3         = &xai.StringEnum{Values: []string{"3", "4", "5", "6", "7", "8", "9", "10", "11", "12", "13", "14", "15"}}
	limitSize              = &xai.StringEnum{Values: []string{internal.Size1920x1080, internal.Size1080x1920, internal.Size1280x720, internal.Size720x1280, internal.Size1080x1080, internal.Size720x720}}
	limitSound             = &xai.StringEnum{Values: []string{internal.SoundOn, internal.SoundOff}}
	limitKeepOriginalSound = &xai.StringEnum{Values: []string{string(KeepOriginalSoundYes), string(KeepOriginalSoundNo)}}
)

// ErrInputReferenceRequired is returned when kling-v2-1 is used without input_reference.
var ErrInputReferenceRequired = errors.New("kling: input_reference is required for kling-v2-1 (img2video only)")

// ErrKeyframeModeRequired is returned when image_tail is set but mode is not "pro".
var ErrKeyframeModeRequired = errors.New("kling: mode must be \"pro\" when using image_tail (keyframe mode)")

// ErrKeyframeSecondsRequired is returned when kling-v2-1 keyframe mode uses seconds other than "10".
var ErrKeyframeSecondsRequired = errors.New("kling: kling-v2-1 keyframe mode requires seconds=\"10\"")

// ParamsChecker is used by Validate to check params without importing kling.Params.
// Implementations must provide HasNonEmptyString and GetString for video validation.
type ParamsChecker interface {
	HasNonEmptyString(name string) bool
	GetString(name string) string
}

// SchemaForVideo returns the InputSchema fields for the given video model.
// Returns defaultVideoSchema if the model is unknown.
func SchemaForVideo(model string) []xai.Field {
	switch model {
	case internal.ModelKlingV21Video:
		return SchemaV21Video()
	case internal.ModelKlingV25Turbo:
		return SchemaV25TurboVideo()
	case internal.ModelKlingVideoO1:
		return SchemaVideoO1()
	case internal.ModelKlingV26, internal.ModelKlingV27, internal.ModelKlingV28, internal.ModelKlingV29:
		return SchemaV26Video()
	case internal.ModelKlingV3:
		return SchemaV3Video()
	case internal.ModelKlingV3Omni:
		return SchemaV3OmniVideo()
	default:
		return defaultVideoSchema()
	}
}

// Restrict returns the Restriction for the given param name on video models.
// Returns nil if the param has no restriction.
func Restrict(model, name string) *xai.Restriction {
	switch name {
	case internal.ParamMode:
		return &xai.Restriction{Limit: limitMode}
	case internal.ParamSeconds:
		if isKlingV3OrOmni(model) {
			return &xai.Restriction{Limit: limitSecondsV3}
		}
		return &xai.Restriction{Limit: limitSeconds}
	case internal.ParamSize:
		return &xai.Restriction{Limit: limitSize}
	case internal.ParamSound:
		if isKlingV26OrNewer(model) || isKlingV3OrOmni(model) {
			return &xai.Restriction{Limit: limitSound}
		}
	case internal.ParamKeepOriginalSound:
		if model == internal.ModelKlingVideoO1 || isKlingV26OrNewer(model) || model == internal.ModelKlingV3Omni {
			return &xai.Restriction{Limit: limitKeepOriginalSound}
		}
	}
	return nil
}

// Validate runs model-specific validation for video generation.
// Returns ErrInputReferenceRequired, ErrKeyframeModeRequired, or ErrKeyframeSecondsRequired when constraints are violated.
func Validate(model string, p ParamsChecker) error {
	m := strings.ToLower(model)
	if err := validateRequiredParams(m, p); err != nil {
		return err
	}
	return validateKeyframeConstraints(m, p)
}

func isKlingV26OrNewer(model string) bool {
	m := strings.ToLower(model)
	return m == internal.ModelKlingV26 || m == internal.ModelKlingV27 ||
		m == internal.ModelKlingV28 || m == internal.ModelKlingV29
}

func isKlingV3OrOmni(model string) bool {
	m := strings.ToLower(model)
	return m == internal.ModelKlingV3 || m == internal.ModelKlingV3Omni
}

func validateRequiredParams(model string, p ParamsChecker) error {
	if model == internal.ModelKlingV21Video && !p.HasNonEmptyString(internal.ParamInputReference) {
		return ErrInputReferenceRequired
	}
	return nil
}

func validateKeyframeConstraints(model string, p ParamsChecker) error {
	if !p.HasNonEmptyString(internal.ParamImageTail) {
		return nil
	}
	mode := p.GetString(internal.ParamMode)
	if mode != internal.ModePro {
		return ErrKeyframeModeRequired
	}
	if model == internal.ModelKlingV21Video {
		sec := p.GetString(internal.ParamSeconds)
		if sec != internal.Seconds10 {
			return ErrKeyframeSecondsRequired
		}
	}
	return nil
}

// defaultVideoSchema returns the union of common video fields for fallback.
func defaultVideoSchema() []xai.Field {
	return []xai.Field{
		{Name: internal.ParamPrompt, Kind: types.String},
		{Name: internal.ParamInputReference, Kind: types.String},
		{Name: internal.ParamImageTail, Kind: types.String},
		{Name: internal.ParamNegativePrompt, Kind: types.String},
		{Name: internal.ParamMode, Kind: types.String},
		{Name: internal.ParamSeconds, Kind: types.String},
		{Name: internal.ParamSize, Kind: types.String},
		{Name: internal.ParamImageList, Kind: types.List},
		{Name: internal.ParamVideoList, Kind: types.List},
		{Name: internal.ParamSound, Kind: types.String},
	}
}
