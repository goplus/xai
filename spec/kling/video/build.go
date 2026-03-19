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
 * WITHOUT WARRANTIES OR CONDITIONS OF KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package video

import (
	"fmt"
	"strings"

	"github.com/goplus/xai/spec/kling/internal"
)

// BuildVideoParams builds typed VideoParams from ParamsReader for the given model.
func BuildVideoParams(model string, p internal.ParamsReader) (VideoParams, error) {
	m := strings.ToLower(model)
	prompt := p.GetString(internal.ParamPrompt)

	switch {
	case m == internal.ModelKlingV21Video || m == internal.ModelKlingV25Turbo:
		return buildBaseVideoParams(model, prompt, p), nil
	case m == internal.ModelKlingVideoO1:
		return buildO1VideoParams(model, prompt, p), nil
	case m == internal.ModelKlingV26 || m == internal.ModelKlingV27 ||
		m == internal.ModelKlingV28 || m == internal.ModelKlingV29:
		return buildV26VideoParams(model, prompt, p), nil
	case m == internal.ModelKlingV3:
		return buildV3VideoParams(model, prompt, p), nil
	case m == internal.ModelKlingV3Omni:
		return buildV3OmniVideoParams(model, prompt, p), nil
	default:
		return nil, fmt.Errorf("kling: unsupported video model %q", model)
	}
}

func buildBaseVideoParams(model, prompt string, p internal.ParamsReader) *BaseVideoParams {
	return &BaseVideoParams{
		ModelName:      model,
		Prompt:         prompt,
		InputReference: p.GetString(internal.ParamInputReference),
		ImageTail:      p.GetString(internal.ParamImageTail),
		NegativePrompt: p.GetString(internal.ParamNegativePrompt),
		Mode:           p.GetString(internal.ParamMode),
		Seconds:        p.GetString(internal.ParamSeconds),
		Size:           p.GetString(internal.ParamSize),
	}
}

func buildO1VideoParams(model, prompt string, p internal.ParamsReader) *O1VideoParams {
	imageList := getImageList(p)
	inputRef := p.GetString(internal.ParamInputReference)
	imageTail := p.GetString(internal.ParamImageTail)
	imageList = mergeLegacyO1Frames(imageList, inputRef, imageTail)
	return &O1VideoParams{
		Prompt:         prompt,
		ImageList:      imageList,
		VideoList:      getVideoList(p),
		NegativePrompt: p.GetString(internal.ParamNegativePrompt),
		Mode:           p.GetString(internal.ParamMode),
		Seconds:        p.GetString(internal.ParamSeconds),
		Size:           p.GetString(internal.ParamSize),
	}
}

func mergeLegacyO1Frames(imageList []ImageInput, inputRef, imageTail string) []ImageInput {
	hasFirstFrame := false
	hasEndFrame := false
	for _, item := range imageList {
		switch item.Type {
		case ImageTypeFirstFrame:
			hasFirstFrame = true
		case ImageTypeEndFrame:
			hasEndFrame = true
		}
	}
	if inputRef != "" && !hasFirstFrame {
		imageList = append([]ImageInput{{Image: inputRef, Type: ImageTypeFirstFrame}}, imageList...)
	}
	if imageTail != "" && !hasEndFrame {
		imageList = append(imageList, ImageInput{Image: imageTail, Type: ImageTypeEndFrame})
	}
	return imageList
}

func buildV3VideoParams(model, prompt string, p internal.ParamsReader) *V3VideoParams {
	return &V3VideoParams{
		ModelName:      model,
		Prompt:         prompt,
		InputReference: p.GetString(internal.ParamInputReference),
		MultiShot:      internal.GetBool(p, internal.ParamMultiShot),
		ShotType:       p.GetString(internal.ParamShotType),
		MultiPrompt:    getMultiPrompt(p),
		Sound:          p.GetString(internal.ParamSound),
		Mode:           p.GetString(internal.ParamMode),
		Seconds:        p.GetString(internal.ParamSeconds),
		Size:           p.GetString(internal.ParamSize),
	}
}

func buildV3OmniVideoParams(model, prompt string, p internal.ParamsReader) *V3OmniVideoParams {
	return &V3OmniVideoParams{
		ModelName:   model,
		Prompt:      prompt,
		MultiShot:   internal.GetBool(p, internal.ParamMultiShot),
		ShotType:    p.GetString(internal.ParamShotType),
		MultiPrompt: getMultiPrompt(p),
		ImageList:   getImageList(p),
		VideoList:   getVideoList(p),
		Sound:       p.GetString(internal.ParamSound),
		Mode:        p.GetString(internal.ParamMode),
		Seconds:     p.GetString(internal.ParamSeconds),
		Size:        p.GetString(internal.ParamSize),
	}
}

func getMultiPrompt(p internal.ParamsReader) []MultiPromptItem {
	v, ok := p.Get(internal.ParamMultiPrompt)
	if !ok {
		return nil
	}
	// Support strongly-typed []MultiPromptItem (recommended)
	if typed, ok := v.([]MultiPromptItem); ok {
		return typed
	}
	// Support []interface{} for backward compatibility
	if slice, ok := v.([]interface{}); ok {
		return parseMultiPromptFromSlice(slice)
	}
	// Support []map[string]interface{} for backward compatibility
	if maps, ok := v.([]map[string]interface{}); ok {
		return parseMultiPromptFromMaps(maps)
	}
	return nil
}

func parseMultiPromptFromSlice(slice []interface{}) []MultiPromptItem {
	var out []MultiPromptItem
	for _, item := range slice {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		out = append(out, parseMultiPromptItemFromMap(m))
	}
	return out
}

func parseMultiPromptFromMaps(maps []map[string]interface{}) []MultiPromptItem {
	var out []MultiPromptItem
	for _, m := range maps {
		out = append(out, parseMultiPromptItemFromMap(m))
	}
	return out
}

func parseMultiPromptItemFromMap(m map[string]interface{}) MultiPromptItem {
	mp := MultiPromptItem{}
	if i, ok := m["index"].(int); ok {
		mp.Index = i
	} else if f, ok := m["index"].(float64); ok {
		mp.Index = int(f)
	}
	if s, ok := m["prompt"].(string); ok {
		mp.Prompt = s
	}
	if s, ok := m["duration"].(string); ok {
		mp.Duration = s
	}
	return mp
}

func buildV26VideoParams(model, prompt string, p internal.ParamsReader) *V26VideoParams {
	return &V26VideoParams{
		ModelName:            model,
		Prompt:               prompt,
		InputReference:       p.GetString(internal.ParamInputReference),
		ImageTail:            p.GetString(internal.ParamImageTail),
		NegativePrompt:       p.GetString(internal.ParamNegativePrompt),
		Mode:                 p.GetString(internal.ParamMode),
		Seconds:              p.GetString(internal.ParamSeconds),
		Size:                 p.GetString(internal.ParamSize),
		Sound:                p.GetString(internal.ParamSound),
		ImageURL:             p.GetString(internal.ParamImageUrl),
		VideoURL:             p.GetString(internal.ParamVideoUrl),
		CharacterOrientation: p.GetString(internal.ParamCharacterOrientation),
		KeepOriginalSound:    KeepOriginalSound(p.GetString(internal.ParamKeepOriginalSound)),
	}
}

func getImageList(p internal.ParamsReader) []ImageInput {
	v, ok := p.Get(internal.ParamImageList)
	if !ok {
		return nil
	}
	// Support strongly-typed []ImageInput (recommended)
	if typed, ok := v.([]ImageInput); ok {
		return typed
	}
	// Support []interface{} for backward compatibility
	if slice, ok := v.([]interface{}); ok {
		return parseImageListFromSlice(slice)
	}
	// Support []map[string]interface{} for backward compatibility
	if maps, ok := v.([]map[string]interface{}); ok {
		return parseImageListFromMaps(maps)
	}
	return nil
}

func parseImageListFromSlice(slice []interface{}) []ImageInput {
	var out []ImageInput
	for _, item := range slice {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		img := parseImageInputFromMap(m)
		if img.Image != "" {
			out = append(out, img)
		}
	}
	return out
}

func parseImageListFromMaps(maps []map[string]interface{}) []ImageInput {
	var out []ImageInput
	for _, m := range maps {
		img := parseImageInputFromMap(m)
		if img.Image != "" {
			out = append(out, img)
		}
	}
	return out
}

func parseImageInputFromMap(m map[string]interface{}) ImageInput {
	img := ImageInput{}
	if s, ok := m["image"].(string); ok {
		img.Image = s
	}
	if s, ok := m["type"].(string); ok {
		img.Type = ImageRefType(s)
	}
	return img
}

func getVideoList(p internal.ParamsReader) []VideoRef {
	v, ok := p.Get(internal.ParamVideoList)
	if !ok {
		return nil
	}
	// Support strongly-typed []VideoRef (recommended)
	if typed, ok := v.([]VideoRef); ok {
		return typed
	}
	// Support []interface{} for backward compatibility
	if slice, ok := v.([]interface{}); ok {
		return parseVideoListFromSlice(slice)
	}
	// Support []map[string]interface{} for backward compatibility
	if maps, ok := v.([]map[string]interface{}); ok {
		return parseVideoListFromMaps(maps)
	}
	return nil
}

func parseVideoListFromSlice(slice []interface{}) []VideoRef {
	var out []VideoRef
	for _, item := range slice {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		ref := parseVideoRefFromMap(m)
		if ref.VideoURL != "" {
			out = append(out, ref)
		}
	}
	return out
}

func parseVideoListFromMaps(maps []map[string]interface{}) []VideoRef {
	var out []VideoRef
	for _, m := range maps {
		ref := parseVideoRefFromMap(m)
		if ref.VideoURL != "" {
			out = append(out, ref)
		}
	}
	return out
}

func parseVideoRefFromMap(m map[string]interface{}) VideoRef {
	ref := VideoRef{}
	if s, ok := m["video_url"].(string); ok {
		ref.VideoURL = s
	}
	if s, ok := m["refer_type"].(string); ok {
		ref.ReferType = VideoReferType(s)
	}
	if s, ok := m["keep_original_sound"].(string); ok {
		ref.KeepOriginalSound = KeepOriginalSound(s)
	}
	return ref
}
