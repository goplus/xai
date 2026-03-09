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

import "github.com/goplus/xai/spec/kling/internal"

// ImageParams is the unified interface for all image model parameters.
type ImageParams interface {
	Model() string
	imageParams()
}

type imageParamsMarker struct{}

func (imageParamsMarker) imageParams() {}

// V1ImageParams for kling-v1.
type V1ImageParams struct {
	imageParamsMarker
	Prompt         string   `json:"prompt"`
	N              *int     `json:"n,omitempty"`
	Image          string   `json:"image,omitempty"`
	ImageFidelity  *float64 `json:"image_fidelity,omitempty"`
	HumanFidelity  *float64 `json:"human_fidelity,omitempty"`
	AspectRatio    string   `json:"aspect_ratio,omitempty"`
	NegativePrompt string   `json:"negative_prompt,omitempty"`
}

func (p *V1ImageParams) Model() string { return internal.ModelKlingV1 }

// V15ImageParams for kling-v1-5.
type V15ImageParams struct {
	imageParamsMarker
	Prompt         string   `json:"prompt"`
	N              *int     `json:"n,omitempty"`
	Image          string   `json:"image,omitempty"`
	ImageReference string   `json:"image_reference,omitempty"`
	ImageFidelity  *float64 `json:"image_fidelity,omitempty"`
	HumanFidelity  *float64 `json:"human_fidelity,omitempty"`
	AspectRatio    string   `json:"aspect_ratio,omitempty"`
	NegativePrompt string   `json:"negative_prompt,omitempty"`
}

func (p *V15ImageParams) Model() string { return internal.ModelKlingV15 }

// SubjectImageItem holds a single subject image URL for multi-image edits.
type SubjectImageItem struct {
	SubjectImage string `json:"subject_image"`
}

// V2ImageParams for kling-v2 and kling-v2-new.
// kling-v2: supports multi-image (subject_image_list, scene_image, style_image) via /edits.
// kling-v2-new: single image only (img2img).
type V2ImageParams struct {
	imageParamsMarker
	ModelName        string             `json:"model"` // kling-v2 or kling-v2-new
	Prompt           string             `json:"prompt"`
	N                *int               `json:"n,omitempty"`
	Image            string             `json:"image,omitempty"`
	AspectRatio      string             `json:"aspect_ratio,omitempty"`
	NegativePrompt   string             `json:"negative_prompt,omitempty"`
	// Multi-image edit (kling-v2 only, 2-4 images)
	SubjectImageList []SubjectImageItem `json:"subject_image_list,omitempty"`
	SceneImage       string             `json:"scene_image,omitempty"`
	StyleImage       string             `json:"style_image,omitempty"`
}

func (p *V2ImageParams) Model() string { return p.ModelName }

// V21ImageParams for kling-v2-1.
// Supports single image (generations) or multi-image (edits via subject_image_list).
type V21ImageParams struct {
	imageParamsMarker
	Prompt          string             `json:"prompt"`
	N               *int               `json:"n,omitempty"`
	Image           string             `json:"image,omitempty"`
	ReferenceImages []string           `json:"reference_images,omitempty"`
	SubjectImageList []SubjectImageItem `json:"subject_image_list,omitempty"`
	SceneImage      string             `json:"scene_image,omitempty"`
	StyleImage      string             `json:"style_image,omitempty"`
	AspectRatio     string             `json:"aspect_ratio,omitempty"`
	NegativePrompt  string             `json:"negative_prompt,omitempty"`
}

func (p *V21ImageParams) Model() string { return internal.ModelKlingV21 }

// O1ImageParams for kling-image-o1.
type O1ImageParams struct {
	imageParamsMarker
	Prompt         string   `json:"prompt"`
	N              int      `json:"n,omitempty"`
	Resolution     string   `json:"resolution,omitempty"`
	AspectRatio    string   `json:"aspect_ratio,omitempty"`
	ReferenceImages []string `json:"reference_images,omitempty"`
}

func (p *O1ImageParams) Model() string { return internal.ModelKlingImageO1 }

// GeminiImageParams for gemini image models (e.g. via qiniu).
type GeminiImageParams struct {
	imageParamsMarker
	ModelName      string   `json:"model"`
	Prompt         string   `json:"prompt"`
	N              *int     `json:"n,omitempty"`
	Size           string   `json:"size,omitempty"`
	AspectRatio    string   `json:"aspect_ratio,omitempty"`
	Image          string   `json:"image,omitempty"`
	ReferenceImages []string `json:"reference_images,omitempty"`
}

func (p *GeminiImageParams) Model() string { return p.ModelName }
