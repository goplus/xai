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
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

func (p *Service) Actions(model xai.Model) []xai.Action {
	return []xai.Action{
		xai.GenVideo,
		xai.GenImage,
		xai.EditImage,
		xai.RecontextImage,
		xai.SegmentImage,
		xai.UpscaleImage,
	}
}

func (p *Service) Operation(model xai.Model, action xai.Action) (op xai.Operation, err error) {
	switch action {
	case xai.GenVideo:
		op = &genVideo{}
	case xai.GenImage:
		op = &genImage{}
	case xai.EditImage:
		op = &editImage{}
	case xai.RecontextImage:
		op = &recontextImage{}
	case xai.SegmentImage:
		op = &segmentImage{}
	case xai.UpscaleImage:
		op = &upscaleImage{}
	default:
		err = xai.ErrNotFound
	}
	return
}

// -----------------------------------------------------------------------------

type simpleResp[T xai.Results] struct {
	ret T
}

func (p simpleResp[T]) Done() bool {
	return true
}

func (p simpleResp[T]) Sleep() {
	panic("unreachable")
}

func (p simpleResp[T]) Retry(ctx context.Context, svc xai.Service) (xai.OperationResponse, error) {
	panic("unreachable")
}

func (p simpleResp[T]) Results() xai.Results {
	return p.ret
}

type genImageResp = simpleResp[*outputImages]
type genImageMaskResp = simpleResp[*outputImageMasks]

func newGenImageResp(ret any, items []*genai.GeneratedImage) genImageResp {
	return genImageResp{ret: &outputImages{results(ret), items}}
}

func newGenImageMaskResp(ret any, items []*genai.GeneratedImageMask) genImageMaskResp {
	return genImageMaskResp{ret: &outputImageMasks{results(ret), items}}
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
	op, err := svc.(*Service).ops.GetVideosOperation(ctx, p.op, nil)
	if err != nil {
		return nil, err
	}
	return genVideoResp{op}, nil
}

func (p genVideoResp) Results() xai.Results {
	ret := p.op.Response
	return &outputVideos{results(ret), ret.GeneratedVideos}
}

type genVideo struct {
	genai.GenerateVideosSource
	genai.GenerateVideosConfig

	model string
}

func (p *genVideo) InputSchema() xai.InputSchema {
	return newInputSchema(p)
}

func (p *genVideo) Params() xai.Params {
	return newParams(p)
}

func (p *genVideo) Call(ctx context.Context, svc xai.Service, prompt string, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	p.Prompt = prompt
	op, err := svc.(*Service).models.GenerateVideosFromSource(ctx, p.model, &p.GenerateVideosSource, &p.GenerateVideosConfig)
	if err != nil {
		return
	}
	return genVideoResp{op}, nil
}

// -----------------------------------------------------------------------------

type genImage struct {
	genai.GenerateImagesConfig

	model string
}

func (p *genImage) InputSchema() xai.InputSchema {
	return newInputSchema(p)
}

func (p *genImage) Params() xai.Params {
	return newParams(p)
}

func (p *genImage) Call(ctx context.Context, svc xai.Service, prompt string, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	op, err := svc.(*Service).models.GenerateImages(ctx, p.model, prompt, &p.GenerateImagesConfig)
	if err != nil {
		return
	}
	return newGenImageResp(op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type editImage struct {
	References []genai.ReferenceImage
	genai.EditImageConfig

	model string
}

func (p *editImage) InputSchema() xai.InputSchema {
	return newInputSchema(p)
}

func (p *editImage) Params() xai.Params {
	return newParams(p)
}

func (p *editImage) Call(ctx context.Context, svc xai.Service, prompt string, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	op, err := svc.(*Service).models.EditImage(ctx, p.model, prompt, p.References, &p.EditImageConfig)
	if err != nil {
		return
	}
	return newGenImageResp(op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type recontextImage struct {
	genai.RecontextImageSource
	genai.RecontextImageConfig

	model string
}

func (p *recontextImage) InputSchema() xai.InputSchema {
	return newInputSchema(p)
}

func (p *recontextImage) Params() xai.Params {
	return newParams(p)
}

func (p *recontextImage) Call(ctx context.Context, svc xai.Service, prompt string, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	p.Prompt = prompt
	op, err := svc.(*Service).models.RecontextImage(ctx, p.model, &p.RecontextImageSource, &p.RecontextImageConfig)
	if err != nil {
		return
	}
	return newGenImageResp(op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type upscaleImage struct {
	Image  *genai.Image
	Factor string // upscale factor
	genai.UpscaleImageConfig

	model string
}

func (p *upscaleImage) InputSchema() xai.InputSchema {
	return newInputSchema(p)
}

func (p *upscaleImage) Params() xai.Params {
	return newParams(p)
}

func (p *upscaleImage) Call(ctx context.Context, svc xai.Service, prompt string, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	op, err := svc.(*Service).models.UpscaleImage(ctx, p.model, p.Image, p.Factor, &p.UpscaleImageConfig)
	if err != nil {
		return
	}
	return newGenImageResp(op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type segmentImage struct {
	genai.SegmentImageSource
	genai.SegmentImageConfig

	model string
}

func (p *segmentImage) InputSchema() xai.InputSchema {
	return newInputSchema(p)
}

func (p *segmentImage) Params() xai.Params {
	return newParams(p)
}

func (p *segmentImage) Call(ctx context.Context, svc xai.Service, prompt string, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	if v, ok := opts.(*options); ok {
		p.HTTPOptions = &v.opts
	}
	p.Prompt = prompt
	op, err := svc.(*Service).models.SegmentImage(ctx, p.model, &p.SegmentImageSource, &p.SegmentImageConfig)
	if err != nil {
		return
	}
	return newGenImageMaskResp(op, op.GeneratedMasks), nil
}

// -----------------------------------------------------------------------------
