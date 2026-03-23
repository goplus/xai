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
	"encoding/base64"
	"io"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/goplus/xai"
	"github.com/goplus/xai/types"
	"github.com/goplus/xai/util"
	"google.golang.org/genai"
)

// -----------------------------------------------------------------------------

type image genai.Image

func (p *image) Type() xai.ImageType {
	return xai.ImageType(p.MIMEType)
}

func (p *image) Blob() xai.BlobData {
	if len(p.ImageBytes) > 0 {
		return xai.BlobFromRaw(p.ImageBytes)
	}
	return nil
}

func (p *image) StgUri() string {
	return p.GCSURI
}

func imageOf(v xai.Image) *genai.Image {
	return (*genai.Image)(v.(*image))
}

// -----------------------------------------------------------------------------

type video genai.Video

func (p *video) Type() xai.VideoType {
	return xai.VideoType(p.MIMEType)
}

func (p *video) Blob() xai.BlobData {
	if len(p.VideoBytes) > 0 {
		return xai.BlobFromRaw(p.VideoBytes)
	}
	return nil
}

func (p *video) StgUri() string {
	return p.URI
}

// -----------------------------------------------------------------------------

func (p *Service) ImageFrom(mime xai.ImageType, src io.Reader) (xai.Image, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return p.ImageFromBytes(mime, data), nil
}

func (p *Service) ImageFromLocal(mime xai.ImageType, fileName string) (xai.Image, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return p.ImageFromBytes(mime, data), nil
}

func (p *Service) ImageFromStgUri(mime xai.ImageType, stgUri string) xai.Image {
	return &image{
		GCSURI:   stgUri,
		MIMEType: string(mime),
	}
}

func (p *Service) ImageFromBytes(mime xai.ImageType, data []byte) xai.Image {
	return &image{
		ImageBytes: data,
		MIMEType:   string(mime),
	}
}

func (p *Service) ImageFromBase64(mime xai.ImageType, data string) (xai.Image, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return &image{
		ImageBytes: b,
		MIMEType:   string(mime),
	}, nil
}

// -----------------------------------------------------------------------------

func (p *Service) VideoFrom(mime xai.VideoType, src io.Reader) (xai.Video, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	return p.VideoFromBytes(mime, data), nil
}

func (p *Service) VideoFromLocal(mime xai.VideoType, fileName string) (xai.Video, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	return p.VideoFromBytes(mime, data), nil
}

func (p *Service) VideoFromStgUri(mime xai.VideoType, stgUri string) xai.Video {
	return &video{
		URI:      stgUri,
		MIMEType: string(mime),
	}
}

func (p *Service) VideoFromBytes(mime xai.VideoType, data []byte) xai.Video {
	return &video{
		VideoBytes: data,
		MIMEType:   string(mime),
	}
}

func (p *Service) VideoFromBase64(mime xai.VideoType, data string) (xai.Video, error) {
	b, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	return &video{
		VideoBytes: b,
		MIMEType:   string(mime),
	}, nil
}

// -----------------------------------------------------------------------------

func (p *Service) GenVideoMask(img xai.Image, maskMode string) xai.GenVideoMask {
	ret := &genai.VideoGenerationMask{
		MaskMode: genai.VideoGenerationMaskMode(maskMode),
	}
	if img != nil {
		ret.Image = imageOf(img)
	}
	return ret
}

// -----------------------------------------------------------------------------

type config[T any] struct {
	conf *T
}

func newConf[T any](conf *T) config[T] {
	return config[T]{conf: conf}
}

func (p config[T]) Schema() xai.InputSchema {
	return newInputSchema(p.conf, nil)
}

func (p config[T]) Params() xai.Params {
	return newParams(p.conf)
}

func (p *Service) ReferenceImage(img xai.Image, id int32, typ xai.ReferenceImageType) (xai.ReferenceImage, xai.Configurable) {
	in := imageOf(img)
	switch typ {
	case xai.RawReferenceImage:
		return genai.NewRawReferenceImage(in, id), nil
	case xai.MaskReferenceImage:
		conf := new(genai.MaskReferenceConfig)
		return genai.NewMaskReferenceImage(in, id, conf), newConf(conf)
	case xai.ControlReferenceImage:
		conf := new(genai.ControlReferenceConfig)
		return genai.NewControlReferenceImage(in, id, conf), newConf(conf)
	case xai.StyleReferenceImage:
		conf := new(genai.StyleReferenceConfig)
		return genai.NewStyleReferenceImage(in, id, conf), newConf(conf)
	case xai.SubjectReferenceImage:
		conf := new(genai.SubjectReferenceConfig)
		return genai.NewSubjectReferenceImage(in, id, conf), newConf(conf)
	case xai.ContentReferenceImage:
		return genai.NewContentReferenceImage(in, id), nil
	default:
		panic("unknown reference image type")
	}
}

// -----------------------------------------------------------------------------

func (p *Service) GenVideoReferenceImages(imgs ...xai.GenVideoReferenceImage) xai.GenVideoReferenceImages {
	items := make([]*genai.VideoGenerationReferenceImage, len(imgs))
	for i, img := range imgs {
		items[i] = &genai.VideoGenerationReferenceImage{
			Image:         imageOf(img.Image),
			ReferenceType: genai.VideoGenerationReferenceType(img.ReferenceType),
		}
	}
	return items
}

// -----------------------------------------------------------------------------

type inputSchema struct {
	t           reflect.Type
	restriction map[string]*xai.Restriction
}

func (p *inputSchema) Restriction(name string) *xai.Restriction {
	return p.restriction[name]
}

func (p *inputSchema) Fields() []xai.Field {
	t := p.t
	n := t.NumField()
	fields := make([]xai.Field, 0, n)
	return getFields(fields, t, n)
}

var rewriteFlds = map[string]string{
	"HTTPOptions":     "",
	"SDKHTTPResponse": "",
	"Labels":          "",
	"OutputGCSURI":    "OutputStgUri",
}

func getFields(fields []xai.Field, t reflect.Type, n int) []xai.Field {
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if f.Anonymous {
			ft := f.Type
			fields = getFields(fields, ft, ft.NumField())
		} else if f.IsExported() {
			name := f.Name
			if newName, ok := rewriteFlds[name]; ok {
				if newName == "" {
					continue
				}
				name = newName
			}
			fields = append(fields, xai.Field{
				Name: name,
				Kind: kindOf(f.Type),
			})
		}
	}
	return fields
}

var allowTypes = map[string]types.Kind{
	"Image":                         types.Image,
	"ProductImage":                  types.Image,
	"ScribbleImage":                 types.Image,
	"ReferenceImage":                types.ReferenceImage,
	"Video":                         types.Video,
	"VideoGenerationReferenceImage": types.GenVideoReferenceImage,
	"VideoGenerationMask":           types.GenVideoMask,
	"GeneratedVideo":                types.OutputVideo,
	"GeneratedImage":                types.OutputImage,
	"GeneratedImageMask":            types.OutputImageMask,
	"SafetyAttributes":              types.SafetyAttributes,
}

func kindOf(t reflect.Type) types.Kind {
	kind := t.Kind()
	if kind == reflect.Pointer {
		t = t.Elem()
		kind = t.Kind()
	}
	switch kind {
	case reflect.Int32, reflect.Int64:
		return types.Int
	case reflect.Float32, reflect.Float64:
		return types.Float
	case reflect.String:
		return types.String
	case reflect.Bool:
		return types.Bool
	case reflect.Slice:
		return kindOf(t.Elem()) | types.List
	case reflect.Struct, reflect.Interface:
		name := t.Name()
		if k, ok := allowTypes[name]; ok {
			return k
		}
		fallthrough
	default:
		log.Panicln("unknown field type:", t)
		return types.Invalid
	}
}

func newInputSchema(params any, restriction map[string]*xai.Restriction) xai.InputSchema {
	return &inputSchema{t: reflect.TypeOf(params).Elem(), restriction: restriction}
}

// -----------------------------------------------------------------------------

type adapter struct {
	util.PointerAsOpt
}

func (adapter) ToUnderlying(val any) any {
	switch v := val.(type) {
	case *image:
		return (*genai.Image)(v)
	case *video:
		return (*genai.Video)(v)
	case *xai.SafetyAttributes:
		return safetyAttributesOf(v)
	}
	return val
}

func (adapter) FromUnderlying(v any, kind reflect.Kind) any {
	if kind == reflect.Pointer {
		switch v := v.(type) {
		case *genai.SafetyAttributes:
			return safetyAttributes(v)
		}
	}
	return v
}

func (adapter) OutputImageFrom(item *genai.GeneratedImage) *xai.OutputImage {
	return &xai.OutputImage{
		Image:             (*image)(item.Image),
		RAIFilteredReason: item.RAIFilteredReason,
		SafetyAttributes:  safetyAttributes(item.SafetyAttributes),
		EnhancedPrompt:    item.EnhancedPrompt,
	}
}

func (adapter) OutputImageMaskFrom(item *genai.GeneratedImageMask) *xai.OutputImageMask {
	return &xai.OutputImageMask{
		Mask:   (*image)(item.Mask),
		Labels: entityLabels(item.Labels),
	}
}

func (adapter) OutputVideoFrom(item *genai.GeneratedVideo) *xai.OutputVideo {
	return &xai.OutputVideo{
		Video: (*video)(item.Video),
	}
}

func newParams(params any) *util.Params[adapter] {
	return util.NewParams[adapter](params)
}

// -----------------------------------------------------------------------------

type callParams struct {
	params util.Params[adapter]
	ctx    context.Context
	opts   *genai.HTTPOptions
}

func (p *callParams) getWaitParams(wp xai.WaitParams) *waitParams {
	if params, ok := wp.(*waitParams); ok {
		return params
	}
	return &waitParams{ctx: p.ctx, opts: p.opts}
}

func (p *callParams) initCallParams(params any) xai.CallParams {
	p.params = *util.NewParams[adapter](params)
	return p
}

func (p *callParams) Ctx(ctx context.Context) xai.CallParams {
	p.ctx = ctx
	return p
}

func (p *callParams) BaseURL(base string) xai.CallParams {
	if p.opts == nil {
		p.opts = &genai.HTTPOptions{}
	}
	p.opts.BaseURL = base
	return p
}

func (p *callParams) Timeout(timeout time.Duration) xai.CallParams {
	if p.opts == nil {
		p.opts = &genai.HTTPOptions{}
	}
	p.opts.Timeout = &timeout
	return p
}

func (p *callParams) Set(name string, val any) xai.CallParams {
	p.params.Set(name, val)
	return p
}

// -----------------------------------------------------------------------------

type waitParams struct {
	ctx      context.Context
	opts     *genai.HTTPOptions
	progress func(xai.OperationResponse)
}

func newWaitParams(cp *callParams) *waitParams {
	return &waitParams{
		ctx:  cp.ctx,
		opts: cp.opts,
	}
}

func (p *waitParams) Ctx(ctx context.Context) xai.WaitParams {
	p.ctx = ctx
	return p
}

func (p *waitParams) Progress(progress func(xai.OperationResponse)) xai.WaitParams {
	p.progress = progress
	return p
}

func (p *waitParams) BaseURL(base string) xai.WaitParams {
	if p.opts == nil {
		p.opts = &genai.HTTPOptions{}
	}
	p.opts.BaseURL = base
	return p
}

func (p *waitParams) Timeout(timeout time.Duration) xai.WaitParams {
	if p.opts == nil {
		p.opts = &genai.HTTPOptions{}
	}
	p.opts.Timeout = &timeout
	return p
}

// -----------------------------------------------------------------------------

func safetyAttributes(v *genai.SafetyAttributes) *xai.SafetyAttributes {
	if v == nil {
		return nil
	}
	return &xai.SafetyAttributes{
		Categories: v.Categories,
		Scores:     v.Scores,
	}
}

func safetyAttributesOf(v *xai.SafetyAttributes) *genai.SafetyAttributes {
	if v == nil {
		return nil
	}
	return &genai.SafetyAttributes{
		Categories: v.Categories,
		Scores:     v.Scores,
	}
}

type entityLabels []*genai.EntityLabel

func (v entityLabels) Len() int {
	return len(v)
}

func (v entityLabels) At(i int) xai.EntityLabel {
	item := v[i]
	return xai.EntityLabel{
		Label: item.Label,
		Score: item.Score,
	}
}

// -----------------------------------------------------------------------------
