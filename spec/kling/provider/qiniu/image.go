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

package qiniu

import (
	"fmt"

	"github.com/goplus/xai/spec/kling"
	"github.com/goplus/xai/spec/kling/image"
)

// Image API endpoints.
const (
	EndpointImageGenerations = "/v1/images/generations"
	EndpointImageEdits       = "/v1/images/edits"
	EndpointImageO1          = "/queue/fal-ai/kling-image/o1"
	EndpointImageTaskStatus  = "/v1/images/tasks/"
	EndpointImageO1Status    = "/queue/fal-ai/kling-image/requests/"
)

// ImageRequest holds the endpoint and body for an image API request.
type ImageRequest struct {
	Endpoint string
	Body     map[string]any
	IsO1     bool
}

// BuildImageRequest builds an ImageRequest from typed ImageParams.
func BuildImageRequest(model string, params kling.ImageParams) (*ImageRequest, error) {
	switch p := params.(type) {
	case *image.V1ImageParams:
		return buildV1ImageRequest(p), nil
	case *image.V15ImageParams:
		return buildV15ImageRequest(p), nil
	case *image.V2ImageParams:
		return buildV2ImageRequest(p), nil
	case *image.V21ImageParams:
		return buildV21ImageRequest(p), nil
	case *image.O1ImageParams:
		return buildO1ImageRequest(p), nil
	case *image.GeminiImageParams:
		return buildGeminiImageRequest(p), nil
	default:
		return nil, fmt.Errorf("qiniu: unsupported image params type: %T", params)
	}
}

func buildV1ImageRequest(p *image.V1ImageParams) *ImageRequest {
	body := map[string]any{
		"model":  p.Model(),
		"prompt": p.Prompt,
	}
	setOptionalInt(body, "n", p.N)
	setOptionalString(body, "image", p.Image)
	setOptionalFloat64(body, "image_fidelity", p.ImageFidelity)
	setOptionalFloat64(body, "human_fidelity", p.HumanFidelity)
	setOptionalString(body, "aspect_ratio", p.AspectRatio)
	setOptionalString(body, "negative_prompt", p.NegativePrompt)

	return &ImageRequest{
		Endpoint: EndpointImageGenerations,
		Body:     body,
	}
}

func buildV15ImageRequest(p *image.V15ImageParams) *ImageRequest {
	body := map[string]any{
		"model":  p.Model(),
		"prompt": p.Prompt,
	}
	setOptionalInt(body, "n", p.N)
	setOptionalString(body, "image", p.Image)
	setOptionalString(body, "image_reference", p.ImageReference)
	setOptionalFloat64(body, "image_fidelity", p.ImageFidelity)
	setOptionalFloat64(body, "human_fidelity", p.HumanFidelity)
	setOptionalString(body, "aspect_ratio", p.AspectRatio)
	setOptionalString(body, "negative_prompt", p.NegativePrompt)

	return &ImageRequest{
		Endpoint: EndpointImageGenerations,
		Body:     body,
	}
}

func buildV2ImageRequest(p *image.V2ImageParams) *ImageRequest {
	body := map[string]any{
		"model":  p.Model(),
		"prompt": p.Prompt,
	}
	setOptionalInt(body, "n", p.N)
	setOptionalString(body, "aspect_ratio", p.AspectRatio)
	setOptionalString(body, "negative_prompt", p.NegativePrompt)

	// Multi-image edit mode: 2-4 subject images
	if len(p.SubjectImageList) >= 2 {
		body["image"] = ""
		subjectList := make([]map[string]string, len(p.SubjectImageList))
		for i, item := range p.SubjectImageList {
			subjectList[i] = map[string]string{"subject_image": item.SubjectImage}
		}
		body["subject_image_list"] = subjectList
		setOptionalString(body, "scene_image", p.SceneImage)
		setOptionalString(body, "style_image", p.StyleImage)

		return &ImageRequest{
			Endpoint: EndpointImageEdits,
			Body:     body,
		}
	}

	// Single image mode
	setOptionalString(body, "image", p.Image)
	if len(p.ReferenceImages) > 0 {
		body["reference_images"] = p.ReferenceImages
	}
	return &ImageRequest{
		Endpoint: EndpointImageGenerations,
		Body:     body,
	}
}

func buildV21ImageRequest(p *image.V21ImageParams) *ImageRequest {
	body := map[string]any{
		"model":  p.Model(),
		"prompt": p.Prompt,
	}
	setOptionalInt(body, "n", p.N)
	setOptionalString(body, "aspect_ratio", p.AspectRatio)
	setOptionalString(body, "negative_prompt", p.NegativePrompt)

	// Multi-image edit mode
	if len(p.SubjectImageList) >= 2 {
		body["image"] = ""
		subjectList := make([]map[string]string, len(p.SubjectImageList))
		for i, item := range p.SubjectImageList {
			subjectList[i] = map[string]string{"subject_image": item.SubjectImage}
		}
		body["subject_image_list"] = subjectList
		setOptionalString(body, "scene_image", p.SceneImage)
		setOptionalString(body, "style_image", p.StyleImage)

		return &ImageRequest{
			Endpoint: EndpointImageEdits,
			Body:     body,
		}
	}

	// Single image mode
	setOptionalString(body, "image", p.Image)
	if len(p.ReferenceImages) > 0 {
		body["reference_images"] = p.ReferenceImages
	}
	return &ImageRequest{
		Endpoint: EndpointImageGenerations,
		Body:     body,
	}
}

func buildO1ImageRequest(p *image.O1ImageParams) *ImageRequest {
	body := map[string]any{
		"prompt": p.Prompt,
	}
	if p.N > 0 {
		body["num_images"] = p.N
	}
	setOptionalString(body, "resolution", p.Resolution)
	setOptionalString(body, "aspect_ratio", p.AspectRatio)
	if len(p.ReferenceImages) > 0 {
		body["image_urls"] = p.ReferenceImages
	}

	return &ImageRequest{
		Endpoint: EndpointImageO1,
		Body:     body,
		IsO1:     true,
	}
}

func buildGeminiImageRequest(p *image.GeminiImageParams) *ImageRequest {
	body := map[string]any{
		"model":  p.Model(),
		"prompt": p.Prompt,
	}
	setOptionalInt(body, "n", p.N)
	setOptionalString(body, "size", p.Size)
	setOptionalString(body, "aspect_ratio", p.AspectRatio)
	setOptionalString(body, "image", p.Image)
	if len(p.ReferenceImages) > 0 {
		body["reference_images"] = p.ReferenceImages
	}

	return &ImageRequest{
		Endpoint: EndpointImageGenerations,
		Body:     body,
	}
}

// GetImageTaskStatusEndpoint returns the endpoint for querying image task status.
func GetImageTaskStatusEndpoint(taskID string, isO1 bool) string {
	if isO1 {
		return EndpointImageO1Status + taskID + "/status"
	}
	return EndpointImageTaskStatus + taskID
}

// Helper functions for building request body.

func setOptionalString(body map[string]any, key, value string) {
	if value != "" {
		body[key] = value
	}
}

func setOptionalInt(body map[string]any, key string, value *int) {
	if value != nil && *value > 0 {
		body[key] = *value
	}
}

func setOptionalFloat64(body map[string]any, key string, value *float64) {
	if value != nil {
		body[key] = *value
	}
}

func setOptionalBool(body map[string]any, key string, value bool) {
	if value {
		body[key] = value
	}
}
