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
	"unsafe"

	xai "github.com/goplus/xai/spec"
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

// implement xai.Service
func (p *ServiceBase) Options() xai.OptionBuilder {
	return new(HTTPOptions)
}

// -----------------------------------------------------------------------------

// ResponseCreator is a function type that creates an xai.OperationResponse from
// a given HTTP response body.
type ResponseCreator func(c *Client, body map[string]any) (xai.OperationResponse, error)

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

type Operation[T serviceAdapter] struct {
	body    map[string]any
	action  xai.Action
	model   xai.Model
	c       *Client
	adapter T
}

func (p *Operation[T]) InputSchema() xai.InputSchema {
	return p.adapter.InputSchema(p.action)
}

func (p *Operation[T]) Params() xai.Params {
	return p
}

func (p *Operation[T]) Set(name string, val any) xai.Params {
	p.adapter.SetParam(p.body, name, val)
	return p
}

func (p *Operation[T]) Call(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
	a := p.adapter.BuildAction(p.action, p.body, p.model)
	req, err := p.c.NewRequest(http.MethodPost, a.Path)
	if err != nil {
		return
	}
	err = req.Json(p.body)
	if err != nil {
		return
	}
	return call(ctx, req, a.NewResponse, opts)
}

func call(ctx context.Context, req *Request, newResponse ResponseCreator, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
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
	return newResponse(req.c, body)
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
	action  xai.Action
	adapter T
}

func NewOperationResponse[T responseAdapter](c *Client, action xai.Action, body map[string]any) *OperationResponse[T] {
	return &OperationResponse[T]{c: c, body: body, action: action}
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

func (p *OperationResponse[T]) Retry(ctx context.Context, svc xai.Service, opts xai.OptionBuilder) (resp xai.OperationResponse, err error) {
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
	return call(ctx, req, qoi.NewResponse, opts)
}

func (p *OperationResponse[T]) Results() xai.Results {
	return p.adapter.Results(p.action, p.body)
}

// -----------------------------------------------------------------------------
