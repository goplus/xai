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

	"github.com/goplus/xai"
	"github.com/goplus/xai/util"
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
		op = &genVideo{svc: p, model: string(model)}
	case xai.GenImage:
		op = &genImage{svc: p, model: string(model)}
	case xai.EditImage:
		op = &editImage{svc: p, model: string(model)}
	case xai.RecontextImage:
		op = &recontextImage{svc: p, model: string(model)}
	case xai.SegmentImage:
		op = &segmentImage{svc: p, model: string(model)}
	case xai.UpscaleImage:
		op = &upscaleImage{svc: p, model: string(model)}
	default:
		err = xai.ErrNotFound
	}
	return
}

// -----------------------------------------------------------------------------

type genVideoResp struct {
	op  *genai.GenerateVideosOperation
	gen *genVideo
}

func (p *genVideoResp) Done() bool {
	return p.op.Done
}

func (p *genVideoResp) Results() xai.Results {
	ret := p.op.Response
	return util.NewVideoResults[*genai.GeneratedVideo, adapter](ret, ret.GeneratedVideos)
}

func (p *genVideoResp) WaitParams() xai.WaitParams {
	return newWaitParams(&p.gen.callParams)
}

func (p *genVideoResp) Sleep() {
	time.Sleep(15 * time.Second)
}

func (p *genVideoResp) Retry(wp xai.WaitParams) (*genVideoResp, error) {
	var conf *genai.GetOperationConfig
	gen := p.gen
	params := gen.getWaitParams(wp)
	ctx := params.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if params.opts != nil {
		conf = &genai.GetOperationConfig{
			HTTPOptions: params.opts,
		}
	}
	op, err := gen.svc.ops.GetVideosOperation(ctx, p.op, conf)
	if err != nil {
		return nil, err
	}
	return &genVideoResp{op: op, gen: p.gen}, nil
}

func (p *genVideoResp) Wait(wp xai.WaitParams) (ret xai.Results, err error) {
	var progress func(xai.OperationResponse)
	if wp != nil {
		progress = wp.(*waitParams).progress
	}
	for !p.Done() {
		if progress != nil {
			progress(p)
		}
		p.Sleep()
		p, err = p.Retry(wp)
		if err != nil {
			return
		}
	}
	return p.Results(), nil
}

type genVideo struct {
	callParams
	genai.GenerateVideosSource
	genai.GenerateVideosConfig

	model string
	svc   *Service
}

func (p *genVideo) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_genVideo)
}

func (p *genVideo) CallParams() xai.CallParams {
	return p.initCallParams(p)
}

func (p *genVideo) Call(cp xai.CallParams) (resp xai.OperationResponse, err error) {
	params := cp.(*callParams)
	ctx := params.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if params.opts != nil {
		p.HTTPOptions = params.opts
	}
	op, err := p.svc.models.GenerateVideosFromSource(ctx, p.model, &p.GenerateVideosSource, &p.GenerateVideosConfig)
	if err != nil {
		return
	}
	return &genVideoResp{op: op, gen: p}, nil
}

// -----------------------------------------------------------------------------

type genImage struct {
	callParams
	Prompt string
	genai.GenerateImagesConfig

	model string
	svc   *Service
}

func (p *genImage) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_genImage)
}

func (p *genImage) CallParams() xai.CallParams {
	return p.initCallParams(p)
}

func (p *genImage) Call(params xai.CallParams) (resp xai.OperationResponse, err error) {
	cp := params.(*callParams)
	ctx := cp.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if cp.opts != nil {
		p.HTTPOptions = cp.opts
	}
	op, err := p.svc.models.GenerateImages(ctx, p.model, p.Prompt, &p.GenerateImagesConfig)
	if err != nil {
		return
	}
	return util.NewImageResultsResp[*genai.GeneratedImage, adapter](op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type editImage struct {
	callParams
	Prompt     string
	References []genai.ReferenceImage
	genai.EditImageConfig

	model string
	svc   *Service
}

func (p *editImage) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_editImage)
}

func (p *editImage) CallParams() xai.CallParams {
	return p.initCallParams(p)
}

func (p *editImage) Call(cp xai.CallParams) (resp xai.OperationResponse, err error) {
	params := cp.(*callParams)
	ctx := params.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if params.opts != nil {
		p.HTTPOptions = params.opts
	}
	op, err := p.svc.models.EditImage(ctx, p.model, p.Prompt, p.References, &p.EditImageConfig)
	if err != nil {
		return
	}
	return util.NewImageResultsResp[*genai.GeneratedImage, adapter](op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type recontextImage struct {
	callParams
	genai.RecontextImageSource
	genai.RecontextImageConfig

	model string
	svc   *Service
}

func (p *recontextImage) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_recontextImage)
}

func (p *recontextImage) CallParams() xai.CallParams {
	return p.initCallParams(p)
}

func (p *recontextImage) Call(cp xai.CallParams) (resp xai.OperationResponse, err error) {
	params := cp.(*callParams)
	ctx := params.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if params.opts != nil {
		p.HTTPOptions = params.opts
	}
	op, err := p.svc.models.RecontextImage(ctx, p.model, &p.RecontextImageSource, &p.RecontextImageConfig)
	if err != nil {
		return
	}
	return util.NewImageResultsResp[*genai.GeneratedImage, adapter](op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type upscaleImage struct {
	callParams
	Image  *genai.Image
	Factor string // upscale factor
	genai.UpscaleImageConfig

	model string
	svc   *Service
}

func (p *upscaleImage) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_upscaleImage)
}

func (p *upscaleImage) CallParams() xai.CallParams {
	return p.initCallParams(p)
}

func (p *upscaleImage) Call(cp xai.CallParams) (resp xai.OperationResponse, err error) {
	params := cp.(*callParams)
	ctx := params.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if params.opts != nil {
		p.HTTPOptions = params.opts
	}
	op, err := p.svc.models.UpscaleImage(ctx, p.model, p.Image, p.Factor, &p.UpscaleImageConfig)
	if err != nil {
		return
	}
	return util.NewImageResultsResp[*genai.GeneratedImage, adapter](op, op.GeneratedImages), nil
}

// -----------------------------------------------------------------------------

type segmentImage struct {
	callParams
	genai.SegmentImageSource
	genai.SegmentImageConfig

	model string
	svc   *Service
}

func (p *segmentImage) InputSchema() xai.InputSchema {
	return newInputSchema(p, restriction_segmentImage)
}

func (p *segmentImage) CallParams() xai.CallParams {
	return p.initCallParams(p)
}

func (p *segmentImage) Call(cp xai.CallParams) (resp xai.OperationResponse, err error) {
	params := cp.(*callParams)
	ctx := params.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if params.opts != nil {
		p.HTTPOptions = params.opts
	}
	op, err := p.svc.models.SegmentImage(ctx, p.model, &p.SegmentImageSource, &p.SegmentImageConfig)
	if err != nil {
		return
	}
	return util.NewImageMaskResultsResp[*genai.GeneratedImageMask, adapter](op, op.GeneratedMasks), nil
}

// -----------------------------------------------------------------------------
