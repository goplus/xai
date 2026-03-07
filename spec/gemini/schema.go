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
	"encoding/base64"
	"io"
	"log"
	"os"
	"reflect"

	xai "github.com/goplus/xai/spec"
	"github.com/goplus/xai/types"
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
	return newInputSchema(p.conf)
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

type opInputSchema struct {
	t reflect.Type
}

func (p *opInputSchema) Fields() []xai.Field {
	t := p.t
	n := t.NumField()
	fields := make([]xai.Field, 0, n)
	return getFields(fields, t, n)
}

var rewriteFlds = map[string]string{
	"HTTPOptions":     "",
	"SDKHTTPResponse": "",
	"Labels":          "",
	"Prompt":          "",
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

func newInputSchema(params any) xai.InputSchema {
	return &opInputSchema{t: reflect.TypeOf(params).Elem()}
}

// -----------------------------------------------------------------------------

type opParams struct {
	v reflect.Value
}

func newParams(params any) *opParams {
	return &opParams{v: reflect.ValueOf(params).Elem()}
}

func (p *opParams) Set(name string, val any) xai.Params {
	fld := p.v.FieldByName(name)
	if fld.CanSet() {
		v := reflect.ValueOf(val)
		vkind := v.Kind()
		if vkind >= reflect.Bool && vkind <= reflect.Float64 {
			if fld.Kind() == reflect.Pointer {
				pv := reflect.New(fld.Type().Elem())
				setBasic(pv.Elem(), v, vkind)
				fld.Set(pv)
			} else {
				setBasic(fld, v, vkind)
			}
		} else {
			fld.Set(v)
		}
	} else {
		log.Println("cannot set field:", name)
	}
	return p
}

func setBasic(fld, v reflect.Value, vkind reflect.Kind) {
	if vkind >= reflect.Int && vkind <= reflect.Int64 {
		if kind := fld.Kind(); kind >= reflect.Int && kind <= reflect.Int64 {
			fld.SetInt(v.Int())
		} else {
			fld.SetFloat(float64(v.Int()))
		}
	} else if vkind >= reflect.Float32 && vkind <= reflect.Float64 {
		fld.SetFloat(v.Float())
	} else {
		fld.Set(v)
	}
}

// -----------------------------------------------------------------------------

type opResults struct {
	v reflect.Value
}

func results(resp any) opResults {
	return opResults{v: reflect.ValueOf(resp).Elem()}
}

func (p *opResults) XGo_Attr(name string) any {
	fld := p.v.FieldByName(name)
	kind := fld.Kind()
	if kind == reflect.Invalid {
		return nil
	}
	if kind == reflect.Pointer {
		if fld.IsNil() {
			return nil
		}
		fld := fld.Elem()
		kind = fld.Kind()
	}
	if kind >= reflect.Int && kind <= reflect.Int64 {
		return fld.Int()
	} else if kind >= reflect.Float32 && kind <= reflect.Float64 {
		return fld.Float()
	}
	return fld.Interface()
}

// -----------------------------------------------------------------------------

type outputVideos struct {
	opResults
	items []*genai.GeneratedVideo
}

func (p *outputVideos) Len() int {
	return len(p.items)
}

func (p *outputVideos) At(i int) xai.Generated {
	item := p.items[i]
	return &xai.OutputVideo{
		Video: (*video)(item.Video),
	}
}

// -----------------------------------------------------------------------------

type outputImages struct {
	opResults
	items []*genai.GeneratedImage
}

func (p *outputImages) Len() int {
	return len(p.items)
}

func (p *outputImages) At(i int) xai.Generated {
	item := p.items[i]
	return &xai.OutputImage{
		Image:             (*image)(item.Image),
		RAIFilteredReason: item.RAIFilteredReason,
		SafetyAttributes:  safetyAttributes(item.SafetyAttributes),
		EnhancedPrompt:    item.EnhancedPrompt,
	}
}

func safetyAttributes(v *genai.SafetyAttributes) *xai.SafetyAttributes {
	if v == nil {
		return nil
	}
	return &xai.SafetyAttributes{
		Categories: v.Categories,
		Scores:     v.Scores,
	}
}

// -----------------------------------------------------------------------------

type outputImageMasks struct {
	opResults
	items []*genai.GeneratedImageMask
}

func (p *outputImageMasks) Len() int {
	return len(p.items)
}

func (p *outputImageMasks) At(i int) xai.Generated {
	item := p.items[i]
	return &xai.OutputImageMask{
		Mask:   (*image)(item.Mask),
		Labels: entityLabels(item.Labels),
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
