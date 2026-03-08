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

package gemini

import (
	"context"
	"iter"

	xai "github.com/goplus/xai/spec"
	"google.golang.org/genai"
)

// BackendService is implemented by services that have a gemini Backend.
type BackendService interface {
	xai.Service
	Backend() Backend
}

// Backend defines the transport/backend capabilities needed by spec/gemini.
// Implementations can be based on google.golang.org/genai, OpenAI-compatible
// gateways, or any vendor-specific protocol.
type Backend interface {
	Actions(model xai.Model) []xai.Action

	GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error)
	GenerateContentStream(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) iter.Seq2[*genai.GenerateContentResponse, error]

	GenerateVideosFromSource(ctx context.Context, model string, source *genai.GenerateVideosSource, config *genai.GenerateVideosConfig) (*genai.GenerateVideosOperation, error)
	GetVideosOperation(ctx context.Context, op *genai.GenerateVideosOperation, config *genai.GetOperationConfig) (*genai.GenerateVideosOperation, error)

	GenerateImages(ctx context.Context, model string, prompt string, config *genai.GenerateImagesConfig) (*genai.GenerateImagesResponse, error)
	EditImage(ctx context.Context, model string, prompt string, references []genai.ReferenceImage, config *genai.EditImageConfig) (*genai.EditImageResponse, error)
	RecontextImage(ctx context.Context, model string, source *genai.RecontextImageSource, config *genai.RecontextImageConfig) (*genai.RecontextImageResponse, error)
	UpscaleImage(ctx context.Context, model string, image *genai.Image, factor string, config *genai.UpscaleImageConfig) (*genai.UpscaleImageResponse, error)
	SegmentImage(ctx context.Context, model string, source *genai.SegmentImageSource, config *genai.SegmentImageConfig) (*genai.SegmentImageResponse, error)
}

// -----------------------------------------------------------------------------

type genAIBackend struct {
	models genai.Models
	ops    genai.Operations
}

func newGenAIBackend(models genai.Models, ops genai.Operations) Backend {
	return &genAIBackend{models: models, ops: ops}
}

func (p *genAIBackend) Actions(model xai.Model) []xai.Action {
	return []xai.Action{
		xai.GenVideo,
		xai.GenImage,
		xai.EditImage,
		xai.RecontextImage,
		xai.SegmentImage,
		xai.UpscaleImage,
	}
}

func (p *genAIBackend) GenerateContent(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) (*genai.GenerateContentResponse, error) {
	return p.models.GenerateContent(ctx, model, contents, config)
}

func (p *genAIBackend) GenerateContentStream(ctx context.Context, model string, contents []*genai.Content, config *genai.GenerateContentConfig) iter.Seq2[*genai.GenerateContentResponse, error] {
	return p.models.GenerateContentStream(ctx, model, contents, config)
}

func (p *genAIBackend) GenerateVideosFromSource(ctx context.Context, model string, source *genai.GenerateVideosSource, config *genai.GenerateVideosConfig) (*genai.GenerateVideosOperation, error) {
	return p.models.GenerateVideosFromSource(ctx, model, source, config)
}

func (p *genAIBackend) GetVideosOperation(ctx context.Context, op *genai.GenerateVideosOperation, config *genai.GetOperationConfig) (*genai.GenerateVideosOperation, error) {
	return p.ops.GetVideosOperation(ctx, op, config)
}

func (p *genAIBackend) GenerateImages(ctx context.Context, model string, prompt string, config *genai.GenerateImagesConfig) (*genai.GenerateImagesResponse, error) {
	return p.models.GenerateImages(ctx, model, prompt, config)
}

func (p *genAIBackend) EditImage(ctx context.Context, model string, prompt string, references []genai.ReferenceImage, config *genai.EditImageConfig) (*genai.EditImageResponse, error) {
	return p.models.EditImage(ctx, model, prompt, references, config)
}

func (p *genAIBackend) RecontextImage(ctx context.Context, model string, source *genai.RecontextImageSource, config *genai.RecontextImageConfig) (*genai.RecontextImageResponse, error) {
	return p.models.RecontextImage(ctx, model, source, config)
}

func (p *genAIBackend) UpscaleImage(ctx context.Context, model string, image *genai.Image, factor string, config *genai.UpscaleImageConfig) (*genai.UpscaleImageResponse, error) {
	return p.models.UpscaleImage(ctx, model, image, factor, config)
}

func (p *genAIBackend) SegmentImage(ctx context.Context, model string, source *genai.SegmentImageSource, config *genai.SegmentImageConfig) (*genai.SegmentImageResponse, error) {
	return p.models.SegmentImage(ctx, model, source, config)
}

// -----------------------------------------------------------------------------
