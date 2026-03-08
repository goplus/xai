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

package kling

import (
	"context"
	"errors"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/kling/image"
	"github.com/goplus/xai/spec/kling/internal"
	"github.com/goplus/xai/spec/kling/video"
)

// ImageParams is the typed params interface for image generation.
// Re-exported from image package for Backend implementations.
type ImageParams = image.ImageParams

// VideoParams is the typed params interface for video generation.
// Re-exported from video package for Backend implementations.
type VideoParams = video.VideoParams

// Scheme is the URI scheme for Kling: "kling".
const Scheme = "kling"

// -----------------------------------------------------------------------------
// Model constants (re-exported from internal)
// -----------------------------------------------------------------------------

const (
	ModelKlingV1      = internal.ModelKlingV1
	ModelKlingV15     = internal.ModelKlingV15
	ModelKlingV2      = internal.ModelKlingV2
	ModelKlingV2New   = internal.ModelKlingV2New
	ModelKlingV21     = internal.ModelKlingV21
	ModelKlingImageO1 = internal.ModelKlingImageO1
	ModelKlingV21Video = internal.ModelKlingV21Video
	ModelKlingV25Turbo = internal.ModelKlingV25Turbo
	ModelKlingVideoO1  = internal.ModelKlingVideoO1
	ModelKlingV26      = internal.ModelKlingV26
	ModelKlingV27      = internal.ModelKlingV27
	ModelKlingV28      = internal.ModelKlingV28
	ModelKlingV29      = internal.ModelKlingV29
	ModelKlingV3       = internal.ModelKlingV3
	ModelKlingV3Omni   = internal.ModelKlingV3Omni
)

// ErrPromptRequired is returned when prompt is not set or empty.
var ErrPromptRequired = internal.ErrPromptRequired

// BuildImageParams builds typed ImageParams from Params for the given model.
func BuildImageParams(model string, p *Params) (image.ImageParams, error) {
	if p == nil {
		return nil, errors.New("kling: params is nil")
	}
	return image.BuildImageParams(model, p)
}

// BuildVideoParams builds typed VideoParams from Params for the given model.
func BuildVideoParams(model string, p *Params) (video.VideoParams, error) {
	if p == nil {
		return nil, errors.New("kling: params is nil")
	}
	return video.BuildVideoParams(model, p)
}

// IsImageModel returns true if the model supports image generation.
func IsImageModel(m string) bool { return image.IsImageModel(m) }

// IsVideoModel returns true if the model supports video generation.
func IsVideoModel(m string) bool { return video.IsVideoModel(m) }

// ImageModels returns all supported image model IDs.
func ImageModels() []string { return image.ImageModels() }

// VideoModels returns all supported video model IDs.
func VideoModels() []string { return video.VideoModels() }

// SchemaForImage returns the InputSchema fields for the given image model.
func SchemaForImage(model string) []xai.Field { return image.SchemaForImage(model) }

// SchemaForVideo returns the InputSchema fields for the given video model.
func SchemaForVideo(model string) []xai.Field { return video.SchemaForVideo(model) }

// Register registers the Kling service with xai. The application creates the
// service with NewService (injecting Executors) and passes it here.
//
// Example:
//
//	svc := kling.NewService(imgExec, vidExec)
//	kling.Register(svc)
//	// Then xai.New(ctx, "kling://") returns svc
func Register(svc *Service) {
	xai.Register(Scheme, func(ctx context.Context, uri string) (xai.Service, error) {
		return svc, nil
	})
}
