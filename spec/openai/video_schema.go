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

package openai

// Video param constants are in params.go.

import (
	"strings"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/types"
)

var (
	enumVideoSeconds = &xai.StringEnum{
		Values: []string{"4", "8", "12"},
	}
	enumVideoSizeSora2 = &xai.StringEnum{
		Values: []string{"720x1280", "1280x720"},
	}
	enumVideoSizeSora2Pro = &xai.StringEnum{
		Values: []string{"1280x720", "1024x1792", "1792x1024"},
	}
	enumVideoSize = &xai.StringEnum{
		Values: []string{"720x1280", "1280x720", "1024x1792", "1792x1024"},
	}
	videoRestrictions = map[string]*xai.Restriction{
		ParamPrompt:  {Required: true},
		ParamSeconds: {Limit: enumVideoSeconds},
		ParamSize:    {Limit: enumVideoSize},
	}
	videoFields = []xai.Field{
		{Name: ParamPrompt, Kind: types.String},
		{Name: ParamInputReference, Kind: types.String},
		{Name: ParamSeconds, Kind: types.String},
		{Name: ParamSize, Kind: types.String},
		{Name: ParamRemixFromVideoID, Kind: types.String},
	}
)

// VideoSchemaFor returns the VideoSchema for the given OpenAI/Sora model.
// Returns nil if the model is not a video model.
func VideoSchemaFor(model string) xai.VideoSchema {
	m := strings.ToLower(strings.TrimSpace(model))
	if !strings.HasPrefix(m, "sora-") {
		return nil
	}
	return &soraVideoSchema{model: m}
}

// soraVideoSchema implements xai.VideoSchema for OpenAI/Sora models.
type soraVideoSchema struct {
	model string
}

// SupportedModes returns the video generation modes supported by this model.
func (s *soraVideoSchema) SupportedModes() []xai.VideoGenMode {
	return []xai.VideoGenMode{
		xai.VideoGenModeText,
		xai.VideoGenModeImage,
	}
}

// Fields returns all input fields for this model.
func (s *soraVideoSchema) Fields() []xai.Field {
	return append([]xai.Field(nil), videoFields...)
}

// Restrict returns the restriction for a field.
func (s *soraVideoSchema) Restrict(name string) *xai.Restriction {
	r := videoRestrictions[name]
	if r == nil {
		return nil
	}
	if name == ParamSize {
		switch s.model {
		case ModelSora2:
			return &xai.Restriction{Limit: enumVideoSizeSora2}
		case ModelSora2Pro:
			return &xai.Restriction{Limit: enumVideoSizeSora2Pro}
		}
	}
	return r
}

// FieldModes returns the modes that a field is applicable to.
// Returns nil if the field is applicable to all modes.
func (s *soraVideoSchema) FieldModes(name string) []xai.VideoGenMode {
	switch name {
	case ParamInputReference:
		return []xai.VideoGenMode{xai.VideoGenModeImage}
	default:
		return nil
	}
}
