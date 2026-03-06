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

func (p *image) MIME() xai.ImageType {
	return xai.ImageType(p.MIMEType)
}

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
	v       reflect.Value
	genName string
}

func newResults(results any, genName string) *opResults {
	return &opResults{v: reflect.ValueOf(results).Elem(), genName: genName}
}

func (p *opResults) Prop(name string) any {
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

func (p *opResults) Len() int {
	if p.genName == "" {
		return 0
	}
	return p.v.FieldByName(p.genName).Len()
}

func (p *opResults) At(i int) any {
	return p.v.FieldByName(p.genName).Index(i).Interface()
}

// -----------------------------------------------------------------------------
