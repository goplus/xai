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

package util

import (
	"context"
	"log"
	"reflect"

	xai "github.com/goplus/xai/spec"
)

// -----------------------------------------------------------------------------

type paramsAdapter interface {
	ToUnderlying(val any) any
	SetBasicOpt(dst, val reflect.Value, vkind reflect.Kind) bool
}

type Params[T paramsAdapter] struct {
	v reflect.Value
}

func NewParams[T paramsAdapter](params any) *Params[T] {
	return &Params[T]{v: reflect.ValueOf(params).Elem()}
}

func (p *Params[T]) Set(name string, val any) xai.Params {
	fld := p.v.FieldByName(name)
	if fld.CanSet() {
		if val == nil {
			fld.SetZero()
			return p
		}

		// convert a spec-layer object to underlying-layer
		var adapter T
		val = adapter.ToUnderlying(val)

		v := reflect.ValueOf(val)
		vkind := v.Kind()
		if vkind >= reflect.Bool && vkind <= reflect.Float64 || vkind == reflect.String {
			if !adapter.SetBasicOpt(fld, v, vkind) {
				SetBasic(fld, v, vkind)
			}
		} else {
			fld.Set(v)
		}
	} else {
		log.Println("cannot set field:", name)
	}
	return p
}

func SetBasic(fld, v reflect.Value, vkind reflect.Kind) {
	if vkind >= reflect.Int && vkind <= reflect.Int64 {
		if kind := fld.Kind(); kind >= reflect.Int && kind <= reflect.Int64 {
			fld.SetInt(v.Int())
		} else {
			fld.SetFloat(float64(v.Int()))
		}
		return
	}
	switch vkind {
	case reflect.String:
		fld.SetString(v.String())
	case reflect.Bool:
		fld.SetBool(v.Bool())
	case reflect.Float32, reflect.Float64:
		fld.SetFloat(v.Float())
	default:
		fld.Set(v)
	}
}

// -----------------------------------------------------------------------------

type PointerAsOpt struct{}

func (PointerAsOpt) SetBasicOpt(fld, v reflect.Value, vkind reflect.Kind) (ok bool) {
	ok = fld.Kind() == reflect.Pointer
	if ok {
		pv := reflect.New(fld.Type().Elem())
		SetBasic(pv.Elem(), v, vkind)
		fld.Set(pv)
	}
	return
}

// -----------------------------------------------------------------------------

type resultsAdapter interface {
	FromUnderlying(val any, kind reflect.Kind) any
}

type Results[T resultsAdapter] struct {
	v reflect.Value
}

func NewResults[T resultsAdapter](ret any) *Results[T] {
	return &Results[T]{v: reflect.ValueOf(ret).Elem()}
}

func (p *Results[T]) XGo_Attr(name string) any {
	fld := p.v.FieldByName(name)
	kind := fld.Kind()
	if kind >= reflect.Int && kind <= reflect.Int64 {
		return fld.Int()
	}
	switch kind {
	case reflect.String:
		return fld.String()
	case reflect.Bool:
		return fld.Bool()
	case reflect.Float32, reflect.Float64:
		return fld.Float()
	case reflect.Invalid:
		return nil
	}
	// convert a underlying-layer object to spec-layer
	var adapter T
	return adapter.FromUnderlying(fld.Interface(), kind)
}

// -----------------------------------------------------------------------------

type imageResultsAdpt[U any] interface {
	resultsAdapter
	OutputImageFrom(image U) *xai.OutputImage
}

type ImageResults[U any, T imageResultsAdpt[U]] struct {
	Results[T]
	items []U
}

func NewImageResults[U any, T imageResultsAdpt[U]](ret any, items []U) *ImageResults[U, T] {
	return &ImageResults[U, T]{
		Results: Results[T]{v: reflect.ValueOf(ret).Elem()},
		items:   items,
	}
}

func (p *ImageResults[U, T]) Len() int {
	return len(p.items)
}

func (p *ImageResults[U, T]) At(i int) xai.Generated {
	var adapter T
	return adapter.OutputImageFrom(p.items[i])
}

// -----------------------------------------------------------------------------

type imageMaskResultsAdpt[U any] interface {
	resultsAdapter
	OutputImageMaskFrom(image U) *xai.OutputImageMask
}

func NewImageMaskResults[U any, T imageMaskResultsAdpt[U]](ret any, items []U) *ImageMaskResults[U, T] {
	return &ImageMaskResults[U, T]{
		Results: Results[T]{v: reflect.ValueOf(ret).Elem()},
		items:   items,
	}
}

type ImageMaskResults[U any, T imageMaskResultsAdpt[U]] struct {
	Results[T]
	items []U
}

func (p *ImageMaskResults[U, T]) Len() int {
	return len(p.items)
}

func (p *ImageMaskResults[U, T]) At(i int) xai.Generated {
	var adapter T
	return adapter.OutputImageMaskFrom(p.items[i])
}

// -----------------------------------------------------------------------------

type videoResultsAdpt[U any] interface {
	resultsAdapter
	OutputVideoFrom(video U) *xai.OutputVideo
}

type VideoResults[U any, T videoResultsAdpt[U]] struct {
	Results[T]
	items []U
}

func NewVideoResults[U any, T videoResultsAdpt[U]](ret any, items []U) *VideoResults[U, T] {
	return &VideoResults[U, T]{
		Results: Results[T]{v: reflect.ValueOf(ret).Elem()},
		items:   items,
	}
}

func (p *VideoResults[U, T]) Len() int {
	return len(p.items)
}

func (p *VideoResults[U, T]) At(i int) xai.Generated {
	var adapter T
	return adapter.OutputVideoFrom(p.items[i])
}

// -----------------------------------------------------------------------------

type SimpleResp[T xai.Results] struct {
	ret T
}

func NewSimpleResp[T xai.Results](v T) SimpleResp[T] {
	return SimpleResp[T]{v}
}

func NewImageResultsResp[U any, T imageResultsAdpt[U]](ret any, items []U) SimpleResp[*ImageResults[U, T]] {
	return NewSimpleResp(NewImageResults[U, T](ret, items))
}

func NewImageMaskResultsResp[U any, T imageMaskResultsAdpt[U]](ret any, items []U) SimpleResp[*ImageMaskResults[U, T]] {
	return NewSimpleResp(NewImageMaskResults[U, T](ret, items))
}

func NewVideoResultsResp[U any, T videoResultsAdpt[U]](ret any, items []U) SimpleResp[*VideoResults[U, T]] {
	return NewSimpleResp(NewVideoResults[U, T](ret, items))
}

func (p SimpleResp[T]) Done() bool {
	return true
}

func (p SimpleResp[T]) Sleep() {
	panic("unreachable")
}

func (p SimpleResp[T]) Retry(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (xai.OperationResponse, error) {
	panic("unreachable")
}

func (p SimpleResp[T]) Results() xai.Results {
	return p.ret
}

// -----------------------------------------------------------------------------
