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
	"time"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/spec/util"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

func (p *Service) Actions(model xai.Model) []xai.Action {
	if p.backend == nil {
		return nil
	}
	return p.backend.Actions(model)
}

func (p *Service) supportsAction(model xai.Model, action xai.Action) bool {
	for _, item := range p.Actions(model) {
		if item == action {
			return true
		}
	}
	return false
}

func (p *Service) Operation(model xai.Model, action xai.Action) (op xai.Operation, err error) {
	if !p.supportsAction(model, action) {
		return nil, xai.ErrNotFound
	}
	switch action {
	case xai.GenVideo:
		op = &genVideo{model: string(model)}
	case xai.GenImage:
		op = &genImage{model: string(model)}
	case xai.EditImage:
		op = &editImage{model: string(model)}
	case xai.RecontextImage:
		op = &recontextImage{model: string(model)}
	case xai.SegmentImage:
		op = &segmentImage{model: string(model)}
	case xai.UpscaleImage:
		op = &upscaleImage{model: string(model)}
	default:
		err = xai.ErrNotFound
	}
	return
}

// -----------------------------------------------------------------------------

// NewSyncResponse wraps results into a synchronous OperationResponse (Done() == true).
func NewSyncResponse(ret xai.Results) xai.OperationResponse {
	return util.NewSimpleResp(ret)
}

// -----------------------------------------------------------------------------

type genVideoResp struct {
	op *genai.GenerateVideosOperation
}

func (p genVideoResp) Done() bool {
	return p.op.Done
}

func (p genVideoResp) Sleep() {
	time.Sleep(15 * time.Second)
}

func (p genVideoResp) Retry(ctx context.Context, svc xai.Service) (xai.OperationResponse, error) {
	op, err := svc.(BackendService).Backend().GetVideosOperation(ctx, p.op, nil)
	if err != nil {
		return nil, err
	}
	return genVideoResp{op}, nil
}

func (p genVideoResp) Results() xai.Results {
	ret := p.op.Response
	return util.NewVideoResults[*genai.GeneratedVideo, adapter](ret, ret.GeneratedVideos)
}

func (p genVideoResp) TaskID() string {
	if p.op != nil && p.op.Name != "" {
		return p.op.Name
	}
	return ""
}

type genVideo struct {
	genai.GenerateVideosSource
	genai.GenerateVideosConfig

	model string
}

func (p *genVideo) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_genVideo)
}

func (p *genVideo) Params() xai.Params {
	return NewParams(p)
}

func (p *genVideo) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	op, err := svc.(BackendService).Backend().GenerateVideosFromSource(ctx, p.model, &p.GenerateVideosSource, &p.GenerateVideosConfig)
	if err != nil {
		return
	}
	return genVideoResp{op}, nil
}

// -----------------------------------------------------------------------------

type genImage struct {
	Prompt string
	genai.GenerateImagesConfig

	model string
}

func (p *genImage) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_genImage)
}

func (p *genImage) Params() xai.Params {
	return NewParams(p)
}

func (p *genImage) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	op, err := svc.(BackendService).Backend().GenerateImages(ctx, p.model, p.Prompt, &p.GenerateImagesConfig)
	if err != nil {
		return
	}
	return util.NewImageResultsResp[*genai.GeneratedImage, adapter](op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type editImage struct {
	Prompt     string
	References []genai.ReferenceImage
	genai.EditImageConfig

	model string
}

func (p *editImage) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_editImage)
}

func (p *editImage) Params() xai.Params {
	return NewParams(p)
}

func (p *editImage) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	op, err := svc.(BackendService).Backend().EditImage(ctx, p.model, p.Prompt, p.References, &p.EditImageConfig)
	if err != nil {
		return
	}
	return util.NewImageResultsResp[*genai.GeneratedImage, adapter](op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type recontextImage struct {
	genai.RecontextImageSource
	genai.RecontextImageConfig

	model string
}

func (p *recontextImage) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_recontextImage)
}

func (p *recontextImage) Params() xai.Params {
	return NewParams(p)
}

func (p *recontextImage) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	op, err := svc.(BackendService).Backend().RecontextImage(ctx, p.model, &p.RecontextImageSource, &p.RecontextImageConfig)
	if err != nil {
		return
	}
	return util.NewImageResultsResp[*genai.GeneratedImage, adapter](op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type upscaleImage struct {
	Image  *genai.Image
	Factor string // upscale factor
	genai.UpscaleImageConfig

	model string
}

func (p *upscaleImage) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_upscaleImage)
}

func (p *upscaleImage) Params() xai.Params {
	return NewParams(p)
}

func (p *upscaleImage) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	op, err := svc.(BackendService).Backend().UpscaleImage(ctx, p.model, p.Image, p.Factor, &p.UpscaleImageConfig)
	if err != nil {
		return
	}
	return util.NewImageResultsResp[*genai.GeneratedImage, adapter](op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type segmentImage struct {
	genai.SegmentImageSource
	genai.SegmentImageConfig

	model string
}

func (p *segmentImage) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_segmentImage)
}

func (p *segmentImage) Params() xai.Params {
	return NewParams(p)
}

func (p *segmentImage) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	op, err := svc.(BackendService).Backend().SegmentImage(ctx, p.model, &p.SegmentImageSource, &p.SegmentImageConfig)
	if err != nil {
		return
	}
	return util.NewImageMaskResultsResp[*genai.GeneratedImageMask, adapter](op, op.GeneratedMasks), nil
}

// -----------------------------------------------------------------------------
