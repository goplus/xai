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

package image

import (
	"fmt"
	"strings"

	"github.com/goplus/xai/spec/kling/internal"
)

// BuildImageParams builds typed ImageParams from ParamsReader for the given model.
// Routing: 0-1 refs → /generations; 2-4 refs (subject_image_list) → /edits. See kling_image.md.
func BuildImageParams(model string, p internal.ParamsReader) (ImageParams, error) {
	m := strings.ToLower(model)
	prompt := p.GetString(internal.ParamPrompt)

	switch {
	case m == internal.ModelKlingV1:
		return buildV1ImageParams(model, prompt, p), nil
	case m == internal.ModelKlingV15:
		return buildV15ImageParams(model, prompt, p), nil
	case m == internal.ModelKlingV2 || m == internal.ModelKlingV2New:
		return buildV2ImageParams(model, prompt, p)
	case m == internal.ModelKlingV21:
		return buildV21ImageParams(model, prompt, p), nil
	case m == internal.ModelKlingImageO1:
		return buildO1ImageParams(model, prompt, p), nil
	case strings.Contains(m, "gemini"):
		return buildGeminiImageParams(model, prompt, p), nil
	default:
		return nil, fmt.Errorf("kling: unsupported image model %q", model)
	}
}

func buildV1ImageParams(model, prompt string, p internal.ParamsReader) *V1ImageParams {
	n := internal.GetInt(p, internal.ParamN)
	return &V1ImageParams{
		Prompt:         prompt,
		N:              &n,
		Image:          p.GetString(internal.ParamImage),
		ImageFidelity:  internal.GetFloat64Ptr(p, internal.ParamImageFidelity),
		HumanFidelity:  internal.GetFloat64Ptr(p, internal.ParamHumanFidelity),
		AspectRatio:    p.GetString(internal.ParamAspectRatio),
		NegativePrompt: p.GetString(internal.ParamNegativePrompt),
	}
}

func buildV15ImageParams(model, prompt string, p internal.ParamsReader) *V15ImageParams {
	n := internal.GetInt(p, internal.ParamN)
	return &V15ImageParams{
		Prompt:         prompt,
		N:              &n,
		Image:          p.GetString(internal.ParamImage),
		ImageReference: p.GetString(internal.ParamImageReference),
		ImageFidelity:  internal.GetFloat64Ptr(p, internal.ParamImageFidelity),
		HumanFidelity:  internal.GetFloat64Ptr(p, internal.ParamHumanFidelity),
		AspectRatio:    p.GetString(internal.ParamAspectRatio),
		NegativePrompt: p.GetString(internal.ParamNegativePrompt),
	}
}

func buildV2ImageParams(model, prompt string, p internal.ParamsReader) (ImageParams, error) {
	n := internal.GetInt(p, internal.ParamN)
	m := strings.ToLower(model)
	img := p.GetString(internal.ParamImage)
	refs := internal.GetStringSlice(p, internal.ParamReferenceImages)
	subjectList := internal.GetSubjectImageList(p, internal.ParamSubjectImageList)
	sceneImg := p.GetString(internal.ParamSceneImage)
	styleImg := p.GetString(internal.ParamStyleImage)

	// kling-v2: multi-image (2-4) via subject_image_list
	if len(subjectList) >= 2 && len(subjectList) <= 4 {
		if m == internal.ModelKlingV2New {
			return nil, fmt.Errorf("kling: %s does not support multi-image; use single image only", model)
		}
		var subjectImageList []SubjectImageItem
		for _, url := range subjectList {
			subjectImageList = append(subjectImageList, SubjectImageItem{SubjectImage: url})
		}
		return &V2ImageParams{
			ModelName:        model,
			Prompt:           prompt,
			N:                &n,
			Image:            "",
			AspectRatio:      p.GetString(internal.ParamAspectRatio),
			NegativePrompt:   p.GetString(internal.ParamNegativePrompt),
			SubjectImageList: subjectImageList,
			SceneImage:       sceneImg,
			StyleImage:       styleImg,
		}, nil
	}

	// Single image: from Image or first of ReferenceImages
	if img == "" && len(refs) > 0 {
		img = refs[0]
	}
	if len(refs) > 1 {
		return nil, fmt.Errorf("kling: %s single-image mode accepts at most one reference; use subject_image_list for multi-image", model)
	}
	// kling-v2-new supports img2img only (no text2image)
	if m == internal.ModelKlingV2New && img == "" {
		return nil, fmt.Errorf("kling: %s requires a reference image (img2img only)", model)
	}
	return &V2ImageParams{
		ModelName:      model,
		Prompt:         prompt,
		N:              &n,
		Image:          img,
		AspectRatio:    p.GetString(internal.ParamAspectRatio),
		NegativePrompt: p.GetString(internal.ParamNegativePrompt),
	}, nil
}

func buildV21ImageParams(model, prompt string, p internal.ParamsReader) *V21ImageParams {
	n := internal.GetInt(p, internal.ParamN)
	img := p.GetString(internal.ParamImage)
	refs := internal.GetStringSlice(p, internal.ParamReferenceImages)
	subjectList := internal.GetSubjectImageList(p, internal.ParamSubjectImageList)
	sceneImg := p.GetString(internal.ParamSceneImage)
	styleImg := p.GetString(internal.ParamStyleImage)

	var subjectImageList []SubjectImageItem
	var refImages []string
	if len(subjectList) >= 2 && len(subjectList) <= 4 {
		for _, url := range subjectList {
			subjectImageList = append(subjectImageList, SubjectImageItem{SubjectImage: url})
		}
		refImages = subjectList // for provider to pass to canvas /edits
	} else {
		// Single image: from Image or first of ReferenceImages
		if img == "" && len(refs) > 0 {
			img = refs[0]
		}
		refImages = refs
	}

	return &V21ImageParams{
		Prompt:          prompt,
		N:               &n,
		Image:           img,
		ReferenceImages: refImages,
		SubjectImageList: subjectImageList,
		SceneImage:       sceneImg,
		StyleImage:       styleImg,
		AspectRatio:      p.GetString(internal.ParamAspectRatio),
		NegativePrompt:   p.GetString(internal.ParamNegativePrompt),
	}
}

func buildO1ImageParams(model, prompt string, p internal.ParamsReader) *O1ImageParams {
	n := internal.GetInt(p, internal.ParamN)
	if n < 1 {
		n = 1
	}
	return &O1ImageParams{
		Prompt:          prompt,
		N:               n,
		Resolution:      p.GetString(internal.ParamResolution),
		AspectRatio:     p.GetString(internal.ParamAspectRatio),
		ReferenceImages: internal.GetStringSlice(p, internal.ParamReferenceImages),
	}
}

func buildGeminiImageParams(model, prompt string, p internal.ParamsReader) *GeminiImageParams {
	n := internal.GetInt(p, internal.ParamN)
	return &GeminiImageParams{
		ModelName:       model,
		Prompt:          prompt,
		N:               &n,
		Size:            p.GetString(internal.ParamSize),
		AspectRatio:     p.GetString(internal.ParamAspectRatio),
		Image:           p.GetString(internal.ParamImage),
		ReferenceImages: internal.GetStringSlice(p, internal.ParamReferenceImages),
	}
}
