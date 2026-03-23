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

package geno

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
	"unsafe"

	"github.com/goplus/xai"
)

// -----------------------------------------------------------------------------

// NameToCStyle converts a CamelCase name to snake_case, e.g. "ImageSize" to "image_size".
func NameToCStyle(name string) string {
	nx := len(name)
	for i, n := 1, nx; i < n; i++ {
		if c := name[i]; c >= 'A' && c <= 'Z' {
			nx++
		}
	}
	b := make([]byte, 0, nx)
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'A' && c <= 'Z' {
			if i > 0 {
				b = append(b, '_')
			}
			b = append(b, c+('a'-'A'))
		} else {
			b = append(b, c)
		}
	}
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// -----------------------------------------------------------------------------

type ServiceBase struct {
	c Client
}

// HTTPClient returns the underlying HTTP client used by the service.
func (p *ServiceBase) HTTPClient() *Client {
	return &p.c
}

// -----------------------------------------------------------------------------

// ResponseCreator is a function type that creates an xai.OperationResponse from
// a given HTTP response body.
type ResponseCreator func(c *Client, body map[string]any, cp *CallParamsBase) (xai.OperationResponse, error)

// ActionInfo contains information about an action, including the path to call for
// the action, the name of the parameter for the model, and a function to create a
// response from the HTTP response body.
type ActionInfo struct {
	Path        string // path to call for the action, e.g. "/v1/images/generations"
	NewResponse ResponseCreator
}

type serviceAdapter interface {
	// Actions returns the list of actions supported by the given model.
	Actions(model xai.Model) []xai.Action

	// InputSchema returns the input schema for the given action.
	InputSchema(action xai.Action) xai.InputSchema

	// SetParam sets the parameter with the given name and value in the request body.
	// name should convert from XAI style to the API native style, e.g. "ImageSize"
	// to "image_size".
	SetParam(body map[string]any, name string, val any)

	// BuildAction builds the request body and returns ActionInfo for the given action.
	// model should be added to body if needed.
	BuildAction(action xai.Action, body map[string]any, model xai.Model) ActionInfo
}

type Service[T serviceAdapter] struct {
	ServiceBase
}

// NewService creates a new Service with the provided HTTP client.
func NewService[T serviceAdapter](client *http.Client) *Service[T] {
	return &Service[T]{
		ServiceBase: ServiceBase{c: *NewClient(client)},
	}
}

// implement xai.Service
func (p *Service[T]) Actions(model xai.Model) []xai.Action {
	var adapter T
	return adapter.Actions(model)
}

// implement xai.Service
func (p *Service[T]) Operation(model xai.Model, action xai.Action) (xai.Operation, error) {
	return &Operation[T]{
		c:      &p.c,
		body:   make(map[string]any, 16),
		model:  model,
		action: action,
	}, nil
}

// -----------------------------------------------------------------------------

type CallParamsBase struct {
	opts *HTTPOptions
	ctx  context.Context
}

func (p *CallParamsBase) getWaitParams(wp xai.WaitParams) *waitParams {
	if params, ok := wp.(*waitParams); ok {
		return params
	}
	return &waitParams{ctx: p.ctx, opts: p.opts}
}

func (p *CallParamsBase) Ctx(ctx context.Context) xai.CallParams {
	p.ctx = ctx
	return p
}

func (p *CallParamsBase) BaseURL(base string) xai.CallParams {
	if p.opts == nil {
		p.opts = &HTTPOptions{}
	}
	p.opts.BaseURL(base)
	return p
}

func (p *CallParamsBase) Timeout(timeout time.Duration) xai.CallParams {
	if p.opts == nil {
		p.opts = &HTTPOptions{}
	}
	p.opts.Timeout(timeout)
	return p
}

func (p *CallParamsBase) Set(name string, val any) xai.CallParams {
	panic("unreachable")
}

// -----------------------------------------------------------------------------

type Operation[T serviceAdapter] struct {
	body    map[string]any
	action  xai.Action
	model   xai.Model
	c       *Client
	adapter T
	CallParamsBase
}

func (p *Operation[T]) InputSchema() xai.InputSchema {
	return p.adapter.InputSchema(p.action)
}

func (p *Operation[T]) Set(name string, val any) xai.CallParams {
	p.adapter.SetParam(p.body, name, val)
	return p
}

func (p *Operation[T]) CallParams() xai.CallParams {
	return p
}

func (p *Operation[T]) Call(xai.CallParams) (resp xai.OperationResponse, err error) {
	a := p.adapter.BuildAction(p.action, p.body, p.model)
	req, err := p.c.NewRequest(http.MethodPost, a.Path)
	if err != nil {
		return
	}
	err = req.Json(p.body)
	if err != nil {
		return
	}
	return call(p.ctx, req, a.NewResponse, p.opts, &p.CallParamsBase)
}

func call(ctx context.Context, req *Request, newResp ResponseCreator, opts *HTTPOptions, cp *CallParamsBase) (resp xai.OperationResponse, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	ret, err := req.Do(ctx, opts)
	if err != nil {
		return
	}
	defer ret.Body.Close()

	var body map[string]any
	dec := json.NewDecoder(ret.Body)
	err = dec.Decode(&body)
	if err != nil {
		return
	}
	return newResp(req.c, body, cp)
}

// -----------------------------------------------------------------------------

type resultAdapter interface {
	// GetAttr retrieves the attribute with the given name from the result body.
	// name should convert from XAI style to the API native style, e.g. "ImageURL"
	// to "image_url".
	GetAttr(result map[string]any, name string) any
}

type results[T resultAdapter] struct {
	result map[string]any
	items  []any
}

func newResults[T resultAdapter](result map[string]any, itemsName string) results[T] {
	items, _ := result[itemsName].([]any)
	return results[T]{
		result: result,
		items:  items,
	}
}

func (p *results[T]) XGo_Attr(name string) any {
	var adapter T
	return adapter.GetAttr(p.result, name)
}

func (p *results[T]) Len() int {
	return len(p.items)
}

// -----------------------------------------------------------------------------

type imageResultAdapter interface {
	resultAdapter
	OutputImage(item any) *xai.OutputImage
}

type ImageResults[T imageResultAdapter] struct {
	results[T]
}

func NewImageResults[T imageResultAdapter](result map[string]any, itemsName string) xai.Results {
	return &ImageResults[T]{
		results: newResults[T](result, itemsName),
	}
}

func (p *ImageResults[T]) At(i int) xai.Generated {
	var adapter T
	return adapter.OutputImage(p.items[i])
}

// -----------------------------------------------------------------------------

type imageMaskResultAdapter interface {
	resultAdapter
	OutputImageMask(item any) *xai.OutputImageMask
}

type ImageMaskResults[T imageMaskResultAdapter] struct {
	results[T]
}

func NewImageMaskResults[T imageMaskResultAdapter](result map[string]any, itemsName string) xai.Results {
	return &ImageMaskResults[T]{
		results: newResults[T](result, itemsName),
	}
}

func (p *ImageMaskResults[T]) At(i int) xai.Generated {
	var adapter T
	return adapter.OutputImageMask(p.items[i])
}

// -----------------------------------------------------------------------------

type videoResultAdapter interface {
	resultAdapter
	OutputVideo(item any) *xai.OutputVideo
}

type VideoResults[T videoResultAdapter] struct {
	results[T]
}

func NewVideoResults[T videoResultAdapter](result map[string]any, itemsName string) xai.Results {
	return &VideoResults[T]{
		results: newResults[T](result, itemsName),
	}
}

func (p *VideoResults[T]) At(i int) xai.Generated {
	var adapter T
	return adapter.OutputVideo(p.items[i])
}

// -----------------------------------------------------------------------------

type QueryInfo struct {
	Path        string
	NewResponse ResponseCreator
}

type responseAdapter interface {
	Done(action xai.Action, body map[string]any) bool
	Sleep(action xai.Action, body map[string]any)
	Results(action xai.Action, body map[string]any) xai.Results
	BuildQuery(action xai.Action, body map[string]any) (QueryInfo, error)
}

type OperationResponse[T responseAdapter] struct {
	body    map[string]any
	c       *Client
	cp      *CallParamsBase
	action  xai.Action
	adapter T
}

func NewOperationResponse[T responseAdapter](c *Client, action xai.Action, body map[string]any, cp *CallParamsBase) *OperationResponse[T] {
	return &OperationResponse[T]{c: c, body: body, action: action, cp: cp}
}

func (p *OperationResponse[T]) Done() bool {
	return p.adapter.Done(p.action, p.body)
}

func (p *OperationResponse[T]) Sleep() {
	p.adapter.Sleep(p.action, p.body)
}

var (
	ErrMissingOperationID = errors.New("missing operation ID in response body")
)

func (p *OperationResponse[T]) Retry(wp xai.WaitParams) (resp *OperationResponse[T], err error) {
	qoi, err := p.adapter.BuildQuery(p.action, p.body)
	if err != nil {
		return
	}
	req, err := p.c.NewRequest(http.MethodGet, qoi.Path)
	if err != nil {
		return
	}
	// query operation has no body,
	// so we can directly call it without setting body
	params := p.cp.getWaitParams(wp)
	ret, err := call(params.ctx, req, qoi.NewResponse, params.opts, p.cp)
	if err != nil {
		return
	}
	return ret.(*OperationResponse[T]), nil
}

func (p *OperationResponse[T]) Results() xai.Results {
	return p.adapter.Results(p.action, p.body)
}

func (p *OperationResponse[T]) WaitParams() xai.WaitParams {
	return newWaitParams(p.cp)
}

func (p *OperationResponse[T]) Wait(wp xai.WaitParams) (ret xai.Results, err error) {
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

// -----------------------------------------------------------------------------

type waitParams struct {
	ctx      context.Context
	opts     *HTTPOptions
	progress func(xai.OperationResponse)
}

func newWaitParams(cp *CallParamsBase) *waitParams {
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
		p.opts = &HTTPOptions{}
	}
	p.opts.BaseURL(base)
	return p
}

func (p *waitParams) Timeout(timeout time.Duration) xai.WaitParams {
	if p.opts == nil {
		p.opts = &HTTPOptions{}
	}
	p.opts.Timeout(timeout)
	return p
}

// -----------------------------------------------------------------------------
